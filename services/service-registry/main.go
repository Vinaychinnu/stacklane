package main

import "log"

func main() {
	server := NewServer()

	log.Println("Service-registry started on :8082")

	err := server.Start()
	if err != nil {
		log.Fatal(err)
	}
}