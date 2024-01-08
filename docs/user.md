# 用户

抽象的用户结构: 好友 群组 公众号

```go
type User struct {
	Uin               int
	HideInputBarFlag  int
	StarFriend        int
	Sex               int
	AppAccountFlag    int
	VerifyFlag        int
	ContactFlag       int
	WebWxPluginSwitch int
	HeadImgFlag       int
	SnsFlag           int
	IsOwner           int
	MemberCount       int
	ChatRoomId        int
	UniFriend         int
	OwnerUin          int
	Statues           int
	AttrStatus        int
	Province          string
	City              string
	Alias             string
	DisplayName       string
	KeyWord           string
	EncryChatRoomId   string
	UserName          string
	NickName          string
	HeadImgUrl        string
	RemarkName        string
	PYInitial         string
	PYQuanPin         string
	RemarkPYInitial   string
	RemarkPYQuanPin   string
	Signature         string

	MemberList Members

	Self *Self
}
```

`User`结构体的属性，部分信息可以通过它的英文名知道它所描述的意思。

其中要注意的是`UserName`这个属性。

`UserName`是当前会话唯一的身份标识，且仅作用于当前会话。下次登录该属性值则会被改变。

不同用户的`UserName`的值是不一样的，可以通过该字段来区分不同的用户。


#### 获取用户唯一标识

不同于`UserName`，`ID`是用户的唯一标识，且不会随着登录而改变。

```go
func (u *User) ID() string
```


#### 获取头像

下载群聊、好友、公众号的头像，具体哪种类型根据当前`User`的抽象类型来判断

```go
func (u *User) SaveAvatar(filename string) error
```



#### 详情

获取制定用户的详细信息, 返回新的用户对象

```go
func (u *User) Detail() (*User, error)
```

#### 判断是否为好友

```go
func (u *User) IsFriend() bool
```

#### 判断是否为群组

```go
func (u *User) IsGroup() bool
```

#### 判断是否为公众号

```go
func (u *User) IsMP() bool
```



## 当前登录用户

当前扫码登录的用户对象

`Self`拥有上面`User`的全部属性和方法

通过调用`bot.GetCurrentUser`来获取

```go
self, err := bot.GetCurrentUser()
```



#### 获取当前用户的所有的好友

```go
Friends, err := self.Friends() // self.Friends(true)
```

`Friends`：可接受`bool`值来判断是否获取最新的好友



#### 获取当前用户的所有的群组

```go
groups, err := self.Groups()  // self.Groups(true)
```

注：群组列表只显示手机端微信：通讯录：群聊里面的群组，若想将别的群组加入通讯录，点击群组，设置为`保存到通讯录`即可（安卓机）

如果需要获取不在通讯录里面的群组，则需要收到来自该群组的消息，然后再次调用`self.Groups()`来获取

`Groups`：可接受`bool`值来判断是否获取最新的群组



#### 获取当前用户所有的公众号

```go
mps, err := self.Mps()  // self.Mps(true)
```

`Mps`：可接受`bool`值来判断是否获取最新的公众号



#### 获取文件传输助手

```go
fh, err := self.FileHelper()
```



#### 发送文本给好友

```go
func (s *Self) SendTextToFriend(friend *Friend, text string) (*SentMessage, error)
```

```go
Friends, err := self.Friends()

if err != nil {
    return
}

if friends.Count() > 0 {
    self.SendTextToFriend(friends.First(), "hello")
    // 或者
    // friends.First().SendText("hello")
}
```

返回的`SentMessage`对象可用于消息撤回



#### 发送图片消息给好友

```go
// 确保获取了有效的好友对象
img, _ := os.Open("your file path")
defer img.Close()
self.SendImageToFriend(friend, img)
// 或者
// friend.SendImage(img)
```



#### 发送文件给好友

```go
file, _ := os.Open("your file path")
defer file.Close()
self.SendFileToFriend(friend, file)
// 或者
// friend.SendFile(img)
```





#### 给好友设置备注

```go
self.SetRemarkNameToFriend(friend, "你的备注")
// 或者
// friend.SetRemarkName("你的备注")
```



#### 发送文本消息给群组

```go
self.SendTextToGroup(group, "hello")
// group.SendText("hello")
```



#### 发送图片给群组

```go
img, _ := os.Open("your file path")
defer img.Close()
self.SendImageToGroup(group, img)
// group.SendImage(img)
```



#### 发送文件给群组

```go
file, _ := os.Open("your file path")
defer file.Close()
self.SendFileToGroup(group, file)
// group.SendFile(file)
```



#### 消息撤回

```go
sentMesaage, _ := friend.SendText("hello")
self.RevokeMessage(sentMesaage)
// sentMesaage.Revoke()
```

只要是`openwechat.SentMessage`对象都可以在2分钟之内撤回



#### 消息转发给多个好友

```go
sentMesaage, _ := friend.SendText("hello")

self.ForwardMessageToFriends(sentMesaage, friends1, friends2)

// sentMesaage.ForwardToFriends(friends1, friends2)
```



#### 转发消息给多个群组

```go
sentMesaage, _ := friend.SendText("hello")

self.ForwardMessageToGroups(sentMesaage, group1, group2)

// sentMesaage.ForwardToGroups(friends1, friends2)
```



#### 拉多个好友入群

```go
self.AddFriendsIntoGroup(group, friend1, friend2) // friend1, friend2 为不定长参数

// group.AddFriendsIn(friend1, friend2)
```

最好自己是群主，这样成功率会高一点。



#### 拉单个好友进多个群

```go
self.AddFriendIntoManyGroups(friend, group1, group2) // group1, group2 为不定长参数

// friend.AddIntoGroup(group1, group2)
```



#### 从群聊中移除用户

```go
member, err := group.Members()

self.RemoveMemberFromGroup(group, member[0], member[1])

// group.RemoveMembers(member[:1])
```

注：这个接口已经被微信官方禁用了，现在已经无法使用。



## 好友



### 好友列表

```go
type Friends []*Friend
```

#### 获取当前用户的好友列表

```go
Friends, err := self.Friends()
```

注：此时获取到的`friends`为好友组，而非好友。好友组是当前wx号所有好友的集合。



#### 统计好友个数

```go
Friends.Count() // => int
```



#### 自定义条件查找好友

```go
func (f Friends) Search(limit int, condFuncList ...func(friend *Friend) bool) (results Friends) 
```

* `limit`：限制查找的个数
* `condFuncList`：不定长参数，查找的条件，必须全部满足才算匹配上
* `results`：返回的满足条件的好友组

```go
// 例：查询昵称为eatmoreapple的1个好友
sult := Friends.Search(1, func(friend *openwechat.Friend) bool {return friend.NickName == "eatmoreapple"})
```



#### 根据昵称查找好友

```go
func (f Friends) SearchByNickName(limit int, nickName string) (results Friends)
```

* `limit`：为限制好友查找的个数
* `nickname`：查询指定昵称的好友  
* `results`：返回满足条件的好友组



#### 根据备注查找好友

```go
func (f Friends) SearchByRemarkName(limit int, remarkName string) (results Friends)
```

* `limit`：为限制好友查找的个数
* `remarkname`：查询指定备注的好友
* `results`：返回满足条件的好友组



#### 群发文本消息

```go
func (f Friends) SendText(text string, delay ...time.Duration) error
```

* `text`：文本消息的内容
* `delay`：每次发送消息的间隔（发送消息过快可能会被wx检测到，最好加上间隔时间）



#### 群发图片

```go
func (f Friends) SendImage(file io.Reader, delay ...time.Duration) error 
```

* `file`：`io.Reader`类型。
* `delay`：每次发送消息的间隔（发送消息过快可能会被wx检测到，最好加上间隔时间）



#### 群发文件

```go
func (f Friends) SendFile(file io.Reader, delay ...time.Duration) error
```

* `file`：`io.Reader`类型。
* `delay`：每次发送消息的间隔（发送消息过快可能会被wx检测到，最好加上间隔时间）



### 单个好友



#### 获取好友头像

```go
friend.SaveAvatar("avatar.png")
```



#### 发送文本信息

```go
friend.SendText("hello")
```



#### 发送图片信息

```go
img, _ := os.Open("your image path")

defer img.Close()

friend.SendImage(img)
```



#### 发送文件信息

```go
file, _ := os.Open("your file path")

defer file.Close()

friend.SendFile(file)
```



#### 设置备注信息

```go
friend.SetRemarkName("你的备注")
```



#### 拉该好友进群

```go
friend.AddIntoGroup(group)
```



## 群组

### 群组列表

```go
type Groups []*Group
```



#### 获取所有的群聊

```go
groups, err := self.Groups()
```

注：该方法在用户成功登陆之后调用



#### 统计群聊个数

```go
groups.Count() // => int
```



#### 自定义条件查找群聊

```go
func (g Groups) Search(limit int, condFuncList ...func(group *Group) bool) (results Groups)
```

* `limit`：限制查找的个数
* `condFuncList`：不定长参数，查找的条件，必须全部满足才算匹配上
* `results`：返回的满足条件的群聊



#### 根据群名查找群聊

```go
func (g Groups) SearchByNickName(limit int, nickName string) (results Groups) 
```

* `limit`：限制查找的个数
* `nickName`：群名称
* `results`：返回的满足条件的群聊



#### 群发文本

```go
func (g Groups) SendText(text string, delay ...time.Duration) error 
```

* `text`：文本消息的内容
* `delay`：每次发送消息的间隔（发送消息过快可能会被wx检测到，最好加上间隔时间）



#### 群发图片

```go
func (g Groups) SendImage(file io.Reader, delay ...time.Duration) error
```

* `file`：`io.Reader`类型。
* `delay`：每次发送消息的间隔（发送消息过快可能会被wx检测到，最好加上间隔时间）



#### 群发文件

```go
func (g Groups) SendFile(file io.Reader, delay ...time.Duration) error
```

* `file`：`io.Reader`类型。
* `delay`：每次发送消息的间隔（发送消息过快可能会被wx检测到，最好加上间隔时间）



### 单个群聊



#### 获取群聊头像

```go
group.SaveAvatar("group.png")
```



#### 获取所有的群员

```go
members, err := group.Members()
```



#### 发送文本信息

```go
group.SendText("hello")
```



#### 发送图片信息

```go
img, _ := os.Open("your image path")

defer img.Close()

group.SendImage(img)
```



#### 发送文件消息

```go
file, _ := os.Open("your file path")

defer file.Close()

group.SendFile(file)
```



#### 拉好友进群

```go
group.AddFriendsIn(friend)
```

