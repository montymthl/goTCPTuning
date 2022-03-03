package ws

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"net/http"
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

func (p *ServerCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	setupLog(p.verbose, p.logFile)
	var listenAddr = fmt.Sprintf("%s:%d", p.listenHost, p.listenPort)
	http.HandleFunc("/echo", echo)
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Print(err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

type ServerCmd struct {
	verbose    bool
	listenHost string
	listenPort int
	logFile    string
}

func (*ServerCmd) Name() string     { return "wss" }
func (*ServerCmd) Synopsis() string { return "Run websocket service." }
func (*ServerCmd) Usage() string {
	return `wss [-h host] [-p port] [-o logFile] [-v]:
  Run websocket service.
`
}

func (p *ServerCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.verbose, "v", false, "Log/Show verbose messages")
	f.StringVar(&p.listenHost, "h", "localhost", "Service listen host")
	f.IntVar(&p.listenPort, "p", 8080, "Service listen port")
	f.StringVar(&p.logFile, "o", "server.log", "Log output file")
}
