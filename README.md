# SSL_BOT
 一个用来检测ssl是否过期的程序.
 
## 配置文件
 将config.yaml.bak 改为config.yaml

* `sites` 是一个站点列表.
* `days` 是一个整数，表示在证书过期前多少天进行提示.
* `timeout` 是一个整数，表示 HTTP 请求的超时时间（以秒为单位）.
* `external` 是一个字符串，表示外部程序的路径.
* `method` 是一个字符串，表示调用外部程序的方式（可以是 `pipe` 或 `args`）
  
  外部程序推荐使用[ding_pigeon](https://github.com/ser163/ding_pigeon) 给钉钉群发送消息
* `args` 是一个字符串，表示命令行参数模板（其中 {message} 将被替换为实际的消息内容）。

## 编译
Linux编译
```shell
go build -ldflags "-s -w" -o ssl_bot main.go
```
windows下交叉编译
```shell
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w" -o ssl_bot main.go
```