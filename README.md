# goTCPTuning
A testing ground for TCP Tuning. Persistent connection and short connection etc.

## websocket

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
shell>client -h
Usage of client:
  -addr string
        http service address (default "localhost:8080")
  -count int
        client connection count (default 100)
  -log string
        Log file (default "client.log")
  -v    Log/Show verbose messages (default true)
```