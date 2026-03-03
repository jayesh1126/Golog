package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"golog/protocol"
)

// Consumer connects to the broker and polls for messages from a specific topic/partition.
func main() {
	// Parse command-line flags
	topic := flag.String("topic", "", "Topic to consume from")
	partition := flag.Int("partition", 0, "Partition to consume from")
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
	}

	// Connect to broker
	conn, err := net.Dial("tcp", "localhost:9092")
	if err != nil {
		fmt.Println("Error connecting to broker:", err)
		return
	}
	defer conn.Close()

	fmt.Printf("Connected to broker. Consuming from topic: %s, partition: %d\n", *topic, *partition)
	fmt.Printf("Starting from offset 0, polling every %v\n\n", *pollInterval)

	offset := int64(0)
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
