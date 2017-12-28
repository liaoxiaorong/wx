## Weichat(Weixin) Writed By Golang

## 微信客户端 Golang实现

## QuickStart

Download precompiled [Releases](https://github.com/liaoxiaorong/wx/releases)

```console
$ ./wx
```

## Install

```console
$ go get github.com/liaoxiaorong/wx
```

## How to use it?

```console
root@v1:/opt/src/github.com/liaoxiaorong/wx# go run main.go -addr 0.0.0.0:7001
2017/12/26 14:28:37 wx.go:87: Please open link in browser: https://login.weixin.qq.com/qrcode/4buajPnyQQ==
2017/12/26 14:28:49 wx.go:116: scan success, please confirm login on your phone
2017/12/26 14:28:52 wx.go:119: login success
2017/12/26 14:28:56 wx.go:288: update 159 contacts
2017/12/26 14:28:56 web.go:102: web server listen: 0.0.0.0:7001
```

Now visit [localhost:700](http://localhost:7001).
