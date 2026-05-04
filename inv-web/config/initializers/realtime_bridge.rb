# This bridge listens for PostgreSQL NOTIFY events and broadcasts them via Turbo Streams.
# This allows the UI to update even when data is inserted externally (e.g., via the Go consumer).

Rails.application.config.after_initialize do
  # Only start the listener in the web server process, not in migrations or rake tasks
  if defined?(Rails::Server) || (defined?(Puma) && !File.basename($0).include?('rake'))
    Thread.new do
      Rails.logger.info "Starting PostgreSQL Real-time Bridge..."
      
      # Use a dedicated connection for the listener
      ActiveRecord::Base.connection_pool.with_connection do |connection|
        conn = connection.raw_connection
        conn.exec("LISTEN inventory_updates")
        
        begin
          loop do
            conn.wait_for_notify do |channel, pid, payload|
              Rails.logger.debug "Received PG NOTIFY: #{payload}"
              data = JSON.parse(payload)
              
              case data["action"]
              when "insert"
                device = Device.find_by(id: data["id"])
                if device
                  # Prepend to the list
                  Turbo::StreamsChannel.broadcast_prepend_to(
                    "devices",
                    target: "devices_list",
                    partial: "devices/device",
                    locals: { device: device }
                  )
                end
              when "update"
                device = Device.find_by(id: data["id"])
                if device
                  # Replace in the list
                  Turbo::StreamsChannel.broadcast_replace_to(
                    "devices",
                    target: ActionView::RecordIdentifier.dom_id(device),
                    partial: "devices/device",
                    locals: { device: device }
                  )
                end
              when "delete"
                # Remove from the list (using the dom_id logic)
                Turbo::StreamsChannel.broadcast_remove_to(
                  "devices",
                  target: "device_#{data["id"]}"
                )
              end
            end
          end
        rescue => e
          Rails.logger.error "PostgreSQL Real-time Bridge Error: #{e.message}"
          sleep 5
          retry
        ensure
          conn.exec("UNLISTEN inventory_updates")
        end
      end
    end
  end
end
