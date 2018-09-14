# 介绍
使用QUIC加上Socks5构建安全上网通道的软件。

# 说明

运行`scripts/gencert.sh`脚本生成证书，
将命令中的localhost替换为域名或者IP地址，
```shell
    ./scripts/gencert.sh localhost
```
命令将生成以下文件
* ca.crt
* ca.key
* ca.srl
* client.crt
* client.csr
* client.key
* server.crt
* server.csr
* server.key

### 服务端
拷贝`ca.crt` `server.crt` `server.key`文件到服务器上，运行下面命令：
```shell
    ./quick server --ca ca.crt --crt server.crt --key server.key
```

### 客户端
运行下面命令：
```shell
    ./quick client --remote localhost:4242 --local 0.0.0.0:1080 --ca ca.crt --crt client.crt --key client.key 
```
如果出现`dial done`客户端连接完成。

### 透明代理
加上参数`--redirect`可以实现透明代理(只支持IPv4)：
```shell
    ./quick client --redirect --remote localhost:4242 --local 0.0.0.0:1080 --ca ca.crt --crt client.crt --key client.key 
```
### 问题
发现QUIC单个对话并发数量过多会出现错误，只能通过创建新的会话解决.

# 致谢

* [golang](https://github.com/golang/go)
* [s5.go](https://github.com/ring04h/s5.go)
* [quic-go](https://github.com/lucas-clemente/quic-go)
* [xjdrew/client.go](https://gist.github.com/xjdrew/97be3811966c8300b724deabc10e38e2)