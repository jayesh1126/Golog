# Golog

Golog is a distributed log-based message broker written in Go.

It is inspired by the core architectural principles of modern log-based
systems such as Kafka, and was implemented from scratch to deeply
understand the mechanics behind:

-   Append-only log storage
-   Partitioned topics
-   Offset-based consumption
-   Consumer groups
-   At-least-once delivery semantics
-   Backpressure and batching strategies

The goal of this project is educational and architectural: to build a
minimal but correct log-based messaging system and explore the
trade-offs involved in durability, concurrency, and performance.

------------------------------------------------------------------------

# Motivation

In production systems, I have worked extensively with Kafka powering
real-time event pipelines, distributed services, and WebSocket streaming
architectures. While using Kafka is operationally straightforward,
understanding its internal guarantees --- log segmentation, offset
tracking, partition coordination, and recovery --- requires implementing
these mechanisms directly.

Golog exists to:

-   Move from using a broker to implementing one
-   Explore concurrency primitives in Go
-   Understand durability and disk I/O trade-offs
-   Benchmark throughput and latency under load
-   Develop a deeper intuition for distributed log systems

------------------------------------------------------------------------

# Current Architecture

Golog is currently implemented as a single-node broker with durable,
partitioned storage.

## Core Components

### Broker

-   TCP server built using Go's `net` package
-   Goroutine per client connection
-   Request routing for produce and fetch operations
-   Thread-safe partition access using `sync.RWMutex`

### Partition

Each partition encapsulates: - In-memory message index - Append-only
disk log file - Offset tracking - Concurrency control

Messages are written to disk immediately and reconstructed on startup to
ensure durability.

### Storage Model

-   Append-only log per topic/partition
-   File-backed persistence
-   Crash recovery via log replay on broker startup
-   Offset derived from in-memory index length

This mirrors the core storage abstraction used in real log-based
systems.

------------------------------------------------------------------------

# Features Implemented

-   Topic and partition abstraction
-   Offset-based message consumption
-   Durable append-only storage
-   Crash recovery via log replay
-   Concurrent producers and consumers
-   Structured JSON request protocol
-   At-least-once delivery (consumer group groundwork implemented)

------------------------------------------------------------------------

# Concurrency Model

-   Goroutines handle concurrent client connections
-   RWMutex ensures safe read/write access to partition logs
-   Reads allow multiple concurrent consumers
-   Writes are serialized per partition

This design enables safe concurrent access without race conditions while
maintaining simplicity.

------------------------------------------------------------------------

# Running the Project

Start the broker:

    go run cmd/broker/main.go

Run a producer:

    go run cmd/producer/main.go

Run a consumer:

    go run cmd/consumer/main.go

Messages are persisted to the `storage/` directory.

------------------------------------------------------------------------

# Design Trade-offs

-   Immediate disk writes favour durability over peak throughput
-   In-memory indexing keeps fetch operations simple and fast
-   JSON protocol chosen for clarity over network efficiency
-   Single-node design keeps focus on log semantics before distributed
    complexity

------------------------------------------------------------------------

# Benchmarking and Observability

Initial performance testing focuses on:

-   Throughput under concurrent producers
-   Latency of fetch operations
-   Memory usage under sustained load

Future iterations will include profiling with `pprof` and batched write
optimisation.

------------------------------------------------------------------------

# Possible Next Improvements

## 1. Log Segmentation

-   Introduce segmented log files
-   Implement size-based rotation
-   Add retention policies

## 2. Replication

-   Leader-follower partition replication
-   Quorum-based acknowledgements
-   Failure detection and leader election

## 3. Consumer Group Coordination

-   Partition rebalancing across consumers
-   Persistent offset commits
-   Improved at-least-once guarantees

## 4. Performance Optimisation

-   Buffered and batched disk writes
-   Zero-copy network responses
-   Custom binary protocol for reduced overhead

## 5. Backpressure Handling

-   Flow control between broker and consumers
-   Rate limiting for producers
-   Bounded in-memory buffers

## 6. Observability

-   Structured logging
-   Metrics exposure (Prometheus-compatible)
-   Distributed tracing integration

------------------------------------------------------------------------

# Long-Term Vision

The long-term goal is to evolve Golog into a multi-node replicated log
system capable of:

-   Stronger delivery guarantees
-   Horizontal scalability
-   Realistic production failure modes
-   Measurable performance characteristics

The emphasis remains on architectural clarity, correctness, and deep
understanding of distributed log design rather than feature parity.

------------------------------------------------------------------------

# Author

Jay Utchanah\
Software Engineer\

This project reflects a practical exploration of distributed systems,
storage engines, and concurrency in Go.
