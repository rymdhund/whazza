package main

import (
	"fmt"
	"net/http"

	"github.com/rymdhund/whazza/internal/tofu"
)

func main() {
	err := ping("localhost", 4433)
	if err != nil {
		fmt.Printf("Error pinging server: %s\n", err)
	} else {
		fmt.Println("Ping!")
	}
}

func ping(server string, port int) error {
	client, err := tofu.HttpClientFromFile(server, port)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s:%d/agent/ping", server, port)
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Invalid status: %d", resp.StatusCode)
	}
	return nil
}
