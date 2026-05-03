package main

import (
	"testing"

	"github.com/clbanning/mxj/v2"
)

func TestTransform(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<REQUEST>
	<DEVICEID>test-device</DEVICEID>
	<CONTENT>
		<HARDWARE>
			<NAME>my-host</NAME>
			<IPADDR>1.2.3.4</IPADDR>
		</HARDWARE>
	</CONTENT>
</REQUEST>`

	mv, err := mxj.NewMapXml([]byte(xmlData))
	if err != nil {
		t.Fatalf("failed to parse xml: %v", err)
	}

	fields := map[string]string{
		"hostname":  "REQUEST.CONTENT.HARDWARE.NAME",
		"ip":        "REQUEST.CONTENT.HARDWARE.IPADDR",
		"device_id": "REQUEST.DEVICEID",
		"missing":   "REQUEST.NONEXISTENT",
	}

	result := transform(mv, fields)

	if result["hostname"] != "my-host" {
		t.Errorf("expected hostname my-host, got %v", result["hostname"])
	}
	if result["ip"] != "1.2.3.4" {
		t.Errorf("expected ip 1.2.3.4, got %v", result["ip"])
	}
	if result["device_id"] != "test-device" {
		t.Errorf("expected device_id test-device, got %v", result["device_id"])
	}
	if result["missing"] != nil {
		t.Errorf("expected missing to be nil, got %v", result["missing"])
	}
}
