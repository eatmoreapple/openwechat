

# openwechat

[![Go Doc](https://pkg.go.dev/badge/github.com/eatMoreApple/openwechat)](https://godoc.org/github.com/eatMoreApple/openwechat)

> golang版个人微信号API, 类似开发公众号一样，开发个人微信号



微信机器人，利用微信号完成一些功能的定制化开发



**可突破网页版登录限制**	



**使用前提**

golang版本大于等于1.11



### 安装

`go get`

```shell
go get github.com/eatMoreApple/openwechat
```



### 快速开始

```go
package main

import (
	"fmt"
	"github.com/eatMoreApple/openwechat"
)

func main() {
	bot := openwechat.DefaultBot()

	// 注册消息处理函数
	bot.MessageHandler = func(msg *openwechat.Message) {
		if msg.IsText() {
			fmt.Println("你收到了一条新的文本消息")
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

	// 阻塞主goroutine, 知道发生异常或者用户主动退出
	bot.Block()
}
```



### 支持功能

> ​	消息回复、给指定对象（好友、群组）发送文本、图片、文件、emoji表情等消息
>
> ​	热登陆（无需重复扫码登录）、自定义消息处理、文件下载、消息防撤回
>
> ​	获取对象信息、设置好友备注、拉好友进群等



**更多功能请查看文档**



### 文档

[点击查看](doc/doc.md)

### 项目主页

[https://github.com/eatMoreApple/openwechat](https://github.com/eatMoreApple/openwechat)



