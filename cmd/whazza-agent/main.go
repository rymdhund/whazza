package main

import (
	"fmt"
	"net/http"
)

func main() {
	err := ping("localhost", 8080)
	if err != nil {
		fmt.Printf("Error pinging server: %s\n", err)
	} else {
		fmt.Println("Ping!")
	}
}

func ping(server string, port int) error {
	url := fmt.Sprintf("http://%s:%d/agent/ping", server, port)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Invalid status: %d", resp.StatusCode)
	}
	return nil
}
