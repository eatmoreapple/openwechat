# 消息

### 接受消息

被动接受的消息对象，由微信服务器发出

消息对象通过绑定在`bot`上的消息回调函数获取

```go
bot.MessageHandler = func (msg *openwechat.Message) {
if msg.IsText() && msg.Content == "ping" {
msg.ReplyText("pong")
}
}
```

以下简写为`msg`

#### 消息内容

```go
msg.Content // 获取消息内容
```

通过访问`Content`属性可直接获取消息内容

由于消息分为很多种类型，它们都共用`Content`属性。一般当消息类型为文本类型的时候，我们才会去访问`Content`属性。

#### 消息类型判断

下面的判断消息类型的方法均返回`bool`值

##### 文本消息

```go
msg.IsText() 
```

##### 图片消息

```go
msg.IsPicture()
```

##### 位置消息

```go
msg.IsLocation()
```

##### 语音消息

```go
msg.IsVoice()
```

##### 是否为好友添加请求

```go
msg.IsFriendAdd()
```

##### 名片消息

```go
msg.IsCard()
```

##### 视频消息

```go
msg.IsVideo()
```

##### 是否被撤回

```go
msg.IsRecalled()
```

##### 系统消息

```go
msg.IsSystem()
```

##### 收到微信转账

```go
msg.IsTransferAccounts()
```

##### 发出红包(自己发出)

```go
msg.IsSendRedPacket()
```

##### 收到红包

```go
msg.IsReceiveRedPacket()
```

但是不能领取！

##### 判断是否为拍一拍

```go
msg.IsIsPaiYiPai() // 拍一拍消息
msg.IsTickled()
```

##### 判断是否拍了拍自己
```go
msg.IsTickledMe()
```

##### 判断是否有新人加入群聊

```go
msg.IsJoinGroup()
```

#### 获取消息的发送者

```go
sender, err := msg.Sender()
```

如果是群聊消息，该方法返回的是群聊对象(需要自己将`User`转换为`Group`对象)

#### 获取消息的接受者

```go
receiver, err := msg.Receiver()
```

#### 获取消息在群里面的发送者

```go
sender, err := msg.SenderInGroup()
```

获取群聊中具体发消息的用户，前提该消息必须来自群聊。

#### 是否由自己发送

```go
msg.IsSendBySelf()
```

#### 是否为拍一拍

```go
msg.IsTickled()
```

#### 消息是否由好友发出

```go
msg.IsSendByFriend()
```

#### 消息是否由群聊发出

```go
msg.IsSendByGroup()
```

#### 回复文本消息

```go
msg.ReplyText("hello")
```

#### 回复图片消息

```go
img, _ := os.Open("your file path")
defer img.Close()
msg.ReplyImage(img)
```

#### 回复文件消息

```go
file, _ := os.Open("your file path")
defer file.Close()
msg.ReplyFile(file)
```

#### 获取消息里的其他信息

##### 名片消息

```go
card, err := msg.Card()
```

该方法调用的前提为`msg.IsCard()`返回为`true`

名片消息可以获取该名片中的微信号

```go
alias := card.Alias
```

`card`结构

```go
// 名片消息内容
type Card struct {
XMLName                 xml.Name `xml:"msg"`
ImageStatus             int      `xml:"imagestatus,attr"`
Scene                   int      `xml:"scene,attr"`
Sex                     int      `xml:"sex,attr"`
Certflag                int      `xml:"certflag,attr"`
BigHeadImgUrl           string   `xml:"bigheadimgurl,attr"`
SmallHeadImgUrl         string   `xml:"smallheadimgurl,attr"`
UserName                string   `xml:"username,attr"`
NickName                string   `xml:"nickname,attr"`
ShortPy                 string   `xml:"shortpy,attr"`
Alias                   string   `xml:"alias,attr"` // Note: 这个是名片用户的微信号
Province                string   `xml:"province,attr"`
City                    string   `xml:"city,attr"`
Sign                    string   `xml:"sign,attr"`
Certinfo                string   `xml:"certinfo,attr"`
BrandIconUrl            string   `xml:"brandIconUrl,attr"`
BrandHomeUr             string   `xml:"brandHomeUr,attr"`
BrandSubscriptConfigUrl string   `xml:"brandSubscriptConfigUrl,attr"`
BrandFlags              string   `xml:"brandFlags,attr"`
RegionCode              string   `xml:"regionCode,attr"`
}
```

##### 获取已撤回的消息

```go
revokeMsg, err := msg.RevokeMsg()
```

该方法调用成功的前提是`msg.IsRecalled()`返回为`true`

撤回消息的结构

```go
type RevokeMsg struct {
SysMsg    xml.Name `xml:"sysmsg"`
Type      string   `xml:"type,attr"`
RevokeMsg struct {
OldMsgId   int64  `xml:"oldmsgid"`
MsgId      int64  `xml:"msgid"`
Session    string `xml:"session"`
ReplaceMsg string `xml:"replacemsg"`
} `xml:"revokemsg"`
}
```

#### 同意好友请求

```go
friend, err := msg.Agree()
// msg.Agree("我同意了")
```

返回的friend即刚添加的好友对象

该方法调用成功的前提是`msg.IsFriendAdd()`返回为`true`

#### 设置为已读

```go
msg.AsRead()
```

该当前消息设置为已读

#### 设置消息的上下文

用于多个消息处理函数之间的通信，并且是协程安全的。

##### 设置值

```go
msg.Set("hello", "world")
```

##### 获取值

```go
value, exist := msg.Get("hello")
```

### 已发送消息

已发送消息指当前用户发送出去的消息

每次调用发送消息的函数都会返回一个`SentMessage`对象

如

```go
sentMsg, err := msg.ReplyText("hello") // 通过回复消息获取
// sentMsg, err := friend.SendText("hello") // 向好友对象发送消息获取
// and so on
```

#### 撤回消息

撤回刚刚发送的消息，撤回消息的有效时间为2分钟，超过了这个时间则无法撤回

```go
sentMsg.Revoke()
```

#### 判断是否可以撤回

```go
sentMsg.CanRevoke()
```

#### 转发给好友

```go
sentMsg.ForwardToFriends(friend1, friend2)
```

将刚发送的消息转发给好友

#### 转发给群聊

```go
sentMsg.ForwardToGroups(group1, group2)
```

将刚发送的消息转发给群聊

### Emoji表情

openwechat提供了微信全套`emoji`表情的支持

`emoji`表情全部维护在`openwechat.Emoji`结构体上

emoji表情可以通过发送`Text`类型的函数发送

如

```go
firend.SendText(openwechat.Emoji.Doge) // 发送狗头表情
msg.ReplyText(openwechat.Emoji.Awesome) // 发送666的表情
```







