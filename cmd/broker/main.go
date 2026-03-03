package main

import (
    "golog/broker"
)

// main initializes a new Broker instance and starts it on the specified address, allowing it to accept incoming connections from producers and consumers.
func main() {
    b := broker.NewBroker()
    b.Start(":9092")
}