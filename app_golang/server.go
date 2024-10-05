package main

import (
	"encoding/json"
	"net/http"

	"os"

	"gopkg.in/yaml.v2"
)

type Message struct {
	Text string `yaml:"text"`
}

var format = os.Getenv("RESPONSE_FORMAT")
var language = os.Getenv("RESPONSE_LANGUAGE")
var ros = os.Getenv("RESPONSE_OS")

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Server-Language", language)
	w.Header().Set("X-Server-OS", ros)
	response := map[string]string{
		"message": "Hello, World!",
	}

	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "yaml":
		w.Header().Set("Content-Type", "application/x-yaml")
		msg := Message{
			Text: "Hello, World!",
		}

		out, err := yaml.Marshal(msg)
		if err != nil {
			yaml.NewEncoder(w).Encode("Error")
		}

		yaml.NewEncoder(w).Encode(string(out))
	default:
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello, World!"))
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
