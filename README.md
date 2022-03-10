# goTCPTuning
A testing ground for TCP Tuning. Persistent connection and short connection etc.

## websocket长连接

服务端监听端口并提供echo服务，默认为`localhost:8080`

客户端通过websocket连接服务端，一次连接count个，默认100个

给客户端发送interrupt信号，断开全部连接，kill信号强制退出

### 编译

```shell
go build tuning.go
```

### 使用

```shell
goTCPTuning>tuning help
Usage: tuning <flags> <subcommand> <subcommand args>

Subcommands:
        commands         list all command names
        flags            describe all known top-level flags
        help             describe subcommands and their syntax

Subcommands for websocket:
        wsc              Run websocket client.
        wss              Run websocket service.


Use "tuning flags" for a list of top-level flags
```

`-o`设置日志保存文件，`-v`显示调试信息，默认关闭

统计TCP连接信息：`netstat -na | awk '/^tcp/ {++S[$NF]} END {for(a in S) print a, S[a]}'`

### 常用优化

1. dial tcp xxx:8080: socket: too many open files

受客户端进程打开文件数限制，默认1024

```shell
echo "* soft nofile 65535" >> /etc/security/limits.conf
echo "* hard nofile 65535" >> /etc/security/limits.conf
```

重新登录，执行`ulimit -n`可以看到，已经是65535了

服务端可能报错：accept tcp xxx:8080: accept4: too many open files

2. Possible SYN flooding on port 8080. Sending cookies.

客户端发包过快，服务端认为是TCP洪水攻击，0表示关闭，1表示并发高时开启，2表示始终开启

查看：`cat /proc/sys/net/ipv4/tcp_syncookies`

修改：`echo "net.ipv4.tcp_syncookies = 0" >> /etc/sysctl.conf`，然后执行：`sysctl -p`

3. 半连接队列和协议栈队列

`echo 16384 > /proc/sys/net/core/netdev_max_backlog`

`echo 16384 > /proc/sys/net/ipv4/tcp_max_syn_backlog`

`echo 16384 > /proc/sys/net/core/somaxconn`

4. 客户端可用端口

`cat /proc/sys/net/ipv4/ip_local_port_range`，输出为: `32768   60999`，最多可用端口为28232个

修改：`echo "net.ipv4.ip_local_port_range = 1024 65500" >> /etc/sysctl.conf`，然后执行：`sysctl -p`

5. 内存限制

单位: page

`echo "54108   262144   524288" > /proc/sys/net/ipv4/tcp_mem`

单位：字节

`echo "4096    16384   268435456" > /proc/sys/net/ipv4/tcp_wmem`

`echo "4096    87380   268435456" > /proc/sys/net/ipv4/tcp_rmem`

### 交叉编译

```
go env -w CGO_ENABLED=0
go env -w GOOS=linux
go env -w GOARCH=amd64
go build tuning.go
```

GOOS取值：darwin、freebsd、linux、windows

GOARCH取值：386、amd64、arm
