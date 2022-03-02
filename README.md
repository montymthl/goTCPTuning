# goTCPTuning
A testing ground for TCP Tuning. Persistent connection and short connection etc.

## websocket长连接

服务端监听端口并提供echo服务，默认为`localhost:8080`

客户端通过websocket连接服务端，一次连接count个，默认100个

客户端日志记录在client.log, 服务端日志记录在server.log

给客户端发送interrupt信号，断开全部连接，kill信号强制退出

### 编译

```shell
go build server.go client.go
```

### 使用

```shell
goTCPTuning>client --help
Usage of client:
  -h string
        Server host (default "localhost")
  -n int
        The count of connections (default 100)
  -o string
        Output log file (default "client.log")
  -p int
        Server port (default 8080)
  -v    Log/Show verbose messages (default true)
```

```shell
goTCPTuning>server --help
Usage of server:
  -h string
        Service listen host (default "localhost")
  -o string
        Output log file (default "server.log")
  -p int
        Service listen port (default 8080)
  -v    Log/Show verbose messages (default true)
```