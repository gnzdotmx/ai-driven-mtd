package mtd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mtd-system/ollama"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

// WeightedStrategy implements a weighted decision-making algorithm
type WeightedStrategy struct {
	weights  MetricsWeights
	settings StrategySettings
}

// MetricsWeights holds the weights for different metric categories
type MetricsWeights struct {
	QualityOfService float64 `json:"quality_of_service"`
	SecurityMetrics  float64 `json:"security_metrics"`
	AssetValue       float64 `json:"asset_value"`
}

// NewWeightedStrategy creates a new WeightedStrategy
func NewWeightedStrategy(weights MetricsWeights, settings StrategySettings) *WeightedStrategy {
	return &WeightedStrategy{
		weights:  weights,
		settings: settings,
	}
}

func (s *WeightedStrategy) Decide(metrics Metrics, config Config, es *elasticsearch.Client) (MovementDecision, error) {
	if len(config.Ports) == 0 || len(config.OSes) == 0 || len(config.Formats) == 0 || len(config.Languages) == 0 {
		return MovementDecision{}, errors.New("configuration lists cannot be empty")
	}

	// Fetch knowledge data
	knowledge, err := ElasticSearch(es, metrics)
	if err != nil {
		log.Printf("Error fetching knowledge: %v", err)
		log.Printf("Moving to a weighted decision without elastic search knowledge")
		return s.fallbackDecide(metrics, config)
	}

	var prevDecisions string
	for _, action := range knowledge {
		prevDecisions += fmt.Sprintf("\t\t%+v\n", action)
	}
	log.Printf("Best matches on Elasticsearch:\n%s", prevDecisions)

	// Construct prompt for Ollama based on decision and knowledge
	prompt := fmt.Sprintf(`
	You are a cloud security expert. You are good at designing and implementing secure cloud environments.
	Your tasks is to analyze the current configuration of the cloud environment and make recommendations for improvements based on previous decisions.
	Given the CURRENT METRICS tell me what configuration changes should be made based on PREVIOUS DECISIONS
	Make sure that the output is in JSON format defined in OUTPUT FORMAT is valid and well-formatted.

	CURRENT METRICS:
	Quality of Service:
		Response Time: %f ms
		Error Rate: %f
	Security Metrics:
		Vulnerability Count: %d
		Intrusion Attempts: %d
	Asset Value: 
		Critical Assets: %d
		High Value Assets: %d
	Strategy settings:
		Thresholds:
			Response Time: %f ms
			Error Rate: %f
			Vulnerability Count: %d
			Intrusion Attempts: %d
		Weights:
			Quality of Service: %f
			Security Metrics: %f
			Asset Value: %f

	PREVIOUS DECISIONS:
%s

	OUTPUT FORMAT:
	{"SwitchLanguage": "python", "SwitchOS": "ubuntu", "SwitchFormat": "json", "SwitchPort": "80", "RotateIP": "true"}
		`, metrics.QualityOfService.ResponseTimeMs, metrics.QualityOfService.ErrorRate,
		metrics.SecurityMetrics.VulnerabilityCount, metrics.SecurityMetrics.IntrusionAttempts,
		metrics.AssetValue.CriticalAssets, metrics.AssetValue.HighValueAssets,
		s.settings.Thresholds.ResponseTimeMs, s.settings.Thresholds.ErrorRate,
		s.settings.Thresholds.VulnerabilityCount, s.settings.Thresholds.IntrusionAttempts,
		s.weights.QualityOfService, s.weights.SecurityMetrics,
		s.weights.AssetValue, prevDecisions)

	// log.Printf("\nUser> \n%s", prompt)
	// Ask Ollama for final decision
	var oLlamaerror = false
	ollamaAnswer, err := ollama.AskOllama(prompt)
	if err != nil {
		log.Printf("Error querying Ollama: %v", err)
		oLlamaerror = true
		log.Printf("Moving to a weighted decision using elastic search knowledge, without Ollama recommendation")
		// return s.fallbackDecide(metrics, config)
		// mu.Unlock()
		// continue
	}

	log.Printf("\n\t\t\tOllama> %s", ollamaAnswer)

	// Parse Ollama response
	// {SwitchLanguage: "python", SwitchOS: "ubuntu", SwitchFormat: "json", SwitchPort: "80", RotateIP: true}
	var resp map[string]string
	err = json.Unmarshal([]byte(ollamaAnswer), &resp)
	if err != nil {
		oLlamaerror = true
		log.Printf("Error parsing Ollama response: %v", err)
		log.Printf("Moving to a weighted decision using elastic search knowledge, without Ollama recommendation")
	}

	if oLlamaerror {
		resp = make(map[string]string)
		resp["SwitchLanguage"] = knowledge[0].RecommendedActions.SwitchLanguage
		resp["SwitchOS"] = knowledge[0].RecommendedActions.SwitchOS
		resp["SwitchFormat"] = knowledge[0].RecommendedActions.SwitchFormat
	}

	// Apply recommended actions
	decision := MovementDecision{
		// IP:        config.IPs[rand.Intn(len(config.IPs))],
		OS:        resp["SwitchOS"],
		Format:    resp["SwitchFormat"],
		Language:  resp["SwitchLanguage"],
		Strategy:  Weighted,
		Timestamp: time.Now(),
	}

	// // Optionally, handle rotate_ip
	// rotateIP, _ := recommendedActions["rotate_ip"].(bool)
	// if rotateIP {
	// 	decision.IP = config.IPs[rand.Intn(len(config.IPs))]
	// }

	return decision, nil
}

// Decide selects the next movement based on weighted scores
func (s *WeightedStrategy) fallbackDecide(metrics Metrics, config Config) (MovementDecision, error) {
	if len(config.Ports) == 0 || len(config.OSes) == 0 || len(config.Formats) == 0 || len(config.Languages) == 0 {
		return MovementDecision{}, errors.New("configuration lists cannot be empty")
	}

	// Calculate scores for each category
	qosScore := calculateQoSScore(metrics.QualityOfService, s.settings.Thresholds)
	securityScore := calculateSecurityScore(metrics.SecurityMetrics, s.settings.Thresholds)
	assetScore := calculateAssetScore(metrics.AssetValue)

	// Weighted total score
	totalScore := qosScore*s.weights.QualityOfService + securityScore*s.weights.SecurityMetrics + assetScore*s.weights.AssetValue

	// Normalize score to select a movement
	// Higher totalScore implies higher priority to change
	// For simplicity, let's use totalScore to influence the probability of selecting a random movement
	threshold := 50.0 // Define a threshold based on your scoring system

	var strategy StrategyType
	if totalScore > threshold {
		strategy = Random
	} else {
		strategy = RoundRobin
	}

	var decision MovementDecision
	var err error

	switch strategy {
	case Random:
		decision, err = NewRandomStrategy().Decide(metrics, config)
	case RoundRobin:
		decision, err = NewRoundRobinStrategy().Decide(metrics, config)
	default:
		err = errors.New("unknown strategy type")
	}

	if err != nil {
		return MovementDecision{}, err
	}

	decision.Strategy = Weighted
	decision.Score = totalScore
	return decision, nil
}
