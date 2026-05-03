package main

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/clbanning/mxj/v2"
	"github.com/segmentio/kafka-go"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddr string            `yaml:"listen_addr"`
	Debug      bool              `yaml:"debug"`
	Kafka      KafkaConfig       `yaml:"kafka"`
	Fields     map[string]string `yaml:"fields"`
}

type KafkaConfig struct {
	Brokers      []string `yaml:"brokers"`
	Topic        string   `yaml:"topic"`
	Compression  string   `yaml:"compression"`
	RequiredAcks int      `yaml:"required_acks"`
}

var (
	config      Config
	configPath  string
	kafkaWriter *kafka.Writer
	ocsSuccess  = `<?xml version="1.0" encoding="UTF-8"?><REPLY><STATUS>SUCCESS</STATUS></REPLY>`
	ocsProlog   = `<?xml version="1.0" encoding="UTF-8"?><REPLY><RESPONSE>SEND</RESPONSE><PROLOG_FREQ>24</PROLOG_FREQ></REPLY>`
)

func main() {
	flag.StringVar(&configPath, "config", "config.yaml", "Path to configuration file")
	flag.Parse()

	loadConfig(configPath)
	initKafka()
	defer kafkaWriter.Close()

	http.HandleFunc("/", ocsHandler)
	http.HandleFunc("/ocsinventory", ocsHandler)

	fmt.Printf("Starting rb-ocs-collector on %s\n", config.ListenAddr)
	if err := http.ListenAndServe(config.ListenAddr, nil); err != nil {
		log.Fatal(err)
	}
}

func loadConfig(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Warning: configuration file not found at %s, using defaults", path)
		config = Config{
			ListenAddr: "0.0.0.0:8088",
			Debug:      true,
		}
		return
	}

	// Expand environment variables (e.g., ${KAFKA_TOPIC})
	expandedContent := os.ExpandEnv(string(content))

	if err := yaml.Unmarshal([]byte(expandedContent), &config); err != nil {
		log.Fatalf("Error decoding config: %v", err)
	}
}

func initKafka() {
	var codec kafka.Compression
	switch config.Kafka.Compression {
	case "gzip":
		codec = kafka.Gzip
	default:
	}

	kafkaWriter = &kafka.Writer{
		Addr:                   kafka.TCP(config.Kafka.Brokers...),
		Topic:                  config.Kafka.Topic,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequiredAcks(config.Kafka.RequiredAcks),
		Compression:            codec,
		Async:                  false,
		WriteTimeout:           10 * time.Second,
		ReadTimeout:            10 * time.Second,
		Logger:                 kafka.LoggerFunc(log.Printf),
		ErrorLogger:            kafka.LoggerFunc(log.Printf),
	}
}

func ocsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	rawPayload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var data []byte
	isCompressed := false
	b := bytes.NewReader(rawPayload)
	zr, err := zlib.NewReader(b)
	if err == nil {
		data, _ = io.ReadAll(zr)
		zr.Close()
		isCompressed = true
	} else {
		data = rawPayload
	}

	decodedStr := string(data)
	if config.Debug {
		limit := 2048
		if len(decodedStr) < limit {
			limit = len(decodedStr)
		}
		log.Printf("Received Request (Compressed: %v): %s", isCompressed, decodedStr[:limit])
	}

	// Prolog
	if strings.Contains(decodedStr, "<QUERY>PROLOG</QUERY>") {
		log.Println("Prolog detected -> Sending SEND directive")

		// Extract DEVICEID
		mv, err := mxj.NewMapXml(data)
		deviceID := ""
		if err == nil {
			d, _ := mv.ValueForPath("REQUEST.DEVICEID")
			if ds, ok := d.(string); ok {
				deviceID = ds
			}
		}

		response := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<REPLY>
<QUERY>PROLOG</QUERY>
<RESPONSE>SEND</RESPONSE>
<PROLOG_FREQ>24</PROLOG_FREQ>
<DEVICEID>%s</DEVICEID>
</REPLY>`, deviceID)

		if config.Debug {
			log.Printf("Sending Response: %s", response)
		}

		sendResponse(w, response, isCompressed)
		return
	}

	// Inventory
	if strings.Contains(decodedStr, "<QUERY>INVENTORY</QUERY>") {
		log.Println("Inventory detected -> Extracting and sending to Kafka")

		// Use mxj for dynamic parsing
		mv, err := mxj.NewMapXml(data)
		if err != nil {
			log.Printf("Error parsing XML with mxj: %v", err)
		} else {
			sendToKafka(mv)
		}

		sendResponse(w, ocsSuccess, isCompressed)
		return
	}

	log.Println("Request did not match OCS patterns.")
	http.Error(w, "Invalid Request", http.StatusBadRequest)
}

func sendResponse(w http.ResponseWriter, content string, compress bool) {
	if compress {
		var b bytes.Buffer
		zw := zlib.NewWriter(&b)
		zw.Write([]byte(content))
		zw.Close()
		w.Header().Set("Content-Type", "application/x-compressed")
		w.Write(b.Bytes())
	} else {
		w.Header().Set("Content-Type", "text/xml; charset=UTF-8")
		fmt.Fprint(w, content)
	}
}

func sendToKafka(mv mxj.Map) {
	outputMap := transform(mv, config.Fields)

	jsonBytes, err := json.Marshal(outputMap)
	if err != nil {
		log.Printf("Error marshaling to JSON: %v", err)
		return
	}

	err = kafkaWriter.WriteMessages(context.Background(),
		kafka.Message{
			Value: jsonBytes,
		},
	)
	if err != nil {
		log.Printf("Error writing to Kafka: %v", err)
	} else if config.Debug {
		log.Printf("Sent to Kafka: %s", string(jsonBytes))
	}
}

func transform(mv mxj.Map, fields map[string]string) map[string]interface{} {
	outputMap := make(map[string]interface{})

	for k, path := range fields {
		val, err := mv.ValueForPath(path)
		if err != nil {
			if config.Debug {
				log.Printf("Warning: could not find value for path '%s': %v", path, err)
			}
			outputMap[k] = nil
			continue
		}
		outputMap[k] = val
	}
	return outputMap
}
