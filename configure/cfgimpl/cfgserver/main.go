package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/meidoworks/nekoq-component/configure/cfgimpl"
	"github.com/meidoworks/nekoq-component/configure/configserver"
)

var connString string
var listenAddr string

func init() {
	flag.StringVar(&connString, "conn", "postgres://admin:admin@192.168.31.201:15432/configuration", "database connection string")
	flag.StringVar(&listenAddr, "listen", ":8080", "listen address")
}

func main() {
	dataPump := cfgimpl.NewDatabaseDataPump(connString)

	opt := configserver.ConfigureOptions{
		Addr: listenAddr,
		TLSConfig: struct {
			Addr string
			Cert string
			Key  string
		}{},
		MaxWaitTimeForUpdate: 60,
		DataPump:             dataPump,
	}
	server := configserver.NewConfigureServer(opt)
	if err := server.Startup(); err != nil {
		log.Fatal(err)
	}
	defer func(server *configserver.ConfigureServer) {
		err := server.Shutdown()
		if err != nil {
			log.Println("error while shutting down server ", err)
		}
	}(server)

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)
	<-s
	log.Println("Shutting down...")
}
