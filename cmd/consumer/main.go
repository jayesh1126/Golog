package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golog/protocol"
)

// Reads the offset from disk for this consumer group
func loadOffset(consumerGroup, topic string, partition int) int64 {
	os.MkdirAll("storage", os.ModePerm)
	offsetFile := filepath.Join("storage", fmt.Sprintf("%s-%s-%d.offset", consumerGroup, topic, partition))

	data, err := os.ReadFile(offsetFile)
	if err != nil {
		// File doesn't exist yet, start from 0
		return 0
	}

	offset, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0
	}

	return offset
}

// Writes the offset to disk for this consumer group
func saveOffset(consumerGroup, topic string, partition int, offset int64) error {
	os.MkdirAll("storage", os.ModePerm)
	offsetFile := filepath.Join("storage", fmt.Sprintf("%s-%s-%d.offset", consumerGroup, topic, partition))

	return os.WriteFile(offsetFile, []byte(fmt.Sprintf("%d\n", offset)), 0644)
}

// Consumer connects to the broker and polls for messages from a specific topic/partition.
func main() {
	// Parse command-line flags
	topic := flag.String("topic", "", "Topic to consume from")
	partition := flag.Int("partition", 0, "Partition to consume from")
	consumerGroup := flag.String("group", "default", "Consumer group name")
	pollInterval := flag.Duration("poll-interval", 1*time.Second, "Polling interval")
	flag.Parse()

	// If topic not provided, ask user
	if *topic == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Enter topic: ")
		scanner.Scan()
		*topic = scanner.Text()

		fmt.Print("Enter partition (default 0): ")
		scanner.Scan()
		if s := scanner.Text(); s != "" {
			p, _ := strconv.Atoi(s)
			*partition = p
		}

		fmt.Print("Enter consumer group (default 'default'): ")
		scanner.Scan()
		if s := scanner.Text(); s != "" {
			*consumerGroup = s
		}
	}

	// Connect to broker
	conn, err := net.Dial("tcp", "localhost:9092")
	if err != nil {
		fmt.Println("Error connecting to broker:", err)
		return
	}
	defer conn.Close()

	// Load offset from disk
	offset := loadOffset(*consumerGroup, *topic, *partition)

	fmt.Printf("Connected to broker. Consuming from topic: %s, partition: %d, group: %s\n", *topic, *partition, *consumerGroup)
	fmt.Printf("Starting from offset %d, polling every %v\n\n", offset, *pollInterval)

	reader := bufio.NewReader(conn)

	// Auto-polling loop
	for {
		// Create FetchRequest
		req := protocol.FetchRequest{
			Type:      "fetch",
			Topic:     *topic,
			Partition: *partition,
			Offset:    offset,
		}

		// Marshal to JSON and send
		data, _ := json.Marshal(req)
		conn.Write(append(data, '\n'))

		// Set read timeout so we don't block forever waiting for responses
		conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))

		hasMessages := false
		// Read response(s)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				// Timeout or EOF - no more messages for now
				break
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			fmt.Println(line)
			hasMessages = true

			// Parse offset from response to track our position
			// Response format: "offset N: message"
			if strings.HasPrefix(line, "offset ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					n := 0
					fmt.Sscanf(parts[1], "%d:", &n)
					offset = int64(n + 1) // Move to next offset for next fetch
					// Save offset to disk after processing each message
					saveOffset(*consumerGroup, *topic, *partition, offset)
				}
			}
		}

		if !hasMessages {
			fmt.Println("[No new messages]")
		}

		// Reset deadline for next iteration
		conn.SetReadDeadline(time.Time{})

		// Wait before polling again
		fmt.Printf("\nPolling again in %v... (offset: %d)\n\n", *pollInterval, offset)
		time.Sleep(*pollInterval)
	}
}
