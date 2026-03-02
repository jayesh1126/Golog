package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
)

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