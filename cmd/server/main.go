package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url := "https://google.com"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	start := time.Now()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// If the timeout is reached, client.Do returns an error
		fmt.Printf("Request failed (possibly due to timeout): %v\n", err)
		return
	}

	latency := time.Since(start)

	defer resp.Body.Close()

	success := resp.StatusCode >= 200 && resp.StatusCode < 400
	if success {
		fmt.Printf("Response Code: %d\n", resp.StatusCode)
		fmt.Printf("Response Status: %v\n", resp.Status)
		fmt.Printf("Latency: %v\n", latency)
		fmt.Printf("Success: %t\n", success)
	} else {
		fmt.Printf("Request Unseccessful: %d", resp.StatusCode)
	}
}
