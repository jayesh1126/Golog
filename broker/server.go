package broker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"golog/protocol"
	"net"
	"sort"
)

// Start begins listening for incoming TCP connections on the specified address.
func (b *Broker) Start(address string) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Println("Broker listening on", address)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		// Handle each connection in a separate goroutine for concurrency
		go b.handleConnection(conn)
	}
}

// handleConnection processes incoming messages from a client (produce/fetch requests).
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
			b.handleProduce(conn, line)
		case "fetch":
			b.handleFetch(conn, line)
		}
	}
}

// handleProduce processes a produce request.
func (b *Broker) handleProduce(conn net.Conn, line string) {
	var req protocol.ProduceRequest
	if err := json.Unmarshal([]byte(line), &req); err != nil {
		conn.Write([]byte("invalid request\n"))
		return
	}

	p, err := b.getOrCreatePartition(req.Topic, req.Partition)
	if err != nil {
		conn.Write([]byte("partition error\n"))
		return
	}

	offset, err := p.Append(req.Message)
	if err != nil {
		conn.Write([]byte("write error\n"))
		return
	}

	conn.Write([]byte(fmt.Sprintf("ack offset %d\n", offset)))
}

// handleFetch processes a fetch request.
func (b *Broker) handleFetch(conn net.Conn, line string) {
	var req protocol.FetchRequest
	json.Unmarshal([]byte(line), &req)

	p, err := b.getOrCreatePartition(req.Topic, req.Partition)
	if err != nil {
		conn.Write([]byte("partition error\n"))
		return
	}

	// Sort segments by base offset (defensive, should already be sorted)
	p.mu.RLock()
	segments := make([]*Segment, len(p.segments))
	copy(segments, p.segments)
	p.mu.RUnlock()

	sort.Slice(segments, func(i, j int) bool {
		return segments[i].baseOffset < segments[j].baseOffset
	})

	messages := p.Fetch(req.Offset)
	for _, msg := range messages {
		conn.Write([]byte(msg + "\n"))
	}
}
