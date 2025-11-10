package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	fmt.Println("🚀 Starting load test...")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	for i := 1; ; i++ {
		start := time.Now()

		resp, err := client.Get("http://localhost/health")
		if err != nil {
			fmt.Printf("❌ Request %d: ERROR - %v\n", i, err)
			time.Sleep(1 * time.Second)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		duration := time.Since(start)

		fmt.Printf("✅ Request %d: %s (%v)\n",
			i,
			string(body),
			duration.Round(time.Millisecond))

		time.Sleep(500 * time.Millisecond)
	}
}
