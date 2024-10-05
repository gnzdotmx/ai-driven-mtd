package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ClientConfig holds the configuration for the client
type ClientConfig struct {
	Servers            []string `json:"servers"`
	RequestIntervalSec int      `json:"request_interval_seconds"`
	TimeoutSec         int      `json:"timeout_seconds"`
	OutputLogFile      string   `json:"output_log_file"`
}

// Metrics holds the metrics for each request
type Metrics struct {
	Timestamp      time.Time `json:"timestamp"`
	ServerURL      string    `json:"server_url"`
	ResponseTimeMs float64   `json:"response_time_ms"`
	StatusCode     int       `json:"status_code"`
	ContentType    string    `json:"content_type"`
	InferredOS     string    `json:"inferred_os,omitempty"`
	InferredLang   string    `json:"inferred_language,omitempty"`
	ResponseBody   string    `json:"response_body"`
	Error          string    `json:"error,omitempty"`
}

// LogData holds an array of Metrics
type LogData struct {
	Entries []Metrics `json:"entries"`
}

// loadConfig reads the client configuration from a JSON file
func loadConfig(filepath string) (ClientConfig, error) {
	var config ClientConfig
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}

// inferServerDetails tries to deduce server OS and language based on response patterns and headers
func inferServerDetails(resp *http.Response, responseBody string) (string, string) {
	var inferredOS string
	var inferredLang string

	// Read custom headers
	inferredLang = resp.Header.Get("X-Server-Language")
	inferredOS = resp.Header.Get("X-Server-OS")

	// Fallback to existing heuristics if headers are absent
	if inferredLang == "" || inferredOS == "" {
		// Existing heuristic-based inference
		if strings.Contains(responseBody, "Hello, World!") {
			if strings.HasPrefix(responseBody, "{") {
				inferredLang = "Golang or Python (JSON)"
			} else if strings.HasPrefix(responseBody, "---") {
				inferredLang = "Golang or Python (YAML)"
			} else {
				inferredLang = "Unknown Language (Plain Text)"
			}
		}

		if strings.Contains(resp.Header.Get("Content-Type"), "json") {
			inferredLang = "Golang or Python (JSON)"
		} else if strings.Contains(resp.Header.Get("Content-Type"), "yaml") || strings.Contains(resp.Header.Get("Content-Type"), "yml") {
			inferredLang = "Golang or Python (YAML)"
		} else {
			inferredLang = "Plain Text or Unknown"
		}
	}

	return inferredOS, inferredLang
}

// sendRequest sends an HTTP GET request to the specified server and measures response time
func sendRequest(client *http.Client, serverURL string) Metrics {
	metrics := Metrics{
		Timestamp: time.Now(),
		ServerURL: serverURL,
	}

	start := time.Now()
	resp, err := client.Get(serverURL)
	elapsed := time.Since(start).Seconds() * 1000 // Convert to milliseconds
	metrics.ResponseTimeMs = elapsed

	if err != nil {
		metrics.Error = err.Error()
		return metrics
	}
	defer resp.Body.Close()

	metrics.StatusCode = resp.StatusCode

	contentType := resp.Header.Get("Content-Type")
	metrics.ContentType = contentType

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		metrics.Error = err.Error()
		return metrics
	}
	metrics.ResponseBody = string(bodyBytes)

	// Infer server details
	inferredOS, inferredLang := inferServerDetails(resp, metrics.ResponseBody)
	metrics.InferredOS = inferredOS
	metrics.InferredLang = inferredLang

	return metrics
}

// appendLog appends a Metrics entry to the log file
func appendLog(logFile string, entry Metrics, mu *sync.Mutex) error {
	mu.Lock()
	defer mu.Unlock()

	var logData LogData

	// Check if log file exists
	if _, err := os.Stat(logFile); err == nil {
		data, err := ioutil.ReadFile(logFile)
		if err != nil {
			return err
		}
		if len(data) > 0 {
			err = json.Unmarshal(data, &logData)
			if err != nil {
				return err
			}
		}
	}

	// Append new entry
	logData.Entries = append(logData.Entries, entry)

	// Write back to file
	updatedData, err := json.MarshalIndent(logData, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(logFile, updatedData, 0644)
}

func main() {
	// Load client configuration
	config, err := loadConfig("../config/client_config.json")
	if err != nil {
		log.Fatalf("Error loading client configuration: %v", err)
	}

	// Initialize HTTP client with timeout
	httpClient := &http.Client{
		Timeout: time.Duration(config.TimeoutSec) * time.Second,
	}

	// Initialize log file
	logFile := config.OutputLogFile
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		initialLog := LogData{Entries: []Metrics{}}
		data, _ := json.MarshalIndent(initialLog, "", "    ")
		ioutil.WriteFile(logFile, data, 0644)
	}

	var mu sync.Mutex

	// Create a ticker based on request interval
	ticker := time.NewTicker(time.Duration(config.RequestIntervalSec) * time.Second)
	defer ticker.Stop()

	log.Printf("Starting client. Sending requests every %d seconds.", config.RequestIntervalSec)

	for {
		select {
		case <-ticker.C:
			for _, server := range config.Servers {
				go func(srv string) {
					metrics := sendRequest(httpClient, srv)
					err := appendLog(logFile, metrics, &mu)
					if err != nil {
						log.Printf("Error logging metrics for %s: %v", srv, err)
					}

					// Print metrics to console
					printMetrics(metrics)
				}(server)
			}
		}
	}
}

// printMetrics prints the Metrics data in a readable format
func printMetrics(m Metrics) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Timestamp: %s\n", m.Timestamp.Format(time.RFC3339)))
	buffer.WriteString(fmt.Sprintf("Server URL: %s\n", m.ServerURL))
	if m.Error != "" {
		buffer.WriteString(fmt.Sprintf("Error: %s\n", m.Error))
	} else {
		buffer.WriteString(fmt.Sprintf("Status Code: %d\n", m.StatusCode))
		buffer.WriteString(fmt.Sprintf("Content-Type: %s\n", m.ContentType))
		buffer.WriteString(fmt.Sprintf("Response Time: %.2f ms\n", m.ResponseTimeMs))
		buffer.WriteString(fmt.Sprintf("Inferred Language: %s\n", m.InferredLang))
		buffer.WriteString(fmt.Sprintf("Inferred OS: %s\n", m.InferredOS))
		buffer.WriteString(fmt.Sprintf("Response Body: %s\n", truncateString(m.ResponseBody, 200)))
	}
	buffer.WriteString(strings.Repeat("-", 50) + "\n")
	fmt.Print(buffer.String())
}

// truncateString truncates a string to the specified length with ellipsis
func truncateString(str string, num int) string {
	if len(str) > num {
		return str[0:num] + "..."
	}
	return str
}
