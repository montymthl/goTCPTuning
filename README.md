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

统计TCP连接信息：`netstat -na | awk '/^tcp/ {++S[$NF]} END {for(a in S) print a, S[a]}'`

### 常见问题

1. dial tcp xxx:8080: socket: too many open files

受客户端进程打开文件数限制，默认1024

```shell
echo "* soft nofile 65535" >> /etc/security/limits.conf
echo "* hard nofile 65535" >> /etc/security/limits.conf
```

重新登录，执行`ulimit -n`可以看到，已经是65535了

2. accept tcp xxx:8080: accept4: too many open files

同上，受服务端进程打开文件数限制

3. dial:read tcp xxx:60342->xxx:8080: read: connection reset by peer

服务端错误，查看服务端报错信息为：`kernel: TCP: request_sock_TCP: Possible SYN flooding on port 8080. Sending cookies.`

客户端发包过快，服务端认为是TCP洪水攻击，0表示关闭，1表示并发高时开启，2表示始终开启

查看：`cat /proc/sys/net/ipv4/tcp_syncookies`

修改：`echo "net.ipv4.tcp_syncookies = 0" >> /etc/sysctl.conf`，然后执行：`sysctl -p`


4. 客户端连接始终达不到3w

`cat /proc/sys/net/ipv4/ip_local_port_range`，输出为: `32768   60999`，最多可用端口为28232个

修改：`echo "net.ipv4.ip_local_port_range = 1024 65500" >> /etc/sysctl.conf`，然后执行：`sysctl -p`

5. dial tcp xxx:8080: i/o timeout

客户端请求超时

### 交叉编译

```
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build main.go
```

GOOS取值：darwin、freebsd、linux、windows

GOARCH取值：386、amd64、arm