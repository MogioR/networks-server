package main

import (
	"context"
	"encoding/json"
	"fmt"
	"messager-server/internal/config"
	"messager-server/internal/database"
	"messager-server/internal/messager"
	routes "messager-server/internal/route"
	"messager-server/internal/storage"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"
	log "github.com/sirupsen/logrus"
)

func main() {
	data, err := os.ReadFile("./configs/config.json")
	if err != nil {
		log.Fatal(err)
	}
	cfg := new(config.Config)
	if err := json.Unmarshal(data, cfg); err != nil {
		log.Fatal(err)
	}
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}

	logger := log.New()
	loglevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	logger.SetLevel(loglevel)
	psDb, err := database.ConnectToPostgres(&cfg.Postgres)
	if err != nil {
		log.Fatal(err)
	}

	storage := storage.New(psDb, logger)
	m := messager.NewMessager(storage)
	router := routes.New(storage, cfg, logger, m)

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("TCP server started on port %s\n", cfg.Api.TCPPort)
	ln, err := net.Listen("tcp", cfg.Api.TCPPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer ln.Close()

	defer cancel()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Warn("Error accepting connection:", err.Error())
				continue
			}
			fmt.Println("Client connected:", conn.RemoteAddr())

			if userId, err := m.Auth(conn); err == nil {
				go m.ConsumerHeandler(ctx, conn, userId)
			} else {
				fmt.Println(err)
				conn.Close()
			}
		}
	}()

	fmt.Printf("HTTP server started on port %s\n", cfg.Api.HTTPPort)
	go func() {
		if err := router.Start(); err != nil {
			errChan <- err
		}
	}()

	select {
	case sig := <-sigChan:
		log.Info("Get signal: ", sig)
		cancel()
	case err := <-errChan:
		log.Warn(err)
	}
}
