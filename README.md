# openwechat
> golang版个人微信号API, 类似开发公众号一样，开发个人微信号



[文档](doc.md)



**使用前提**

golang版本大于等于1.11



### 安装

`go get`

```shell
go get github.com/eatMoreApple/openwechat
```





### 快速开始

#### 登录微信

```go
package main

import (
	"fmt"
	"github.com/eatMoreApple/openwechat"
)

func main() {
	messageHandler := func(msg *openwechat.Message) {
		fmt.Println(msg)
	}
	bot := openwechat.DefaultBot()
    
    // 注册消息处理函数
	bot.MessageHandler = messageHandler
    // 设置默认的登录回调
    // 可以设置通过该uuid获取到登录的二维码
	bot.UUIDCallback = openwechat.PrintlnQrcodeUrl
    // 登录
	if err := bot.Login(); err != nil {
		fmt.Println(err)
		return
	}
    // 阻塞主程序,直到用户退出或发生异常
	bot.Block()
}
```



#### 突破网页版登录限制 

网页版登陆不上的可以尝试使用该模式登陆

```go
bot := openwechat.DefaultBot(openwechat.Desktop)
```



#### 回复消息

```go
messageHandler := func(msg *openwechat.Message) {
		msg.ReplyText("hello")
}
```



#### 获取消息的发送者

```go
messageHandler := func(msg *openwechat.Message) {
		sender, err := msg.Sender()
}
```



#### 获取所有的好友

```go
// 登录之后调用
self, err := bot.GetCurrentUser()
if err != nil {
    fmt.Println(err)
    return
}
friends, err := self.Friends()
```



#### 发送消息给好友

```go
self, err := bot.GetCurrentUser()
if err != nil {
    fmt.Println(err)
    return
}
friends, err := self.Friends()
if err != nil {
    fmt.Println(err)
    return
}
if friends.Count() > 0 {
    // 发送给第一个好友
    friends.First().SendText("你好")
}
```



#### 发送图片消息

```go
friends, err := self.Friends()
if err != nil {
    fmt.Println(err)
    return
}
if friends.Count() > 0 {
    // 发送给第一个好友
    img, _ := os.Open("test.png")
    defer img.Close()
    friends.First().SendImage(img)
}
bot.Block()
```



#### 发送Emoji表情

```go
friend.SendText(openwechat.Emoji.Dagger)  
```



#### 搜索好友

```go
friends, err := self.Friends()
if err != nil {
    fmt.Println(err)
    return
}
results := friends.Search(1, func(friend *Friend) bool { return friend.User.RemarkName == "阿青" }, func(friend *Friend) bool { return friend.Sex == 2 }) // 查找数量为1符合条件的好友
fmt.Println(results)
```



#### 搜索群组

```go
groups, err := self.Groups()
if err != nil {
    fmt.Println(err)
    return
}
results := group.Search(1, func(group *Group) bool { return group.NickName == "厉害了" }) 
fmt.Println(results)
```





更多功能请在代码中探索。。。

// todo: add more support 



