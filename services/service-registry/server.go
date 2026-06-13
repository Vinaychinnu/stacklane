package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Server struct {
	httpServer *http.Server
	config     Config
	services   []Service
}

func NewServer(config Config) *Server {

	server := &Server{
		config:   config,
		services: []Service{},
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/services", server.handleServices)

	server.httpServer = &http.Server{
		Addr:    ":" + config.Port,
		Handler: loggingMiddleware(mux),
	}

	return server

}

func (s *Server) Start() error {
	go func() {
		log.Printf("http server listening on :%s", s.config.Port)

		err := s.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	s.waitForShutdown()

	return nil
}

func (s *Server) waitForShutdown() {
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	log.Println("shutdown signal recieved")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("server stopped gracefully")
}

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodGet:
		s.listServices(w)

	case http.MethodPost:
		s.registerService(w, r)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) registerService(w http.ResponseWriter, r *http.Request) {

	var service Service
	err := json.NewDecoder(r.Body).Decode(&service)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(service.Name) == "" {
		http.Error(w, "service name is required", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(service.Team) == "" {
		http.Error(w, "service team is required", http.StatusBadRequest)
		return
	}

	s.services = append(s.services, service)

	w.WriteHeader(http.StatusCreated)

}

func (s *Server) listServices(w http.ResponseWriter) {

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.services)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)
	})
}
