# openwechat

[![Go Doc](https://pkg.go.dev/badge/github.com/eatMoreApple/openwechat)](https://godoc.org/github.com/eatMoreApple/openwechat)
[![Release](https://img.shields.io/github/v/release/eatmoreapple/openwechat.svg?style=flat-square)](https://github.com/eatmoreapple/openwechat/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/eatmoreapple/openwechat)](https://goreportcard.com/badge/github.com/eatmoreapple/openwechat)
[![Stars](https://img.shields.io/github/stars/eatmoreapple/openwechat.svg?style=flat-square)](https://img.shields.io/github/stars/eatmoreapple/openwechat.svg?style=flat-square)
[![Forks](https://img.shields.io/github/forks/eatmoreapple/openwechat.svg?style=flat-square)](https://img.shields.io/github/forks/eatmoreapple/openwechat.svg?style=flat-square)
[![Sourcegraph](https://sourcegraph.com/github.com/eatmoreapple/openwechat/-/badge.svg)](https://sourcegraph.com/github.com/eatmoreapple/openwechat?badge)

> golang版个人微信号API, 突破登录限制，类似开发公众号一样，开发个人微信号



微信机器人:smiling_imp:，利用微信号完成一些功能的定制化开发⭐

* 模块简单易用，易于扩展
* 支持定制化开发，如日志记录，自动回复
* 突破登录限制&#x1F4E3;
* 无需重复扫码登录
* 支持多个微信号同时登陆

### 安装

```shell
go get -u github.com/eatmoreapple/openwechat
```

### 快速开始

```go
package main

import (
	"fmt"
	"github.com/eatmoreapple/openwechat"
)

func main() {
	bot := openwechat.DefaultBot(openwechat.Desktop) // 桌面模式

	// 注册消息处理函数
	bot.MessageHandler = func(msg *openwechat.Message) {
		if msg.IsText() && msg.Content == "ping" {
			msg.ReplyText("pong")
		}
	}
	// 注册登陆二维码回调
	bot.UUIDCallback = openwechat.PrintlnQrcodeUrl

	// 登陆
	if err := bot.Login(); err != nil {
		fmt.Println(err)
		return
	}

	// 获取登陆的用户
	self, err := bot.GetCurrentUser()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 获取所有的好友
	friends, err := self.Friends()
	fmt.Println(friends, err)

	// 获取所有的群组
	groups, err := self.Groups()
	fmt.Println(groups, err)

	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	bot.Block()
}
```

### 支持功能

* 消息回复、给指定对象（好友、群组）发送文本、图片、文件、emoji表情等消息
* 热登陆（无需重复扫码登录）、自定义消息处理、文件下载、消息防撤回
* 获取对象信息、设置好友备注、拉好友进群等
* 更多功能请查看文档

### 文档

[点击查看](https://openwechat.readthedocs.io/zh/latest/)

### 项目主页

[https://github.com/eatmoreapple/openwechat](https://github.com/eatmoreapple/openwechat)

## Thanks

<a href="https://www.jetbrains.com/?from=openwechat"><img src="https://account.jetbrains.com/static/images/jetbrains-logo-inv.svg" height="200" alt="JetBrains"/></a>

### 添加微信(eatmoreapple):apple:（备注: openwechat），进群交流:smiling_imp:

** 如果喜欢的话，可以请作者吃包辣条。谢谢支持 :smile: **

<img width="210px" height="210px" src="https://github.com/eatmoreapple/eatmoreapple/blob/main/img/wechat_pay.jpg" align="left">

<img width="210px" height="210px" src="https://raw.githubusercontent.com/eatmoreapple/eatMoreApple/main/img/wechat.jpg" align="left">





























