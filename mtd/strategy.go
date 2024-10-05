package mtd

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"time"
)

// Metrics represents the structure of metrics.json
type Metrics struct {
	QualityOfService struct {
		ResponseTimeMs float64 `json:"response_time_ms"`
		ErrorRate      float64 `json:"error_rate"`
	} `json:"quality_of_service"`
	SecurityMetrics struct {
		VulnerabilityCount int `json:"vulnerability_count"`
		IntrusionAttempts  int `json:"intrusion_attempts"`
	} `json:"security_metrics"`
	AssetValue struct {
		CriticalAssets  int `json:"critical_assets"`
		HighValueAssets int `json:"high_value_assets"`
	} `json:"asset_value"`
	StrategySettings struct {
		Thresholds struct {
			ResponseTimeMs     float64 `json:"response_time_ms"`
			ErrorRate          float64 `json:"error_rate"`
			VulnerabilityCount int     `json:"vulnerability_count"`
			IntrusionAttempts  int     `json:"intrusion_attempts"`
		} `json:"thresholds"`
		Weights struct {
			QualityOfService float64 `json:"quality_of_service"`
			SecurityMetrics  float64 `json:"security_metrics"`
			AssetValue       float64 `json:"asset_value"`
		} `json:"weights"`
	} `json:"strategy_settings"`
}

// StrategyType defines the type of strategy to use
type StrategyType string

const (
	RoundRobin StrategyType = "round_robin"
	Random     StrategyType = "random"
	Weighted   StrategyType = "weighted"
)

// MovementDecision encapsulates the decision for movement
type MovementDecision struct {
	IP        string
	Port      string
	OS        string
	Format    string
	Language  string
	Strategy  StrategyType
	Score     float64 // Used for weighted strategy
	Timestamp time.Time
}

// Strategy defines the interface for different strategies
type Strategy interface {
	Decide(metrics Metrics, config Config) (MovementDecision, error)
}

// Config represents the configuration for strategies
type Config struct {
	IPs       []string `json:"ips"`
	Ports     []string `json:"ports"`
	OSes      []string `json:"oses"`
	Formats   []string `json:"formats"`
	Languages []string `json:"languages"`
}

// StrategySettings holds thresholds
type StrategySettings struct {
	Thresholds struct {
		ResponseTimeMs     float64 `json:"response_time_ms"`
		ErrorRate          float64 `json:"error_rate"`
		VulnerabilityCount int     `json:"vulnerability_count"`
		IntrusionAttempts  int     `json:"intrusion_attempts"`
	} `json:"thresholds"`
}

// Helper functions to calculate scores
func calculateQoSScore(qos struct {
	ResponseTimeMs float64 `json:"response_time_ms"`
	ErrorRate      float64 `json:"error_rate"`
}, thresholds struct {
	ResponseTimeMs     float64 `json:"response_time_ms"`
	ErrorRate          float64 `json:"error_rate"`
	VulnerabilityCount int     `json:"vulnerability_count"`
	IntrusionAttempts  int     `json:"intrusion_attempts"`
}) float64 {
	// Lower response time and error rate are better
	responseTimeScore := math.Max(0, thresholds.ResponseTimeMs-qos.ResponseTimeMs)
	errorRateScore := math.Max(0, thresholds.ErrorRate-qos.ErrorRate)
	return responseTimeScore + errorRateScore
}

func calculateSecurityScore(security struct {
	VulnerabilityCount int `json:"vulnerability_count"`
	IntrusionAttempts  int `json:"intrusion_attempts"`
}, thresholds struct {
	ResponseTimeMs     float64 `json:"response_time_ms"`
	ErrorRate          float64 `json:"error_rate"`
	VulnerabilityCount int     `json:"vulnerability_count"`
	IntrusionAttempts  int     `json:"intrusion_attempts"`
}) float64 {
	// Lower counts are better
	vulnScore := math.Max(0, float64(thresholds.VulnerabilityCount)-float64(security.VulnerabilityCount))
	intrusionScore := math.Max(0, float64(thresholds.IntrusionAttempts)-float64(security.IntrusionAttempts))
	return vulnScore + intrusionScore
}

func calculateAssetScore(asset struct {
	CriticalAssets  int `json:"critical_assets"`
	HighValueAssets int `json:"high_value_assets"`
}) float64 {
	// Higher asset value increases the need for movement
	return float64(asset.CriticalAssets*2 + asset.HighValueAssets)
}

// LoadMetrics reads metrics from metrics.json
func LoadMetrics(filepath string) (Metrics, error) {
	var metrics Metrics
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return metrics, err
	}
	err = json.Unmarshal(data, &metrics)
	return metrics, err
}
