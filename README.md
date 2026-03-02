# Golog

Golog is a distributed log-based message broker written in Go.

It is inspired by the core design principles of systems like Apache Kafka, but implemented from scratch to understand the mechanics of:

Append-only log storage

Partitioned topics

Offset-based consumption

Consumer groups

At-least-once delivery

Backpressure and batching

The goal of this project is educational and architectural: to build a minimal but correct log-based messaging system and explore the trade-offs involved in durability, concurrency, and performance.

# Motivation

I’ve worked extensively with Kafka in production systems (real-time event pipelines, WebSocket streaming, distributed services). While using Kafka is straightforward, understanding its internals — log segmentation, offset tracking, partitioning, replication — requires building one.

This project exists to:

Move beyond using a broker to implementing one

Explore concurrency primitives in Go

Understand durability and I/O trade-offs

Benchmark throughput and latency under load
