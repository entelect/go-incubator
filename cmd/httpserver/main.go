package main

import (
	"fmt"
	"os"
	"os/signal"

	config "go-incubator/internal/configuration"
)

func main() {
	fmt.Println("service started")
	defer fmt.Println("service stopped")

	cfg, err := config.ReadConfig("INCUBATOR_")
	if err != nil {
		fmt.Printf("error reading config: %v\n", err)
		return
	}

	fmt.Printf("starting http listener on port %d\n", cfg.HttpPort)

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt)

	<-osSignals
}
