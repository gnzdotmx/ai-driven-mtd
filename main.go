package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"mtd-system/mtd"
	"os"
	"os/exec"
	"time"
)

type Config struct {
	Ports     []string `json:"ports"`
	OSes      []string `json:"oses"`
	Formats   []string `json:"formats"`
	Languages []string `json:"languages"`
}

func loadAppConfig(filepath string) (Config, error) {
	var config Config
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}

func loadMetricsConfig(filepath string) (mtd.Metrics, error) {
	return mtd.LoadMetrics(filepath)
}

func selectRandom(options []string) string {
	rand.Seed(time.Now().UnixNano())
	return options[rand.Intn(len(options))]
}

func executeScript(scriptPath string, arg []string) error {
	cmd := exec.Command(scriptPath, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func executeScriptNoArg(scriptPath string) error {
	cmd := exec.Command("bash", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	config, err := loadAppConfig("config/config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	metrics, err := loadMetricsConfig("config/metrics.json")
	if err != nil {
		log.Fatalf("Error loading metrics config: %v", err)
	}

	// Initialize Elasticsearch
	es, err := mtd.InitializeElasticsearch()
	if err != nil {
		log.Fatalf("Error initializing Elasticsearch: %v", err)
	}

	// Initialize strategies
	// roundRobin := mtd.NewRoundRobinStrategy()
	// randomStrategy := mtd.NewRandomStrategy()
	weightedStrategy := mtd.NewWeightedStrategy(mtd.MetricsWeights{
		QualityOfService: metrics.StrategySettings.Weights.QualityOfService,
		SecurityMetrics:  metrics.StrategySettings.Weights.SecurityMetrics,
		AssetValue:       metrics.StrategySettings.Weights.AssetValue,
	}, mtd.StrategySettings{
		Thresholds: metrics.StrategySettings.Thresholds,
	})

	// Select strategy (WeightedStrategy in this example)
	strategy := weightedStrategy

	// interval := 1.0 * time.Minute
	// fmt.Printf("\nChangeInterval: %d min", interval)
	// ticker := time.NewTicker(interval) // Set desired interval
	// defer ticker.Stop()

	// For demonstration, reduce ticker interval
	// ticker = time.NewTicker(interval)

	// var mu sync.Mutex
	// currentPortIndex := 0
	// currentOSIndex := 0
	// currentFormatIndex := 04
	// currentLanguageIndex := 0

	// for {
	// 	select {
	// 	case <-ticker.C:
	// 		mu.Lock()
	// Round-Robin Selection

	// fmt.Println(len(config.Ports), len(config.OSes), len(config.Formats), len(config.Languages))
	log.Printf("Available configurations:\n\t\t%+v", config)
	decision, err := strategy.Decide(metrics, mtd.Config{
		Ports:     config.Ports,
		OSes:      config.OSes,
		Formats:   config.Formats,
		Languages: config.Languages,
	}, es)
	if err != nil {
		log.Printf("Error deciding movement: %v", err)
	}

	// log.Printf("Applying MTD changes: %+v", decision)

	// mu.Unlock()

	args := []string{"./scripts/set_env.sh", decision.Port, decision.Format, decision.Language, decision.OS}
	err = executeScript("bash", args)
	if err != nil {
		log.Printf("Error switching OS: %v", err)
	}

	log.Printf("MTD changes applied: PORT=%s OS=%s, Format=%s, Language=%s", decision.Port, decision.OS, decision.Format, decision.Language)
	// 	}
	// }
}
