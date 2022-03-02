package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

var connMap = make(map[string]*websocket.Conn)
var connArr []string
var lock sync.Mutex
var done = make(chan struct{})

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

func newConnect(u url.URL) {
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Print("dial:", err)
		return
	}

	if c != nil {
		lock.Lock()
		addr := c.LocalAddr().String()
		connMap[addr] = c
		connArr = append(connArr, addr)
		lock.Unlock()
	}
	defer func(c *websocket.Conn) {
		addr := c.LocalAddr().String()
		lock.Lock()
		delete(connMap, addr)
		for i := 0; i < len(connArr); i++ {
			if connArr[i] == addr {
				connArr = append(connArr[:i], connArr[i+1:]...)
			}
		}
		lock.Unlock()
		if len(connArr) == 0 {
			close(done)
		}
	}(c)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Print(c.LocalAddr().String()+", read:", err)
			return
		}
		log.Print("receive: ", string(message))
	}
}

func main() {
	var serverHost = flag.String("h", "localhost", "Server host")
	var serverPort = flag.Int("p", 8080, "Server port")
	var instanceCount = flag.Int("n", 100, "The count of connections")
	var logFile = flag.String("o", "client.log", "Output log file")
	var verbose bool
	flag.BoolVar(&verbose, "v", true, "Log/Show verbose messages")
	flag.Parse()
	SetupClientLog(verbose, *logFile)

	var serverAddr = fmt.Sprintf("%s:%d", *serverHost, *serverPort)
	interrupt := make(chan os.Signal, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(quit, os.Kill)

	u := url.URL{Scheme: "ws", Host: serverAddr, Path: "/echo"}
	for i := 0; i < *instanceCount; i++ {
		go newConnect(u)
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
			if len(connArr) == 0 {
				continue
			}
			count++
			var i = rand.Intn(len(connArr))
			lock.Lock()
			var conn = connMap[connArr[i]]
			lock.Unlock()
			err := conn.WriteMessage(websocket.TextMessage, []byte(conn.LocalAddr().String()+":"+strconv.Itoa(count)))
			if err != nil {
				log.Print("write:", err)
				return
			}
		case <-interrupt:
			log.Printf("interrupted with %d connections", len(connArr))
			var connArrCopied = make([]string, len(connArr))
			var connMapCopied = make(map[string]*websocket.Conn, len(connArr))
			lock.Lock()
			copy(connArrCopied, connArr)
			for k, v := range connMap {
				connMapCopied[k] = v
			}
			lock.Unlock()
			for i := 0; i < len(connArrCopied); i++ {
				conn := connMapCopied[connArrCopied[i]]
				err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, strconv.Itoa(i)))
				if err != nil {
					log.Print("write close:", err)
					return
				}
			}
			select {
			case <-done:
				log.Print("done")
			}
			return
		case <-quit:
			log.Print("killed")
			return
		}
	}
}
