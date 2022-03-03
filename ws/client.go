package ws

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"github.com/gorilla/websocket"
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

func (p *ClientCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	setupLog(p.verbose, p.logFile)
	var serverAddr = fmt.Sprintf("%s:%d", p.serverHost, p.serverPort)

	interrupt := make(chan os.Signal, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(quit, os.Kill)

	u := url.URL{Scheme: "ws", Host: serverAddr, Path: "/echo"}
	for i := 0; i < p.count; i++ {
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
				return subcommands.ExitFailure
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
					return subcommands.ExitFailure
				}
			}
			select {
			case <-done:
				log.Print("done")
			}
			return subcommands.ExitSuccess
		case <-quit:
			log.Print("killed")
			return subcommands.ExitSuccess
		}
	}
}

type ClientCmd struct {
	verbose    bool
	logFile    string
	serverHost string
	serverPort int
	count      int
}

func (*ClientCmd) Name() string     { return "wsc" }
func (*ClientCmd) Synopsis() string { return "Run websocket client." }
func (*ClientCmd) Usage() string {
	return `wsc [-h host] [-p port] [-o logFile] [-v] -n <count>:
  Run websocket client.
`
}

func (p *ClientCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.verbose, "v", false, "Log/Show verbose messages")
	f.StringVar(&p.serverHost, "h", "localhost", "Server's host")
	f.IntVar(&p.serverPort, "p", 8080, "Server's port")
	f.IntVar(&p.count, "n", 100, "Count of connections")
	f.StringVar(&p.logFile, "o", "client.log", "Log output file")
}
