package mtd

import (
	"errors"
	"math/rand"
	"time"
)

// RandomStrategy implements the random selection algorithm
type RandomStrategy struct{}

// NewRandomStrategy creates a new RandomStrategy
func NewRandomStrategy() *RandomStrategy {
	rand.Seed(time.Now().UnixNano())
	return &RandomStrategy{}
}

// Decide selects the next movement randomly
func (s *RandomStrategy) Decide(metrics Metrics, config Config) (MovementDecision, error) {
	if len(config.Ports) == 0 || len(config.OSes) == 0 || len(config.Formats) == 0 || len(config.Languages) == 0 {
		return MovementDecision{}, errors.New("configuration lists cannot be empty")
	}

	decision := MovementDecision{
		Port:      config.Ports[rand.Intn(len(config.Ports))],
		OS:        config.OSes[rand.Intn(len(config.OSes))],
		Format:    config.Formats[rand.Intn(len(config.Formats))],
		Language:  config.Languages[rand.Intn(len(config.Languages))],
		Strategy:  Random,
		Timestamp: time.Now(),
	}

	return decision, nil
}
