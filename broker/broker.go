package broker

import (
	"sync"
)

// Broker manages topics and partitions, allowing producers to append messages and consumers to fetch them.
type Broker struct {
	// The topics map is a nested map where the first key is the topic name and the second key is the partition ID,
	// mapping to a Partition struct that manages a sequence of log segments for durable message storage.
	topics map[string]map[int]*Partition
	mu     sync.RWMutex
}

// NewBroker initializes a new Broker instance with an empty topics map.
// Returns pointer for shared state across the application.
// Passing by value would create a copy; changes wouldn't be reflected.
// By returning a pointer, all parts of the application work with the same Broker instance.
func NewBroker() *Broker {
	return &Broker{
		topics: make(map[string]map[int]*Partition),
	}
}

// getOrCreatePartition retrieves an existing partition or creates a new one if it doesn't exist.
// Ensures thread safety and loads existing messages from disk.
func (b *Broker) getOrCreatePartition(topic string, id int) (*Partition, error) {
	// Locking the broker's mutex to ensure thread safety when accessing and modifying the topics
	b.mu.Lock()
	// Unlocking is deferred to ensure it happens even if there's an error during partition creation or retrieval.
	// This prevents potential deadlocks and ensures that the broker's state remains consistent.
	defer b.mu.Unlock()

	if b.topics[topic] == nil {
		b.topics[topic] = make(map[int]*Partition)
	}

	if b.topics[topic][id] == nil {
		p, err := NewPartition(topic, id)
		if err != nil {
			return nil, err
		}
		b.topics[topic][id] = p
	}

	return b.topics[topic][id], nil
}
