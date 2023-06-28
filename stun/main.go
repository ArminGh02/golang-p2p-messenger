package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ArminGh02/golang-p2p-messenger/internal/peer"
	"github.com/ArminGh02/golang-p2p-messenger/internal/stun"
	"github.com/ArminGh02/golang-p2p-messenger/internal/stun/redis"
)

func main() {
	logger := logrus.New()
	logger.Out = os.Stdout

	redis, err := redis.New[*peer.Peer](&redis.Config{
		URL: "redis://localhost:6379",
	})
	if err != nil {
		logger.Fatalln("Error instantiating Redis:", err)
	}

	pong, err := redis.Ping(context.Background())
	if err != nil {
		logger.Fatalln("Error pinging Redis:", err)
	}

	logger.Infoln("Connected to Redis:", pong)

	stun := stun.New(redis, logger)

	mux := http.NewServeMux()
	mux.Handle("/peer/", stun.PeerHandler())
	// TODO add healthz

	// TODO use config
	srv := http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
		// ErrorLog:  ?,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)

		sig := <-c
		logger.Println("Got signal:", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.Fatalln(err)
		}
	}()

	logger.Infoln("Starting server on address", srv.Addr)
	logger.Fatalln(srv.ListenAndServe())
}
