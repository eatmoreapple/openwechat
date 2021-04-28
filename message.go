package openwechat

import (
    "context"
    "encoding/xml"
    "errors"
    "fmt"
    "net/http"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"
    "unicode"
)

type Message struct {
    IsAt    bool
    AppInfo struct {
        Type  int
        AppID string
    }
    AppMsgType            int
    HasProductId          int
    ImgHeight             int
    ImgStatus             int
    ImgWidth              int
    ForwardFlag           int
    MsgType               int
    Status                int
    StatusNotifyCode      int
    SubMsgType            int
    VoiceLength           int
    CreateTime            int64
    NewMsgId              int64
    PlayLength            int64
    MediaId               string
    MsgId                 string
    EncryFileName         string
    FileName              string
    FileSize              string
    Content               string
    FromUserName          string
    OriContent            string
    StatusNotifyUserName  string
    Ticket                string
    ToUserName            string
    Url                   string
    senderInGroupUserName string
    RecommendInfo         RecommendInfo
    Bot                   *Bot
    mu                    sync.RWMutex
    Context               context.Context
    item                  map[string]interface{}
}

// 获取消息的发送者
func (m *Message) Sender() (*User, error) {
    members, err := m.Bot.self.Members(true)
    if err != nil {
        return nil, err
    }
    if m.FromUserName == m.Bot.self.User.UserName {
        return m.Bot.self.User, nil
    }
    user := members.SearchByUserName(1, m.FromUserName)
    if user == nil {
        return nil, noSuchUserFoundError
    }
    return user.First().Detail()
}

// 获取消息在群里面的发送者
func (m *Message) SenderInGroup() (*User, error) {
    if !m.IsSendByGroup() {
        return nil, errors.New("message is not from group")
    }
    group, err := m.Sender()
    if err != nil {
        return nil, err
    }
    group, err = group.Detail()
    if err != nil {
        return nil, err
    }
    users := group.MemberList.SearchByUserName(1, m.senderInGroupUserName)
    if users == nil {
        return nil, noSuchUserFoundError
    }
    return users.First(), nil
}

// 获取消息的接收者
func (m *Message) Receiver() (*User, error) {
    if m.IsSendByGroup() {
        if sender, err := m.Sender(); err != nil {
            return nil, err
        } else {
            users := sender.MemberList.SearchByUserName(1, m.ToUserName)
            if users == nil {
                return nil, noSuchUserFoundError
            }
            return users.First(), nil
        }
    } else {
        users := m.Bot.self.MemberList.SearchByUserName(1, m.ToUserName)
        if users == nil {
            return nil, noSuchUserFoundError
        }
        return users.First(), nil
    }
}

// 判断消息是否由自己发送
func (m *Message) IsSendBySelf() bool {
    return m.FromUserName == m.Bot.self.User.UserName
}

// 判断消息是否由好友发送
func (m *Message) IsSendByFriend() bool {
    return !m.IsSendByGroup() && strings.HasPrefix(m.FromUserName, "@")
}

// 判断消息是否由群组发送
func (m *Message) IsSendByGroup() bool {
    return strings.HasPrefix(m.FromUserName, "@@")
}

// 回复消息
func (m *Message) Reply(msgType int, content, mediaId string) (*SentMessage, error) {
    msg := NewSendMessage(msgType, content, m.Bot.self.User.UserName, m.FromUserName, mediaId)
    info := m.Bot.storage.LoginInfo
    request := m.Bot.storage.Request
    return m.Bot.Caller.WebWxSendMsg(msg, info, request)
}

// 回复文本消息
func (m *Message) ReplyText(content string) (*SentMessage, error) {
    return m.Reply(TextMessage, content, "")
}

// 回复图片消息
func (m *Message) ReplyImage(file *os.File) (*SentMessage, error) {
    info := m.Bot.storage.LoginInfo
    request := m.Bot.storage.Request
    return m.Bot.Caller.WebWxSendImageMsg(file, request, info, m.Bot.self.UserName, m.FromUserName)
}

func (m *Message) IsText() bool {
    return m.MsgType == 1 && m.Url == ""
}

func (m *Message) IsMap() bool {
    return m.MsgType == 1 && m.Url != ""
}

func (m *Message) IsPicture() bool {
    return m.MsgType == 3 || m.MsgType == 47
}

func (m *Message) IsVoice() bool {
    return m.MsgType == 34
}

func (m *Message) IsFriendAdd() bool {
    return m.MsgType == 37 && m.FromUserName == "fmessage"
}

func (m *Message) IsCard() bool {
    return m.MsgType == 42
}

func (m *Message) IsVideo() bool {
    return m.MsgType == 43 || m.MsgType == 62
}

func (m *Message) IsMedia() bool {
    return m.MsgType == 49
}

// 判断是否撤回
func (m *Message) IsRecalled() bool {
    return m.MsgType == 10002
}

func (m *Message) IsSystem() bool {
    return m.MsgType == 10000
}

func (m *Message) IsNotify() bool {
    return m.MsgType == 51 && m.StatusNotifyCode != 0
}

// 判断当前的消息是不是微信转账
func (m *Message) IsTransferAccounts() bool {
    return m.IsMedia() && m.FileName == "微信转账"
}

// 判断当前是否发出红包
func (m *Message) IsSendRedPacket() bool {
    return m.IsSystem() && m.Content == "发出红包，请在手机上查看"
}

// 判断当前是否收到红包
func (m *Message) IsReceiveRedPacket() bool {
    return m.IsSystem() && m.Content == "收到红包，请在手机上查看"
}

func (m *Message) IsSysNotice() bool {
    return m.MsgType == 9999
}

// 判断是否为操作通知消息
func (m *Message) StatusNotify() bool {
    return m.MsgType == 51
}

// 判断消息是否为文件类型的消息
func (m *Message) HasFile() bool {
    return m.IsPicture() || m.IsVoice() || m.IsVideo() || m.IsMedia()
}

// 获取文件消息的文件
func (m *Message) GetFile() (*http.Response, error) {
    if !m.HasFile() {
        return nil, errors.New("invalid message type")
    }
    if m.IsPicture() {
        return m.Bot.Caller.Client.WebWxGetMsgImg(m, m.Bot.storage.LoginInfo)
    }
    if m.IsVoice() {
        return m.Bot.Caller.Client.WebWxGetVoice(m, m.Bot.storage.LoginInfo)
    }
    if m.IsVideo() {
        return m.Bot.Caller.Client.WebWxGetVideo(m, m.Bot.storage.LoginInfo)
    }
    if m.IsMedia() {
        return m.Bot.Caller.Client.WebWxGetMedia(m, m.Bot.storage.LoginInfo)
    }
    return nil, errors.New("unsupported type")
}

// 获取card类型
func (m *Message) Card() (*Card, error) {
    if !m.IsCard() {
        return nil, errors.New("card message required")
    }
    var card Card
    content := XmlFormString(m.Content)
    err := xml.Unmarshal([]byte(content), &card)
    return &card, err
}

// 获取FriendAddMessageContent内容
func (m *Message) FriendAddMessageContent() (*FriendAddMessage, error) {
    if !m.IsFriendAdd() {
        return nil, errors.New("friend add message required")
    }
    var f FriendAddMessage
    content := XmlFormString(m.Content)
    err := xml.Unmarshal([]byte(content), &f)
    return &f, err
}

// 获取撤回消息的内容
func (m *Message) RevokeMsg() (*RevokeMsg, error) {
    if !m.IsRecalled() {
        return nil, errors.New("recalled message required")
    }
    var r RevokeMsg
    content := XmlFormString(m.Content)
    err := xml.Unmarshal([]byte(content), &r)
    return &r, err
}

// 同意好友的请求
func (m *Message) Agree(verifyContents ...string) error {
    if !m.IsFriendAdd() {
        return fmt.Errorf("friend add message required")
    }
    var builder strings.Builder
    for _, v := range verifyContents {
        builder.WriteString(v)
    }
    return m.Bot.Caller.WebWxVerifyUser(m.Bot.storage, m.RecommendInfo, builder.String())
}

// 往消息上下文中设置值
// goroutine safe
func (m *Message) Set(key string, value interface{}) {
    m.mu.Lock()
    defer m.mu.Unlock()
    if m.item == nil {
        m.item = make(map[string]interface{})
    }
    m.item[key] = value
}

// 从消息上下文中获取值
// goroutine safe
func (m *Message) Get(key string) (value interface{}, exist bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    value, exist = m.item[key]
    return
}

// 消息初始化,根据不同的消息作出不同的处理
func (m *Message) init(bot *Bot) {
    m.Bot = bot
    if m.IsSendByGroup() {
        data := strings.Split(m.Content, ":<br/>")
        m.Content = strings.Join(data[1:], "")
        m.senderInGroupUserName = data[0]
        receiver, err := m.Receiver()
        if err == nil {
            displayName := receiver.DisplayName
            if displayName == "" {
                displayName = receiver.NickName
            }
            atFlag := "@" + displayName
            index := len(atFlag) + 1 + 1
            if strings.HasPrefix(m.Content, atFlag) && unicode.IsSpace(rune(m.Content[index])) {
                m.IsAt = true
                m.Content = m.Content[index+1:]
            }
        }
    }
}

// 发送消息的结构体
type SendMessage struct {
    Type         int
    Content      string
    FromUserName string
    ToUserName   string
    LocalID      string
    ClientMsgId  string
    MediaId      string
}

// SendMessage的构造方法
func NewSendMessage(msgType int, content, fromUserName, toUserName, mediaId string) *SendMessage {
    id := strconv.FormatInt(time.Now().UnixNano()/1e2, 10)
    return &SendMessage{
        Type:         msgType,
        Content:      content,
        FromUserName: fromUserName,
        ToUserName:   toUserName,
        LocalID:      id,
        ClientMsgId:  id,
        MediaId:      mediaId,
    }
}

// 文本消息的构造方法
func NewTextSendMessage(content, fromUserName, toUserName string) *SendMessage {
    return NewSendMessage(TextMessage, content, fromUserName, toUserName, "")
}

// 媒体消息的构造方法
func NewMediaSendMessage(msgType int, fromUserName, toUserName, mediaId string) *SendMessage {
    return NewSendMessage(msgType, "", fromUserName, toUserName, mediaId)
}

// 一些特殊类型的消息会携带该结构体信息
type RecommendInfo struct {
    OpCode     int
    Scene      int
    Sex        int
    VerifyFlag int
    AttrStatus int64
    QQNum      int64
    Alias      string
    City       string
    Content    string
    NickName   string
    Province   string
    Signature  string
    Ticket     string
    UserName   string
}

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

// 好友添加消息信息内容
type FriendAddMessage struct {
    XMLName           xml.Name `xml:"msg"`
    Shortpy           int      `xml:"shortpy,attr"`
    ImageStatus       int      `xml:"imagestatus,attr"`
    Scene             int      `xml:"scene,attr"`
    PerCard           int      `xml:"percard,attr"`
    Sex               int      `xml:"sex,attr"`
    AlbumFlag         int      `xml:"albumflag,attr"`
    AlbumStyle        int      `xml:"albumstyle,attr"`
    SnsFlag           int      `xml:"snsflag,attr"`
    Opcode            int      `xml:"opcode,attr"`
    FromUserName      string   `xml:"fromusername,attr"`
    EncryptUserName   string   `xml:"encryptusername,attr"`
    FromNickName      string   `xml:"fromnickname,attr"`
    Content           string   `xml:"content,attr"`
    Country           string   `xml:"country,attr"`
    Province          string   `xml:"province,attr"`
    City              string   `xml:"city,attr"`
    Sign              string   `xml:"sign,attr"`
    Alias             string   `xml:"alias,attr"`
    WeiBo             string   `xml:"weibo,attr"`
    AlbumBgImgId      string   `xml:"albumbgimgid,attr"`
    SnsBgImgId        string   `xml:"snsbgimgid,attr"`
    SnsBgObjectId     string   `xml:"snsbgobjectid,attr"`
    MHash             string   `xml:"mhash,attr"`
    MFullHash         string   `xml:"mfullhash,attr"`
    BigHeadImgUrl     string   `xml:"bigheadimgurl,attr"`
    SmallHeadImgUrl   string   `xml:"smallheadimgurl,attr"`
    Ticket            string   `xml:"ticket,attr"`
    GoogleContact     string   `xml:"googlecontact,attr"`
    QrTicket          string   `xml:"qrticket,attr"`
    ChatRoomUserName  string   `xml:"chatroomusername,attr"`
    SourceUserName    string   `xml:"sourceusername,attr"`
    ShareCardUserName string   `xml:"sharecardusername,attr"`
    ShareCardNickName string   `xml:"sharecardnickname,attr"`
    CardVersion       string   `xml:"cardversion,attr"`
    BrandList         struct {
        Count int   `xml:"count,attr"`
        Ver   int64 `xml:"ver,attr"`
    } `xml:"brandlist"`
}

// 撤回消息Content
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

// 已发送的信息
type SentMessage struct {
    *SendMessage
    Self  *Self
    MsgId string
}

// 撤回该消息
func (s *SentMessage) Revoke() error {
    return s.Self.RevokeMessage(s)
}
