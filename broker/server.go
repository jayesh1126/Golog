package broker

import (
    "bufio"
    "encoding/json"
    "fmt"
    "net"
    "sync"
	"os"

	"golog/protocol"
)

// Partition represents a topic partition with an in-memory log and a file for persistence.
type Partition struct {
	topic string
	id    int
	log   []string
	file  *os.File
	mu    sync.RWMutex
}

// Broker manages topics and partitions, allowing producers to append messages and consumers to fetch them.
type Broker struct {
	topics map[string]map[int]*Partition
	mu     sync.RWMutex
}

// NewBroker initializes a new Broker instance with an empty topics map.
func NewBroker() *Broker {
	return &Broker{
		topics: make(map[string]map[int]*Partition),
	}
}

// getOrCreatePartition retrieves an existing partition for the given topic and ID, or creates a new one if it doesn't exist, ensuring thread safety and loading existing messages from disk.
func (b *Broker) getOrCreatePartition(topic string, id int) (*Partition, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.topics[topic] == nil {
		b.topics[topic] = make(map[int]*Partition)
	}

	if b.topics[topic][id] == nil {

		// ensure storage folder exists
		os.MkdirAll("storage", os.ModePerm)

		filename := fmt.Sprintf("storage/%s-%d.log", topic, id)

		file, err := os.OpenFile(
			filename,
			os.O_APPEND|os.O_CREATE|os.O_RDWR,
			0644,
		)
		if err != nil {
			return nil, err
		}

		p := &Partition{
			topic: topic,
			id:    id,
			log:   []string{},
			file:  file,
		}

		// Load existing messages from disk
		file.Seek(0, 0)
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			p.log = append(p.log, scanner.Text())
		}
		file.Seek(0, 2)

		b.topics[topic][id] = p
	}

	return b.topics[topic][id], nil
}

// Append adds a new message to the partition's in-memory log and appends it to the corresponding file on disk, returning the offset of the new message or an error if the write operation fails.
func (p *Partition) Append(message string) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	offset := len(p.log)

	// append to memory
	p.log = append(p.log, message)

	// append to disk
	_, err := p.file.WriteString(message + "\n")
	if err != nil {
		return 0, err
	}

	return offset, nil
}

// Start begins listening for incoming TCP connections on the specified address and handles them concurrently.
func (b *Broker) Start(address string) {
    ln, err := net.Listen("tcp", address)
    if err != nil {
        panic(err)
    }
    defer ln.Close()
    fmt.Println("Broker listening on", address)

    for {
        conn, _ := ln.Accept()
		// Handle each connection in a separate goroutine to allow concurrent processing of multiple clients.
        go b.handleConnection(conn)
    }
}

// handleConnection processes incoming messages from a client connection, handling both produce and fetch requests based on the message type.
func (b *Broker) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Text()

		var base struct {
			Type string `json:"type"`
		}

		err := json.Unmarshal([]byte(line), &base)
		if err != nil {
			conn.Write([]byte("invalid json\n"))
			continue
		}

		switch base.Type {

		case "produce":
			var req protocol.ProduceRequest
			json.Unmarshal([]byte(line), &req)

			p, err := b.getOrCreatePartition(req.Topic, req.Partition)
			if err != nil {
				conn.Write([]byte("partition error\n"))
				continue
			}

			offset, err := p.Append(req.Message)
			if err != nil {
				conn.Write([]byte("write error\n"))
				continue
			}

			conn.Write([]byte(fmt.Sprintf("ack offset %d\n", offset)))

		case "fetch":
			var req protocol.FetchRequest
			json.Unmarshal([]byte(line), &req)

			p, err := b.getOrCreatePartition(req.Topic, req.Partition)
			if err != nil {
				conn.Write([]byte("partition error\n"))
				continue
			}

			p.mu.RLock()
			if int(req.Offset) < len(p.log) {
				for i := int(req.Offset); i < len(p.log); i++ {
					conn.Write([]byte(fmt.Sprintf("offset %d: %s\n", i, p.log[i])))
				}
			}
			p.mu.RUnlock()
		}
	}
}