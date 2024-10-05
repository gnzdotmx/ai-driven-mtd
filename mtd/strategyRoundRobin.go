package mtd

import (
	"errors"
	"sync"
	"time"
)

// RoundRobinStrategy implements the round-robin algorithm
type RoundRobinStrategy struct {
	mu           sync.Mutex
	currentIndex int
}

// NewRoundRobinStrategy creates a new RoundRobinStrategy
func NewRoundRobinStrategy() *RoundRobinStrategy {
	return &RoundRobinStrategy{}
}

// Decide selects the next movement using round-robin
func (s *RoundRobinStrategy) Decide(metrics Metrics, config Config) (MovementDecision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(config.Ports) == 0 || len(config.OSes) == 0 || len(config.Formats) == 0 || len(config.Languages) == 0 {
		return MovementDecision{}, errors.New("configuration lists cannot be empty")
	}

	decision := MovementDecision{
		Port:      config.Ports[s.currentIndex%len(config.Ports)],
		OS:        config.OSes[s.currentIndex%len(config.OSes)],
		Format:    config.Formats[s.currentIndex%len(config.Formats)],
		Language:  config.Languages[s.currentIndex%len(config.Languages)],
		Strategy:  RoundRobin,
		Timestamp: time.Now(),
	}

	s.currentIndex++
	return decision, nil
}
