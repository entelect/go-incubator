package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"

	config "go-incubator/internal/configuration"
	"go-incubator/internal/http"
	"go-incubator/internal/persistence/memdb"
)

func main() {
	fmt.Println("service started")
	defer fmt.Println("service stopped")

	cfg, err := config.ReadConfig("INCUBATOR_")
	if err != nil {
		fmt.Printf("error reading config: %v\n", err)
		return
	}

	db, err := memdb.NewMemDB()
	if err != nil {
		fmt.Printf("error creating persistence layer: %v\n", err)
		return
	}

	httpServer, err := http.NewHttpServer(cfg.HttpPort, cfg.ApiKey, &db)
	if err != nil {
		fmt.Printf("error creating http server: %v\n", err)
		return
	}

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	httpServer.Start(wg)
	defer httpServer.Stop()

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt)

	<-osSignals
}
