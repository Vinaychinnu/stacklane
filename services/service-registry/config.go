package main

import "os"

type Config struct {
	Port string
}

func LoadConfig() Config {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8082"
	}

	return Config{
		Port: port,
	}
}
