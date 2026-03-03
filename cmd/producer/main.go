package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
)

// Producer is a simple client that connects to the broker and sends messages read from standard input.
func main() {
    conn, _ := net.Dial("tcp", "localhost:9092")
    defer conn.Close()
    scanner := bufio.NewScanner(os.Stdin)

    for scanner.Scan() {
        line := scanner.Text()
        conn.Write([]byte(line + "\n"))
        reply := make([]byte, 1024)
        n, _ := conn.Read(reply)
        fmt.Println(string(reply[:n]))
    }
}