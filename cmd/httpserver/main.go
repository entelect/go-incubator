package main

import (
	"fmt"
	"os"
	"os/signal"
)

func main() {
	fmt.Println("service started")
	defer fmt.Println("service stopped")

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt)

	<-osSignals
}
