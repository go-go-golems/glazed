package logging

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// LogstashWriter implements io.Writer interface
// for sending logs to Logstash via TCP or UDP
type LogstashWriter struct {
	Host              string
	Port              int
	Protocol          string // "tcp" or "udp"
	AppName           string
	Environment       string
	conn              net.Conn
	reconnectInterval time.Duration
	lastReconnect     time.Time
}

// NewLogstashWriter creates a new LogstashWriter for sending logs to Logstash
func NewLogstashWriter(host string, port int, protocol, appName, environment string) *LogstashWriter {
	return &LogstashWriter{
		Host:              host,
		Port:              port,
		Protocol:          protocol,
		AppName:           appName,
		Environment:       environment,
		reconnectInterval: 5 * time.Second,
		lastReconnect:     time.Time{},
	}
}

// Write implements the io.Writer interface
func (w *LogstashWriter) Write(p []byte) (int, error) {
	// Parse the JSON log entry
	var event map[string]interface{}
	if err := json.Unmarshal(p, &event); err != nil {
		// If not valid JSON, create a simple message
		event = map[string]interface{}{
			"message": string(p),
			"level":   "info",
		}
	}

	// Add additional fields
	event["@timestamp"] = time.Now().UTC().Format(time.RFC3339)
	event["app"] = w.AppName
	event["environment"] = w.Environment
	event["type"] = "go"

	// Convert back to JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		return 0, err
	}

	// Ensure we have a connection
	if err := w.ensureConnection(); err != nil {
		return 0, err
	}

	// Send the log to Logstash
	n, err := w.conn.Write(append(jsonData, '\n'))
	if err != nil {
		// If write fails, clear the connection so we'll reconnect next time
		w.conn = nil
		return n, err
	}
	if n != len(jsonData)+1 {
		return n, fmt.Errorf("failed to write all data to Logstash")
	}

	// we have to return the original length of the data
	return len(p), nil
}

// ensureConnection makes sure we have an active connection to Logstash
func (w *LogstashWriter) ensureConnection() error {
	if w.conn != nil {
		return nil
	}

	// Rate limit reconnection attempts
	now := time.Now()
	if !w.lastReconnect.IsZero() && now.Sub(w.lastReconnect) < w.reconnectInterval {
		return fmt.Errorf("reconnection rate limited")
	}
	w.lastReconnect = now

	// Connect to Logstash
	addr := fmt.Sprintf("%s:%d", w.Host, w.Port)
	conn, err := net.Dial(w.Protocol, addr)
	if err != nil {
		return fmt.Errorf("failed to connect to Logstash at %s: %v", addr, err)
	}

	w.conn = conn
	return nil
}

// Close closes the connection to Logstash
func (w *LogstashWriter) Close() error {
	if w.conn != nil {
		err := w.conn.Close()
		w.conn = nil
		return err
	}
	return nil
}

// SetupLogstashLogger configures a logger that sends logs to Logstash
func SetupLogstashLogger(host string, port int, protocol, appName, environment string) *LogstashWriter {
	return NewLogstashWriter(host, port, protocol, appName, environment)
}
