## openwechat



### 安装

```shell
go get github.com/eatMoreApple/openwechat
```



### 用户登陆

#### 创建Bot对象

登陆之前需要先创建`Bot`对象来登录

```go
bot := openwechat.DefaultBot()

// 注册消息处理函数
bot.MessageHandler = func(msg *openwechat.Message) {
  if msg.IsText() {
    fmt.Println("你收到了一条新的文本消息")
  }
}
// 注册登陆二维码回调
bot.UUIDCallback = openwechat.PrintlnQrcodeUrl
```



#### 普通登陆

每次运行程序需要重新扫码登录

```go
bot.Login()
```



#### 热登陆

单位时间内运行程序不需要重新扫码登录，直到用户主动退出导致凭证信息失效

```go
storage := openwechat.NewJsonFileHotReloadStorage("storage.json")
bot.HotLogin(storage)
```



#### Desktop模式

`Desktop`可突破部分用户的登录限制，如果普通登陆不上，可尝试使用该模式

```go
bot := openwechat.DefaultBot(openwechat.Desktop)
```



### 消息处理

可通过绑定在`Bot`上的消息回调函数来对消息进行定制化处理

```go
bot := openwechat.DefaultBot(openwechat.Desktop)

messageHandle := func(msg *openwechat.Message) {
  if msg.IsText() {
    fmt.Println("你收到了一条新的文本消息")
  }
}

// 注册消息处理函数
bot.MessageHandler = messageHandle
```



#### 回复文本消息

```go
msg.ReplyText("test message")
```



#### 回复图片消息

```go
file, _ := os.Open("test.png")
defer file.Close()
msg.ReplyImage(file)
```



#### 回复文件消息

```go
file, _ := os.Open("your file name")
defer file.Close()
msg.ReplyFile(file)
```



#### 获取消息的发送者

```go
sender, err := msg.Sender()
```



#### 获取消息的接受者

```go
receiver, err := msg.SenderInGroup()
```



#### 判断消息是否由好友发送

```go
msg.IsSendByFriend() // bool
```



#### 判断消息是否由群组发送

```go
msg.IsSendByGroup() // bool
```



#### 判断消息类型

```go
msg.IsText()             // 是否为文本消息
msg.IsPicture()          // 是否为图片消息
msg.IsVoice()            // 是否为语音消息
msg.IsVideo()            // 是否为视频消息
msg.IsCard()             // 是否为名片消息
msg.IsFriendAdd()        // 是否为添加好友消息
msg.IsRecalled()         // 是否为撤回消息
msg.IsTransferAccounts() // 判断当前的消息是不是微信转账
msg.IsSendRedPacket()    // 是否发出红包
msg.IsReceiveRedPacket() // 判断当前是否收到红包
```



#### 判断消息是否携带文件

```go
msg.HasFile() bool
```



#### 获取消息中的文件

自行读取response处理

```go
resp, err := msg.GetFile() // *http.Response, error
```



#### Card消息

```go
card, err := msg.Card()
if err == nil {
    fmt.Println(card.Alias)   // 获取名片消息中携带的微信号
}
```



#### 同意好友请求

该方法只在消息类型为`IsFriendAdd`为`true`的时候生效

```go
msg.Agree() 

msg.Agree("我同意了你的好友请求")
```



#### Set

从消息上下文中设置值（协成安全）

```go
msg.Set("hello", "world")
```



#### Get

从消息上下文中获取值（协成安全）

```go
value, exist := msg.Get("hello")
```



#### 消息分发

```go
type MessageDispatcher interface {
	Dispatch(msg *Message)
}

func DispatchMessage(dispatcher MessageDispatcher) func(msg *Message) {
	return func(msg *Message) { dispatcher.Dispatch(msg) }
}
```

消息分发处理接口跟 DispatchMessage 结合封装成 MessageHandler



##### MessageMatchDispatcher

> ​	MessageMatchDispatcher impl MessageDispatcher interface

###### example

```go
dispatcher := NewMessageMatchDispatcher()
dispatcher.OnText(func(msg *Message){
	msg.ReplyText("hello")
})
bot := DefaultBot()
bot.MessageHandler = DispatchMessage(dispatcher)
```

###### 注册消息处理函数

```go
dispatcher.RegisterHandler(matchFunc matchFunc, handlers ...MessageContextHandler)
```

`matchFunc`为匹配函数，返回为`true`代表执行对应的`MessageContextHandler`



###### 注册文本消息处理函数

```go
dispatcher.OnText(handlers ...MessageContextHandler)
```



###### 注册图片消息的处理函数

```go
dispatcher.OnImage(handlers ...MessageContextHandler)
```



###### 注册语音消息的处理函数

```go
dispatcher.OnVoice(handlers ...MessageContextHandler)
```





### 登陆用户

登陆成功后调用

```go
self, err := bot.GetCurrentUser()
```



#### 文件传输助手

```go
fileHelper, err := self.FileHelper()
```



#### 好友列表

```go
friends, err := self.Friends()
```



#### 群组列表

注：群组列表只显示手机端微信：通讯录：群聊里面的群组，若想将别的群组加入通讯录，点击群组，设置为`保存到通讯录`即可（安卓机）

```go
groups, err := self.Groups()
```



#### 公众号列表

```go
mps, err := self.Mps()
```



### 好友对象

好友对象通过调用`self.Friends()`获取

```go
friends, err := self.Friends()
```



#### 搜索好友

根据条件查找好友，返回好友列表

```go
friends.SearchByRemarkName(1, "多吃点苹果") // 根据备注查找, limit 参数为限制查找的数量

friends.SearchByNickName(1, "多吃点苹果") // 根据昵称查找

friends.Search(openwechat.ALL, func(friend *openwechat.Friend) bool {
		return friend.Sex == openwechat.MALE
})  // 自定义条件查找(可多个条件)
```



#### 获取第一个好友

返回好友对象

```go
firend := friends.First() // 可能为nil
```



#### 获取最后一个好友

```go
firend := friends.Last() // 可能为nil
```



#### 好友数量统计

```go
count := friends.Count()
```



#### 发送消息

```go
friend := friends.First()
if friend != nil {
		friend.SendText("hello")
  	// SendFile 	发送文件
  	// SendImage	发送图片
}
```



#### 设置备注消息

```go
friend := friends.First()
if friend != nil {
		friend.SetRemarkName("remark name")
}
```



#### 拉入群聊

```go
groups, _ := self.Groups()

friend := friends.First() // ensure it won't be bil

friend.AddIntoGroup(groups...)
```





### 群组对象

好友对象通过调用`self.Groups()`获取

```go
groups, err := self.Groups()
```



#### 发送消息

```go
group := groups.First()
if group != nil {
		group.SendText("hello")
  	// SendFile 	发送文件
  	// SendImage	发送图片
}
```



#### 获取群员列表

```go
group := groups.First()
if group != nil {
  	members, err := group.Members()
}
```



#### 拉好友入群

```go
group := groups.First()  // ensure it won't be bil

group.AddFriendsIn(friend1, friend2)
```



### Emoji

emoji表情可当做一个文本消息发送，具体见`openwechat.Emoji`

```go
friend.SendText(openwechat.Emoji.Doge) // [旺柴]
```



#### 格式化带emoji表情的昵称

```go
fmt.Println(openchat.FormatEmoji(`多吃点苹果<span class="emoji emoji1f34f"></span>`)) 
```



**更多功能请在源码中探索**

```go
// TODO ADD MORE SUPPORT
```









