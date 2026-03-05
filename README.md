# Golog

A minimal Kafka-like message broker written in Go.

Single-node, durable storage with partitioned topics, offset-based consumption, persistent consumer offsets, and auto-polling consumers supporting independent consumer groups.

## Current Features

- Partitioned topics with append-only logs
- Durable message storage
- Offset-based message consumption
- Persistent consumer offsets
- Consumer groups with independent progress tracking
- Concurrent producers and consumers

------------------------------------------------------------------------

## Getting Started

### Start the Broker

``` bash
go run cmd/broker/main.go
```

The broker listens on:

    localhost:9092

Messages are persisted to the `storage/` directory.

------------------------------------------------------------------------

### Run a Producer

``` bash
go run cmd/producer/main.go
```

The producer connects to the broker and waits for JSON input.

Example request:

``` json
{
  "type": "produce",
  "topic": "orders",
  "partition": 0,
  "message": "order-123"
}
```

Response:

    ack offset 0

------------------------------------------------------------------------

### Run a Consumer

``` bash
go run cmd/consumer/main.go --topic orders --partition 0 --group group1 --poll-interval 1s
```
The `--group` flag identifies the consumer group.

Each group tracks its own offset independently, allowing multiple consumers to process the same topic without interfering with each other.

The consumer:

- Polls the broker every 1 second
- Fetches messages from the last committed offset
- Persists offsets to disk after processing
- Supports independent consumer groups

For interactive topic selection (without flags):

``` bash
go run cmd/consumer/main.go
```

------------------------------------------------------------------------

## Architecture

### Broker

-   TCP server built using Go's `net` package
-   Goroutine per client connection
-   Request routing for produce and fetch operations

### Partition

-   Append-only log
-   In-memory index for fast reads
-   Disk-backed persistence
-   Offset derived from log length

### Concurrency

-   `sync.RWMutex` ensures thread-safe partition access
-   Multiple readers allowed concurrently
-   Writes are serialized per partition

### Protocol

-   JSON-based request/response format
-   Simple and human-readable for development clarity

### Consumer Offsets

Consumers persist their current offset to disk after processing messages.  
On restart, the consumer reloads its last committed offset and resumes from that position.

This provides simple fault tolerance and allows consumers to recover from crashes without reprocessing the entire log.
------------------------------------------------------------------------

## Testing the System

### Terminal 1 --- Start Broker

``` bash
go run cmd/broker/main.go
```

### Terminal 2 --- Produce Messages

``` bash
go run cmd/producer/main.go
```

Then enter:

``` json
{
  "type": "produce",
  "topic": "test",
  "partition": 0,
  "message": "hello"
}
```

``` json
{
  "type": "produce",
  "topic": "test",
  "partition": 0,
  "message": "world"
}
```

### Terminal 3 --- Start Consumer (Group 1)

``` bash
go run cmd/consumer/main.go --topic test --partition 0 --group group1
```

You should see the messages appear via auto-polling and the consumer saves its offset to:
storage/group1-test-0.offset

### Terminal 3 --- Restart Consumer (Offset Recovery)

Stop the consumer and run it again
``` bash
go run cmd/consumer/main.go --topic test --partition 0 --group group1
```

The consumer resumes from the last saved offset, avoiding reprocessing old messages.

### Terminal 4 --- Start Another Consumer Group (Group 2)

Stop the consumer and run it again
``` bash
go run cmd/consumer/main.go --topic test --partition 0 --group group2
```

This consumer starts from offset 0 because it has its own independent offset file:
storage/analytics-test-0.offset

Both groups can consume the same topic independently.
------------------------------------------------------------------------

## Design Decisions

-   Immediate disk writes prioritise durability over peak throughput
-   In-memory indexing keeps fetch operations efficient
-   JSON protocol favours simplicity over network efficiency
-   Single-node design focuses on log semantics before distributed
    complexity

------------------------------------------------------------------------

## Next Steps

### Log Segmentation

-   Size-based log rotation
-   Retention policies
-   Segment compaction

### Replication

-   Multi-node deployment
-   Leader-follower model
-   Quorum-based acknowledgements

### Consumer Groups

- Partition rebalancing across consumers
- Broker-managed offset storage
- Heartbeats and group coordination

### Batching

-   Buffered disk writes
-   Reduced syscall overhead
-   Throughput optimisation

### Observability

-   Structured logging
-   Metrics exposure
-   Profiling and benchmarking
