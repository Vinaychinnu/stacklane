package main

import "log"

func main() {
	config := LoadConfig()

	server := NewServer(config)

	log.Printf("Service-registry started on :%s", config.Port)

	err := server.Start()
	if err != nil {
		log.Fatal(err)
	}
}