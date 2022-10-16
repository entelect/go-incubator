package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"

	config "go-incubator/internal/configuration"
	"go-incubator/internal/http"
	"go-incubator/internal/persistence"
	"go-incubator/internal/persistence/memdb"
	"go-incubator/internal/persistence/mysqldb"
)

func main() {
	fmt.Println("service started")
	defer fmt.Println("service stopped")

	cfg, err := config.ReadConfig("INCUBATOR_")
	if err != nil {
		fmt.Printf("error reading config: %v\n", err)
		return
	}

	var db persistence.Persistence
	switch cfg.Database.DBMS {
	case "inmem":
		fmt.Println("using inmem database")
		imp, err := memdb.NewMemDB()
		if err != nil {
			fmt.Printf("error creating memdb database: %v\n", err)
			return
		}
		db = &imp
	case "mysql":
		fmt.Println("using mysql database")
		imp, err := mysqldb.NewMySqlDB(cfg.Database.ConString)
		if err != nil {
			fmt.Printf("error creating mysql database: %v\n", err)
			return
		}
		db = &imp
	default:
		fmt.Printf("unknown DBMS (%s) specified\n", cfg.Database.DBMS)
		return
	}

	httpServer, err := http.NewHttpServer(cfg.HttpPort, cfg.ApiKey, db)
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
