package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
)

var upgrade = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer func(c *websocket.Conn) {
		err := c.Close()
		if err != nil {
			log.Print(err)
		}
	}(c)
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Print("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Print("write:", err)
			break
		}
	}
}

func SetupLog(verbose bool, logFile string) {
	var Logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.InfoLevel)
	if verbose {
		Logger = Logger.Level(zerolog.DebugLevel).With().Caller().Logger()
	}
	if len(logFile) > 0 {
		fp, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Print(err)
			return
		}
		Logger = Logger.Output(fp)
	}
	log.Logger = Logger
}

func main() {
	var listenHost = flag.String("h", "localhost", "Service listen host")
	var listenPort = flag.Int("p", 8080, "Service listen port")
	var logFile = flag.String("o", "server.log", "Output log file")
	var verbose bool
	flag.BoolVar(&verbose, "v", true, "Log/Show verbose messages")
	flag.Parse()
	SetupLog(verbose, *logFile)

	var listenAddr = fmt.Sprintf("%s:%d", *listenHost, *listenPort)
	http.HandleFunc("/echo", echo)
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Print(err)
		return
	}
}
