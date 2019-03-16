package main

import (
    "fmt"
    "time"
)

func main() {
    fmt.Println("Initializing test...")
    // Try to reconnect if connection is closed
    ticker := time.NewTicker(200 * time.Millisecond)
    go func() {
        for range ticker.C {
            fmt.Println("Tick at")
        }
    }()
    time.Sleep(1000 * time.Millisecond)
    defer ticker.Stop()
    time.Sleep(10000 * time.Millisecond)
}

func captureBounds(count int) []uint8 {
	// Get main display's bounds
	// bounds := screenshot.GetDisplayBounds(0)
    // Capture a screenshot

    // Create a wait group and add the four routines


    // Return the LED data
    return nil
}
