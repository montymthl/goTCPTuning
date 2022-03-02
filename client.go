package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var instanceCount = flag.Int("count", 100, "client connection count")

var connMap = make(map[string] *websocket.Conn)
var connArr []string

func SetupClientLog(verbose bool, logFile string) {
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

func newConnect(u url.URL) *websocket.Conn {
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Print("dial:", err)
		return nil
	}

	go func() {
		defer func(c *websocket.Conn) {
			addr := c.LocalAddr().String()
			delete(connMap, addr)
			for i := 0; i < len(connArr); i++ {
				if connArr[i] == addr {
					connArr = append(connArr[:i], connArr[i+1:]...)
				}
			}
		}(c)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Print(c.LocalAddr().String() + ", read:", err)
				return
			}
			log.Print("receive: ", string(message))
		}
	}()
	return c
}

func main() {
	flag.Parse()
	SetupClientLog(true, "client.log")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	for i := 0; i < *instanceCount; i++ {
		conn := newConnect(u)
		if conn != nil {
			addr := conn.LocalAddr().String()
			connMap[addr] = conn
			connArr = append(connArr, addr)
		}
	}

	defer func() {
		for i := 0; i < len(connArr); i++ {
			conn := connMap[connArr[i]]
			err := conn.Close()
			if err != nil {
				log.Print(err)
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var count = 0
	rand.Seed(time.Now().UnixNano())

	for {
		select {
		case <-ticker.C:
			count++
			var i = rand.Intn(len(connArr))
			var conn = connMap[connArr[i]]
			err := conn.WriteMessage(websocket.TextMessage, []byte(conn.LocalAddr().String() + ":" + strconv.Itoa(count)))
			if err != nil {
				log.Print("write:", err)
				return
			}
		case <-interrupt:
			log.Print("interrupt")
			for i := 0; i < len(connArr); i++ {
				conn := connMap[connArr[i]]
				err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Print("write close:", err)
					return
				}
			}
			select {
			case <-time.After(time.Second * 3):
			}
			return
		}
	}
}