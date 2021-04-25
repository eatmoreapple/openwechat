## openwechat

> ​	golang版个人微信号API, 类似开发公众号一样，开发个人微信号

### 安装

```shell
go get github.com/eatMoreApple/openwechat
```



### Bot

* `DefaultBot`：`Bot`对象的默认的构造函数

  ```go
  func DefaultBot(modes ...mode) *Bot
  ```

  可通过设置`modes`参数来控制当前的登录行为

  * `Normal`：网页版，如果`modes`参数不传，则为该模式

  * `Desktop`：桌面版，可突破网页版登录限制，网页版模式登录不了的可以尝试该模式

    ```go
    bot := openwechat.DefaultBot(openwechat.Desktop)
    ```

#### 属性

* `Caller`：负责解析与微信服务器交互的响应
* `ScanCallBack`：扫码之后的回调函数, 可以获得扫码用户的头像
* `UUIDCallback`：发起登陆请求后获取uuid的回调函数, 可以通过uuid获取登陆二维码
* `LoginCallBack`：用户确认登陆后的回调函数
* `MessageHandler`：收到消息后的回调函数



#### 方法

* `Login`：发起登陆请求，该方法会一直阻塞，直到用户扫码或者二维码过期

  ```go
  func (b *Bot) Login() error
  ```

* `Logout`：用户退出

  ```go
  func (b *Bot) Logout() error
  ```

* `GetCurrentUser`：获取当前登录的用户（登录后调用）

  ```go
  func (b *Bot) GetCurrentUser() (*Self, error)
  ```

* `Alive`：判断当前登录的用户是否退出

  ```go
  func (b *Bot) Alive() bool
  ```

* `Block`：当消息同步发生了错误或者用户主动在手机上退出，该方法会立即返回，否则会一直阻塞

  ```go
  func (b *Bot) Block() error
  ```



### Self

当前登录的用户对象，调用`Bot.GetCurrentUser()`获取

#### 主要属性

* `Bot`：对应`Bot`对象的指针
* `UserName`：唯一身份标识符(重新登录后改变)
* `NickName`：微信昵称
* `RemarkName`：备注
* `Signature`：签名



#### 主要方法

* `SaveAvatar`：下载头像

  ```go
  func (u *User) SaveAvatar(filename string) error
  ```

* `Members`：获取所有的聊天对象

  ```go
  func (s *Self) Members(update ...bool) (Members, error)
  ```

* `FileHelper`：获取文件传输助手对象

  ```go
  func (s *Self) FileHelper() (*Friend, error)
  ```

* `Friends`：获取所有的好友对象

  ```go
  func (s *Self) Friends(update ...bool) (Friends, error)
  ```

* `Groups`：获取所有的群组对象

  ```go
  func (s *Self) Groups(update ...bool) (Groups, error)
  ```

* `Mps`：获取所有的公众号对象

  ```go
  func (s *Self) Mps(update ...bool) (Mps, error) 
  ```

* `UpdateMembersDetail`：更新所有的联系人详情

  ```go
  func (s *Self) UpdateMembersDetail() error
  ```



### Friend

好友对象

#### 主要属性

* `Self`：当前绑定的登录的用户

* `UserName`：唯一身份标识符(重新登录后改变)
* `NickName`：微信昵称
* `RemarkName`：备注
* `Signature`：签名



#### 主要方法

* `SendMsg`：向其发送消息

  ```go
  func (f *Friend) SendMsg(msg *SendMessage) error
  ```

* `SendText`：向其发送文本消息

  ```go
  func (f *Friend) SendText(content string) error
  ```

* `SendImage`：向其发送图片消息

  ```go
  func (f *Friend) SendImage(file *os.File) error
  ```

* `SetRemarkName`：对其设置备注

  ```go
  func (f *Friend) SetRemarkName(name string) error
  ```

* `SaveAvatar`：下载其头像

  ```go
  func (u *User) SaveAvatar(filename string) error
  ```
  
* `AddIntoGroup`：将其拉入聊天的群组

  ```go
  func (f *Friend) AddIntoGroup(groups ...*Group) error
  ```

  





### Group

群组对象

#### 主要属性

* `Self`：当前绑定的登录的用户

* `UserName`：唯一身份标识符(重新登录后改变)
* `NickName`：微信昵称
* `RemarkName`：备注
* `Signature`：签名



#### 主要方法

* `SendMsg`：向其发送消息

  ```go
  func (f *Group) SendMsg(msg *SendMessage) error
  ```

* `SendText`：向其发送文本消息

  ```go
  func (f *Group) SendText(content string) error
  ```

* `SendImage`：向其发送图片消息

  ```go
  func (f *Group) SendImage(file *os.File) error
  ```

* `SetRemarkName`：对其设置备注

  ```go
  func (f *Group) SetRemarkName(name string) error
  ```

* `SaveAvatar`：下载其头像

  ```go
  func (u *User) SaveAvatar(filename string) error
  ```
  
* `AddFriendsIn`：将好友拉入该群组

  ```go
  func (g *Group) AddFriendsIn(friends ...*Friend) error
  ```
  
* `Members`：获取该群组所有的成员

  ```go
  func (g *Group) Members() (Members, error)
  ```





### Message

可通过绑定在`Bot.MessageHandler`的回调函数获得

* `Sender`：获取消息的发送者

  ```go
  func (m *Message) Sender() (*User, error)
  ```

* `SenderInGroup`：如果是群消息，则可以获取发送消息的群员

  ```go
  func (m *Message) SenderInGroup() (*User, error)
  ```

* `Receiver`：消息的接受者

  ```go
  func (m *Message) Receiver() (*User, error)
  ```

* `IsSendBySelf`：判断当前消息是否由自己发送

  ```go
  func (m *Message) IsSendBySelf() bool
  ```

* `IsSendByFriend`：判断当前消息是否由好友发送

  ```go
  func (m *Message) IsSendByFriend() bool
  ```

* `IsSendByGroup`：判断当前消息是否由群组发送

  ```go
  func (m *Message) IsSendByGroup() bool
  ```

* `Reply`：回复当前消息

  ```go
  func (m *Message) Reply(msgType int, content, mediaId string) error
  ```

* `ReplyText`：回复文本消息

  ```go
  func (m *Message) ReplyText(content string) error
  ```

* `ReplyImage`：回复图片消息

  ```go
  func (m *Message) ReplyImage(file *os.File) error
  ```

* `HasFile`：判断当前消息中是否携带文件

  ```go
  func (m *Message) HasFile() bool
  ```

* `GetFile`：获取文件的响应对象

  ```go
  func (m *Message) GetFile() (*http.Response, error)
  ```

* `Card`：获取当前消息的名片详情（可获取名片中的微信号）

  ```go
  func (m *Message) Card() (*Card, error)
  ```

* `Set`：往当前对象中存入值

  ```go
  func (m *Message) Set(key string, value interface{})
  ```

* `Get`：从当前对象中取出存入的值

  ```go
  func (m *Message) Get(key string) (value interface{}, exist bool)
  ```

  





