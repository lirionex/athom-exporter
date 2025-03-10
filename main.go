package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type SensorResponse struct {
	ID    string  `json:"id"`
	Value float32 `json:"value"`
	State string  `json:"state"`
}

func FormatOpenMetrics(metricName string, labels map[string]string, value interface{}) string {
	// Sort the labels by key to ensure consistent output
	labelPairs := make([]string, 0, len(labels))
	for k, v := range labels {
		labelPairs = append(labelPairs, fmt.Sprintf("%s=%q", k, v))
	}
	sort.Strings(labelPairs)

	// Join labels into a comma-separated list
	labelStr := ""
	if len(labelPairs) > 0 {
		labelStr = "{" + strings.Join(labelPairs, ",") + "}"
	}

	// Return the formatted string
	return fmt.Sprintf("%s%s %v\n", metricName, labelStr, value)
}

func getSensor(target string, sensor string) (*SensorResponse, error) {
	apiURL := fmt.Sprintf("%s/sensor/%s", target, sensor)
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from target: %w", err)
	}
	defer resp.Body.Close()

	// Check if the HTTP status is OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var result SensorResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &result, nil
}

// getMetrics handles the /metrics endpoint
func getMetrics(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	log.Printf("Received request from %s: %s %s %s", clientIP, r.Method, r.URL.Path, r.URL.RawQuery)

	// Check if the "target" query parameter is provided
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "Missing 'target' query parameter", http.StatusBadRequest)
		return
	}

	sensors := []string{"power", "wifi_signal_db", "voltage", "current"}
	var response strings.Builder

	for _, s := range sensors {
		sensor, err := getSensor(target, s)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching sensor %s: %v", s, err), http.StatusInternalServerError)
			return
		}
		response.WriteString(FormatOpenMetrics("athom_sensor_"+s, map[string]string{"id": sensor.ID}, sensor.Value))
	}

	w.Header().Set("Content-Type", "text/plain")
	if _, err := w.Write([]byte(response.String())); err != nil {
		log.Printf("Error writing response to %s: %v", clientIP, err)
	}
	log.Printf("Served metrics to %s", clientIP)
}

func main() {

	bindAddress := os.Getenv("BIND_ADDRESS")
	if bindAddress == "" {
		bindAddress = ":5573"
	}

	http.HandleFunc("/metrics", getMetrics)
	server := &http.Server{
		Addr:         bindAddress,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting athom-exporter on %s", bindAddress)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
