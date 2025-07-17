package main

import (
    "fmt"
    "os"
    "time"
)

func main() {
    config := os.Getenv("CONFIG")
    if config == "" {
        fmt.Println("No CONFIG found")
        return
    }
    fmt.Println("Loaded config:")
    fmt.Println(config)
    fmt.Println("Starting Twickets bot...")
    for {
        fmt.Println("Polling Twickets event...")
        time.Sleep(15 * time.Second)
    }
}