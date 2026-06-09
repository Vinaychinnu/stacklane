package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func NewServer() *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)

	return &Server{
		httpServer: &http.Server{
			Addr: ":8082",
			Handler: loggingMiddleware(mux),
		},
	}
}

func (s *Server) Start() error {
	go func() {
		log.Println("https server listening on :8082")

		err := s.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed{
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
	if err != nil{
		log.Fatal(err)
	}

	log.Println("server stopped gracefully")
}

func healthHandler(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)
	})
}