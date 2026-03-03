# Golog

A minimal Kafka-like message broker written in Go.

Single-node, durable storage with partitioned topics, offset-based
consumption, and auto-polling consumers.

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
go run cmd/consumer/main.go --topic orders --partition 0 --poll-interval 1s
```

The consumer:

-   Polls the broker every 1 second
-   Fetches new messages from the latest offset
-   Tracks offsets automatically

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

### Terminal 3 --- Consume Messages

``` bash
go run cmd/consumer/main.go --topic test --partition 0
```

You should see the messages appear via auto-polling.

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

-   Partition rebalancing across consumers
-   Persistent offset commits
-   Stronger delivery guarantees

### Batching

-   Buffered disk writes
-   Reduced syscall overhead
-   Throughput optimisation

### Observability

-   Structured logging
-   Metrics exposure
-   Profiling and benchmarking
