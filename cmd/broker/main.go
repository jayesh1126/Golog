package main

import (
    "golog/broker"
)

func main() {
    b := broker.NewBroker()
    b.Start(":9092")
}