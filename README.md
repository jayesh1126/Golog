# Golog

A minimal Kafka-like message broker written in Go.

Single-node, durable storage with partitioned topics, offset-based consumption, persistent consumer offsets, and auto-polling consumers supporting independent consumer groups.

## Current Features

-  **Partitioned Topics** with append-only logs
- **Log Segmentation** with automatic rotation (max 10 messages per segment)
- **Durable Message Storage** persisted to disk
- **Offset-based Consumption** with efficient segment indexing
- **Persistent Consumer Offsets** tracked per consumer group
- **Concurrent Producers and Consumers** via goroutines
- **Thread-safe Operations** with fine-grained locking (sync.RWMutex)

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

### Partition & Segments

- Append-only log split into **segments** for scalability
- Each segment stores up to 10 messages
- Segments persist to disk as `{topic}-{partition}-{baseOffset}.log`
- On startup, broker reloads all segments from disk
- Offsets are derived from segment base offset + message index

**Example file structure:**
```
storage/
  orders-0-0.log      (partition 0, messages 0-9)
  orders-0-10.log     (partition 0, messages 10-19)
  orders-1-0.log      (partition 1, messages 0-9)
```
### Consumer Groups

- Multiple consumers can join the same group
- Broker performs **round-robin partition assignment**
- Each partition is assigned to exactly one consumer in a group
- **Rebalancing** occurs when consumers join/leave
- Generation IDs ensure consistency during rebalancing
- **Heartbeats** keep consumers alive (every 3 seconds)

**Example rebalancing:**
```
1 consumer joins:  [consumer-1] → assigned partitions [0, 1, 2, 3]
2nd consumer joins: 
  [consumer-1] → rebalanced to [0, 2]
  [consumer-2] → rebalanced to [1, 3]
```

### Concurrency

-   `sync.RWMutex` ensures thread-safe partition access
-   Multiple readers allowed concurrently
-   Writes are serialized per partition

### Protocol

-   JSON-based request/response format
-   Simple and human-readable for development clarity

### Consumer Offsets

Offsets are persisted to `storage/{consumer-group}-{topic}-{partition}.offset`:
- Each consumer group tracks progress independently
- Offsets saved to disk after processing messages
- On restart, consumers resume from last committed offset
- Enables fault tolerance and crash recovery

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
