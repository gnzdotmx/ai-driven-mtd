package mtd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
)

type Criteria struct {
	ResponseTimeMs     float64 `json:"response_time_ms"`
	ErrorRate          float64 `json:"error_rate"`
	VulnerabilityCount int     `json:"vulnerability_count"`
	IntrusionAttempts  int     `json:"intrusion_attempts"`
}

type RecommendedActions struct {
	SwitchLanguage string `json:"switch_language"`
	SwitchFormat   string `json:"switch_format"`
	SwitchOS       string `json:"switch_os"`
	RotateIP       bool   `json:"rotate_ip"`
}

type Policy struct {
	PolicyName         string             `json:"policy_name"`
	Criteria           Criteria           `json:"criteria"`
	RecommendedActions RecommendedActions `json:"recommended_actions"`
}

// InitializeElasticsearch initializes and returns an Elasticsearch client
func InitializeElasticsearch() (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			os.Getenv("ELASTICSEARCH_URL"),
		},
		Username: os.Getenv("ELASTICSEARCH_USER"),
		Password: os.Getenv("ELASTICSEARCH_PASSWORD"),
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// Ping to verify connection
	res, err := es.Info()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("error connecting to Elasticsearch: %s", res.String())
	}

	log.Println("Connected to Elasticsearch")
	return es, nil
}

// FetchKnowledge fetches relevant knowledge based on current metrics
func ElasticSearch(es *elasticsearch.Client, metrics Metrics) ([]Policy, error) {
	log.Printf(`
	Searching on Elasticsearch for:
		response time: %f
		error rate: %f
		vulnerability count: %d
		intrusion attempts: %d
	`, metrics.QualityOfService.ResponseTimeMs, metrics.QualityOfService.ErrorRate,
		metrics.SecurityMetrics.VulnerabilityCount, metrics.SecurityMetrics.IntrusionAttempts)

	query := map[string]interface{}{
		"size": 5,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"multi_match": map[string]interface{}{
							"query":  metrics.QualityOfService.ResponseTimeMs,
							"fields": []string{"criteria.response_time_ms"},
							"type":   "best_fields",
						},
					},
					{
						"multi_match": map[string]interface{}{
							"query":  metrics.QualityOfService.ErrorRate,
							"fields": []string{"criteria.error_rate"},
							"type":   "best_fields",
						},
					},
					{
						"multi_match": map[string]interface{}{
							"query":  metrics.SecurityMetrics.VulnerabilityCount,
							"fields": []string{"criteria.vulnerability_count"},
							"type":   "best_fields",
						},
					},
					{
						"multi_match": map[string]interface{}{
							"query":  metrics.SecurityMetrics.IntrusionAttempts,
							"fields": []string{"criteria.intrusion_attempts"},
							"type":   "best_fields",
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(os.Getenv("ELASTICSEARCH_INDEX")),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error fetching knowledge: %s", res.String())
	}

	// Parse response
	var esResponse struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResponse); err != nil {
		return nil, err
	}

	if esResponse.Hits.Total.Value == 0 {
		return nil, fmt.Errorf("no knowledge data found")
	}

	// Load all retrieved policies
	policies := make([]Policy, 0)
	for _, hit := range esResponse.Hits.Hits {
		var policy Policy
		sourceBytes, err := json.Marshal(hit.Source)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(sourceBytes, &policy); err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	return policies, nil
}
