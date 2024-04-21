package openwechat

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Message struct {
	isAt    bool
	AppInfo struct {
		Type  int
		AppID string
	}
	AppMsgType            AppMessageType
	HasProductId          int
	ImgHeight             int
	ImgStatus             int
	ImgWidth              int
	ForwardFlag           int
	MsgType               MessageType
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
	senderUserNameInGroup string
	RecommendInfo         RecommendInfo
	bot                   *Bot
	mu                    sync.RWMutex
	context               context.Context
	item                  map[string]interface{}
	Raw                   []byte `json:"-"`
	RawContent            string `json:"-"` // 消息原始内容
}

// Sender 获取消息的发送者
func (m *Message) Sender() (*User, error) {
	if m.IsSendBySelf() {
		return m.Owner().User, nil
	}
	// 首先尝试从缓存里面查找, 如果没有找到则从服务器获取
	members, err := m.bot.self.Members()
	if err != nil {
		return nil, err
	}
	user, exist := members.GetByUserName(m.FromUserName)
	if !exist {
		// 找不到, 从服务器获取
		user = newFriend(m.FromUserName, m.Owner()).User
		err = user.Detail()
	}
	if m.IsSendByGroup() && len(user.MemberList) == 0 {
		err = user.Detail()
	}
	return user, err
}

// SenderInGroup 获取消息在群里面的发送者
func (m *Message) SenderInGroup() (*User, error) {
	if !m.IsComeFromGroup() {
		return nil, errors.New("message is not from group")
	}
	// 拍一拍系列的系统消息
	// https://github.com/eatmoreapple/openwechat/issues/66
	if m.IsSystem() {
		// 判断是否有自己发送
		if m.IsSendBySelf() {
			return m.Owner().User, nil
		}
		return nil, errors.New("can not found sender from system message")
	}
	user, err := m.Sender()
	if err != nil {
		return nil, err
	}
	if user.IsFriend() {
		return user, nil
	}
	group := &Group{user}
	return group.SearchMemberByUsername(m.senderUserNameInGroup)
}

// Receiver 获取消息的接收者
// 如果消息是群组消息，则返回群组
// 如果消息是好友消息，则返回好友
// 如果消息是系统消息，则返回当前用户
func (m *Message) Receiver() (*User, error) {
	if m.IsSystem() || m.ToUserName == m.bot.self.UserName {
		return m.bot.self.User, nil
	}
	// https://github.com/eatmoreapple/openwechat/issues/113
	if m.ToUserName == FileHelper {
		return m.Owner().FileHelper().User, nil
	}

	if m.IsSendByGroup() {
		groups, err := m.Owner().Groups()
		if err != nil {
			return nil, err
		}
		username := m.FromUserName
		if m.IsSendBySelf() {
			username = m.ToUserName
		}
		users := groups.SearchByUserName(1, username)
		if users.Count() == 0 {
			group := newUser(m.Owner(), username)
			if err := group.Detail(); err == nil {
				return group, nil
			}
			return nil, ErrNoSuchUserFoundError
		}
		return users.First().User, nil
	} else {
		members, err := m.Owner().Members()
		if err != nil {
			return nil, err
		}
		user, exist := members.GetByUserName(m.ToUserName)
		if !exist {
			return nil, ErrNoSuchUserFoundError
		}
		return user, nil
	}
}

// IsSendBySelf 判断消息是否由自己发送
func (m *Message) IsSendBySelf() bool {
	return m.FromUserName == m.Owner().UserName
}

// IsSendByFriend 判断消息是否由好友发送
func (m *Message) IsSendByFriend() bool {
	return !m.IsSendByGroup() && strings.HasPrefix(m.FromUserName, "@") && !m.IsSendBySelf()
}

// IsSendByGroup 判断消息是否由群组发送
func (m *Message) IsSendByGroup() bool {
	return strings.HasPrefix(m.FromUserName, "@@") || (m.IsSendBySelf() && strings.HasPrefix(m.ToUserName, "@@"))
}

// IsSelfSendToGroup 判断消息是否由自己发送到群组
func (m *Message) IsSelfSendToGroup() bool {
	return m.IsSendBySelf() && strings.HasPrefix(m.ToUserName, "@@")
}

// ReplyText 回复文本消息
func (m *Message) ReplyText(content string) (*SentMessage, error) {
	// 判断是否由自己发送
	username := m.FromUserName
	if m.IsSelfSendToGroup() {
		username = m.ToUserName
	}
	return m.Owner().sendTextToUser(username, content)
}

// ReplyImage 回复图片消息
func (m *Message) ReplyImage(file io.Reader) (*SentMessage, error) {
	// 判断是否由自己发送
	username := m.FromUserName
	if m.IsSelfSendToGroup() {
		username = m.ToUserName
	}
	return m.Owner().sendImageToUser(username, file)
}

// ReplyVideo 回复视频消息
func (m *Message) ReplyVideo(file io.Reader) (*SentMessage, error) {
	// 判断是否由自己发送
	username := m.FromUserName
	if m.IsSelfSendToGroup() {
		username = m.ToUserName
	}
	return m.Owner().sendVideoToUser(username, file)
}

// ReplyFile 回复文件消息
func (m *Message) ReplyFile(file io.Reader) (*SentMessage, error) {
	// 判断是否由自己发送
	username := m.FromUserName
	if m.IsSelfSendToGroup() {
		username = m.ToUserName
	}
	return m.Owner().sendFileToUser(username, file)
}

func (m *Message) IsText() bool {
	return m.MsgType == MsgTypeText && m.Url == ""
}

func (m *Message) IsLocation() bool {
	return m.MsgType == MsgTypeText && strings.Contains(m.Url, "apis.map.qq.com") && strings.Contains(m.Content, "pictype=location")
}

func (m *Message) IsRealtimeLocation() bool {
	return m.IsRealtimeLocationStart() || m.IsRealtimeLocationStop()
}

func (m *Message) IsRealtimeLocationStart() bool {
	return m.MsgType == MsgTypeApp && m.AppMsgType == AppMsgTypeRealtimeShareLocation
}

func (m *Message) IsRealtimeLocationStop() bool {
	return m.MsgType == MsgTypeSys && m.Content == "位置共享已经结束"
}

func (m *Message) IsPicture() bool {
	return m.MsgType == MsgTypeImage
}

// IsEmoticon 是否为表情包消息
func (m *Message) IsEmoticon() bool {
	return m.MsgType == MsgTypeEmoticon
}

func (m *Message) IsVoice() bool {
	return m.MsgType == MsgTypeVoice
}

func (m *Message) IsFriendAdd() bool {
	return m.MsgType == MsgTypeVerify && m.FromUserName == "fmessage"
}

func (m *Message) IsCard() bool {
	return m.MsgType == MsgTypeShareCard
}

func (m *Message) IsVideo() bool {
	return m.MsgType == MsgTypeVideo || m.MsgType == MsgTypeMicroVideo
}

func (m *Message) IsMedia() bool {
	return m.MsgType == MsgTypeApp
}

// IsRecalled 判断是否撤回
func (m *Message) IsRecalled() bool {
	return m.MsgType == MsgTypeRecalled
}

func (m *Message) IsSystem() bool {
	return m.MsgType == MsgTypeSys
}

func (m *Message) IsNotify() bool {
	return m.MsgType == 51 && m.StatusNotifyCode != 0
}

// IsTransferAccounts 判断当前的消息是不是微信转账
func (m *Message) IsTransferAccounts() bool {
	return m.IsMedia() && m.FileName == "微信转账"
}

// IsSendRedPacket 否发出红包判断当前是
func (m *Message) IsSendRedPacket() bool {
	return m.IsSystem() && m.Content == "发出红包，请在手机上查看"
}

// IsReceiveRedPacket 判断当前是否收到红包
func (m *Message) IsReceiveRedPacket() bool {
	return m.IsSystem() && m.Content == "收到红包，请在手机上查看"
}

// IsRenameGroup 判断当前是否是群组重命名
func (m *Message) IsRenameGroup() bool {
	return m.IsSystem() && strings.Contains(m.Content, "修改群名为")
}

func (m *Message) IsSysNotice() bool {
	return m.MsgType == 9999
}

// StatusNotify 判断是否为操作通知消息
func (m *Message) StatusNotify() bool {
	return m.MsgType == 51
}

// HasFile 判断消息是否为文件类型的消息
func (m *Message) HasFile() bool {
	return m.IsPicture() || m.IsVoice() || m.IsVideo() || m.HasAttachment() || m.IsEmoticon()
}

// HasAttachment 是否有附件
func (m *Message) HasAttachment() bool {
	return m.IsMedia() && m.AppMsgType == AppMsgTypeAttach
}

// GetFile 获取文件消息的文件
func (m *Message) GetFile() (*http.Response, error) {
	if !m.HasFile() {
		return nil, errors.New("invalid message type")
	}
	switch {
	case m.IsPicture() || m.IsEmoticon():
		return m.bot.Caller.Client.WebWxGetMsgImg(m.Context(), m, m.bot.Storage.LoginInfo)
	case m.IsVoice():
		return m.bot.Caller.Client.WebWxGetVoice(m.Context(), m, m.bot.Storage.LoginInfo)
	case m.IsVideo():
		return m.bot.Caller.Client.WebWxGetVideo(m.Context(), m, m.bot.Storage.LoginInfo)
	case m.IsMedia() && m.AppMsgType == AppMsgTypeAttach:
		return m.bot.Caller.Client.WebWxGetMedia(m.Context(), m, m.bot.Storage.LoginInfo)
	default:
		return nil, errors.New("unsupported type")
	}
}

// GetPicture 获取图片消息的响应
func (m *Message) GetPicture() (*http.Response, error) {
	if !(m.IsPicture() || m.IsEmoticon()) {
		return nil, errors.New("picture message required")
	}
	return m.bot.Caller.Client.WebWxGetMsgImg(m.Context(), m, m.bot.Storage.LoginInfo)
}

// GetVoice 获取录音消息的响应
func (m *Message) GetVoice() (*http.Response, error) {
	if !m.IsVoice() {
		return nil, errors.New("voice message required")
	}
	return m.bot.Caller.Client.WebWxGetVoice(m.Context(), m, m.bot.Storage.LoginInfo)
}

// GetVideo 获取视频消息的响应
func (m *Message) GetVideo() (*http.Response, error) {
	if !m.IsVideo() {
		return nil, errors.New("video message required")
	}
	return m.bot.Caller.Client.WebWxGetVideo(m.Context(), m, m.bot.Storage.LoginInfo)
}

// GetMedia 获取媒体消息的响应
func (m *Message) GetMedia() (*http.Response, error) {
	if !m.IsMedia() {
		return nil, errors.New("media message required")
	}
	return m.bot.Caller.Client.WebWxGetMedia(m.Context(), m, m.bot.Storage.LoginInfo)
}

// SaveFile 保存文件到指定的 io.Writer
func (m *Message) SaveFile(writer io.Writer) error {
	resp, err := m.GetFile()
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, err = io.Copy(writer, resp.Body)
	return err
}

// SaveFileToLocal 保存文件到本地
func (m *Message) SaveFileToLocal(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	return m.SaveFile(file)
}

// Card 获取card类型
func (m *Message) Card() (*Card, error) {
	if !m.IsCard() {
		return nil, errors.New("card message required")
	}
	var card Card
	err := xml.Unmarshal(stringToByte(m.Content), &card)
	return &card, err
}

// FriendAddMessageContent 获取FriendAddMessageContent内容
func (m *Message) FriendAddMessageContent() (*FriendAddMessage, error) {
	if !m.IsFriendAdd() {
		return nil, errors.New("friend add message required")
	}
	var f FriendAddMessage
	err := xml.Unmarshal(stringToByte(m.Content), &f)
	return &f, err
}

// RevokeMsg 获取撤回消息的内容
func (m *Message) RevokeMsg() (*RevokeMsg, error) {
	if !m.IsRecalled() {
		return nil, errors.New("recalled message required")
	}
	var r RevokeMsg
	err := xml.Unmarshal(stringToByte(m.Content), &r)
	return &r, err
}

// Agree 同意好友的请求
func (m *Message) Agree(verifyContents ...string) (*Friend, error) {
	if !m.IsFriendAdd() {
		return nil, errors.New("friend add message required")
	}
	opt := &CallerWebWxVerifyUserOptions{
		VerifyContent: strings.Join(verifyContents, ""),
		RecommendInfo: m.RecommendInfo,
		BaseRequest:   m.bot.Storage.Request,
		LoginInfo:     m.bot.Storage.LoginInfo,
	}
	err := m.bot.Caller.WebWxVerifyUser(m.Context(), opt)
	if err != nil {
		return nil, err
	}
	friend := newFriend(m.RecommendInfo.UserName, m.Owner())
	if err = friend.Detail(); err != nil {
		return nil, err
	}
	return friend, nil
}

// AsRead 将消息设置为已读
func (m *Message) AsRead() error {
	opt := &CallerWebWxStatusAsReadOptions{
		BaseRequest: m.bot.Storage.Request,
		LoginInfo:   m.bot.Storage.LoginInfo,
		Message:     m,
	}
	return m.bot.Caller.WebWxStatusAsRead(m.Context(), opt)
}

// IsArticle 判断当前的消息类型是否为文章
func (m *Message) IsArticle() bool {
	return m.AppMsgType == AppMsgTypeUrl
}

// MediaData 获取当前App Message的具体内容
func (m *Message) MediaData() (*AppMessageData, error) {
	if !m.IsMedia() {
		return nil, errors.New("media message required")
	}
	var data AppMessageData
	if err := xml.Unmarshal(stringToByte(m.Content), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// Set 往消息上下文中设置值
// goroutine safe
func (m *Message) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.item == nil {
		m.item = make(map[string]interface{})
	}
	m.item[key] = value
}

// Get 从消息上下文中获取值
// goroutine safe
func (m *Message) Get(key string) (value interface{}, exist bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exist = m.item[key]
	return
}

// 消息初始化,根据不同的消息作出不同的处理
func (m *Message) init(bot *Bot) {
	m.bot = bot
	defaultMessageObserver.OnMessageReceive(m)
}

// SendMessage 发送消息的结构体
type SendMessage struct {
	Type         MessageType
	Content      string
	FromUserName string
	ToUserName   string
	LocalID      string
	ClientMsgId  string
	MediaId      string `json:"MediaId,omitempty"`
}

// NewSendMessage SendMessage的构造方法
func NewSendMessage(msgType MessageType, content, fromUserName, toUserName, mediaId string) *SendMessage {
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

// NewTextSendMessage 文本消息的构造方法
func NewTextSendMessage(content, fromUserName, toUserName string) *SendMessage {
	return NewSendMessage(MsgTypeText, content, fromUserName, toUserName, "")
}

// NewMediaSendMessage 媒体消息的构造方法
func NewMediaSendMessage(msgType MessageType, fromUserName, toUserName, mediaId string) *SendMessage {
	return NewSendMessage(msgType, "", fromUserName, toUserName, mediaId)
}

// RecommendInfo 一些特殊类型的消息会携带该结构体信息
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

// Card 名片消息内容
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

// FriendAddMessage 好友添加消息信息内容
type FriendAddMessage struct {
	XMLName           xml.Name `xml:"msg"`
	Shortpy           string   `xml:"shortpy,attr"`
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

// RevokeMsg 撤回消息Content
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

// SentMessage 已发送的信息
type SentMessage struct {
	*SendMessage
	self  *Self
	MsgId string
}

// Revoke 撤回该消息
func (s *SentMessage) Revoke() error {
	return s.self.RevokeMessage(s)
}

// CanRevoke 是否可以撤回该消息
func (s *SentMessage) CanRevoke() bool {
	i, err := strconv.ParseInt(s.ClientMsgId, 10, 64)
	if err != nil {
		return false
	}
	start := time.Unix(i/10000000, 0)
	return time.Since(start) < 2*time.Minute
}

// ForwardToFriends 转发该消息给好友
// 该方法会阻塞直到所有好友都接收到消息
// 这里为了兼容以前的版本，默认休眠0.5秒，如果需要更快的速度，可以使用 SentMessage.ForwardToFriendsWithDelay
func (s *SentMessage) ForwardToFriends(friends ...*Friend) error {
	return s.ForwardToFriendsWithDelay(time.Second/2, friends...)
}

// ForwardToFriendsWithDelay 转发该消息给好友，延迟指定时间
func (s *SentMessage) ForwardToFriendsWithDelay(delay time.Duration, friends ...*Friend) error {
	return s.self.ForwardMessageToFriends(s, delay, friends...)
}

// ForwardToGroups 转发该消息给群组
// 该方法会阻塞直到所有群组都接收到消息
// 这里为了兼容以前的版本，默认休眠0.5秒，如果需要更快的速度，可以使用 SentMessage.ForwardToGroupsDelay
func (s *SentMessage) ForwardToGroups(groups ...*Group) error {
	return s.ForwardToGroupsWithDelay(time.Second/2, groups...)
}

// ForwardToGroupsWithDelay 转发该消息给群组， 延迟指定时间
func (s *SentMessage) ForwardToGroupsWithDelay(delay time.Duration, groups ...*Group) error {
	return s.self.ForwardMessageToGroups(s, delay, groups...)
}

type appmsg struct {
	Type      int    `xml:"type"`
	AppId     string `xml:"appid,attr"` // wxeb7ec651dd0aefa9
	SdkVer    string `xml:"sdkver,attr"`
	Title     string `xml:"title"`
	Des       string `xml:"des"`
	Action    string `xml:"action"`
	Content   string `xml:"content"`
	Url       string `xml:"url"`
	LowUrl    string `xml:"lowurl"`
	ExtInfo   string `xml:"extinfo"`
	AppAttach struct {
		TotalLen int64  `xml:"totallen"`
		AttachId string `xml:"attachid"`
		FileExt  string `xml:"fileext"`
	} `xml:"appattach"`
}

func (f appmsg) XmlByte() ([]byte, error) {
	return xml.Marshal(f)
}

func NewFileAppMessage(stat os.FileInfo, attachId string) *appmsg {
	m := &appmsg{AppId: appMessageAppId, Title: stat.Name()}
	m.AppAttach.AttachId = attachId
	m.AppAttach.TotalLen = stat.Size()
	m.Type = 6
	m.AppAttach.FileExt = getFileExt(stat.Name())
	return m
}

// AppMessageData 获取APP消息的正文
// See https://github.com/eatmoreapple/openwechat/issues/62
type AppMessageData struct {
	XMLName xml.Name `xml:"msg"`
	AppMsg  struct {
		Appid             string         `xml:"appid,attr"`
		SdkVer            string         `xml:"sdkver,attr"`
		Title             string         `xml:"title"`
		Des               string         `xml:"des"`
		Action            string         `xml:"action"`
		Type              AppMessageType `xml:"type"`
		ShowType          string         `xml:"showtype"`
		Content           string         `xml:"content"`
		URL               string         `xml:"url"`
		DataUrl           string         `xml:"dataurl"`
		LowUrl            string         `xml:"lowurl"`
		LowDataUrl        string         `xml:"lowdataurl"`
		RecordItem        string         `xml:"recorditem"`
		ThumbUrl          string         `xml:"thumburl"`
		MessageAction     string         `xml:"messageaction"`
		Md5               string         `xml:"md5"`
		ExtInfo           string         `xml:"extinfo"`
		SourceUsername    string         `xml:"sourceusername"`
		SourceDisplayName string         `xml:"sourcedisplayname"`
		CommentUrl        string         `xml:"commenturl"`
		AppAttach         struct {
			TotalLen          string `xml:"totallen"`
			AttachId          string `xml:"attachid"`
			EmoticonMd5       string `xml:"emoticonmd5"`
			FileExt           string `xml:"fileext"`
			FileUploadToken   string `xml:"fileuploadtoken"`
			OverwriteNewMsgId string `xml:"overwrite_newmsgid"`
			FileKey           string `xml:"filekey"`
			CdnAttachUrl      string `xml:"cdnattachurl"`
			AesKey            string `xml:"aeskey"`
			EncryVer          string `xml:"encryver"`
		} `xml:"appattach"`
		WeAppInfo struct {
			PagePath       string `xml:"pagepath"`
			Username       string `xml:"username"`
			Appid          string `xml:"appid"`
			AppServiceType string `xml:"appservicetype"`
		} `xml:"weappinfo"`
		WebSearch string `xml:"websearch"`
	} `xml:"appmsg"`
	FromUsername string `xml:"fromusername"`
	Scene        string `xml:"scene"`
	AppInfo      struct {
		Version string `xml:"version"`
		AppName string `xml:"appname"`
	} `xml:"appinfo"`
	CommentUrl string `xml:"commenturl"`
}

// IsFromApplet 判断当前的消息类型是否来自小程序
func (a *AppMessageData) IsFromApplet() bool {
	return a.AppMsg.Appid != ""
}

// IsArticle 判断当前的消息类型是否为文章
func (a *AppMessageData) IsArticle() bool {
	return a.AppMsg.Type == AppMsgTypeUrl
}

// IsFile 判断当前的消息类型是否为文件
func (a AppMessageData) IsFile() bool {
	return a.AppMsg.Type == AppMsgTypeAttach
}

// IsComeFromGroup 判断消息是否来自群组
// 可能是自己或者别的群员发送
func (m *Message) IsComeFromGroup() bool {
	return m.IsSendByGroup() || (strings.HasPrefix(m.ToUserName, "@@") && m.IsSendBySelf())
}

func (m *Message) String() string {
	return fmt.Sprintf("<%s:%s>", m.MsgType, m.MsgId)
}

// IsAt 判断消息是否为@消息
func (m *Message) IsAt() bool {
	return m.isAt
}

// IsPaiYiPai 判断消息是否为拍一拍
// 不要问我为什么取名为PaiYiPai，因为我也不知道取啥名字好
func (m *Message) IsPaiYiPai() bool {
	return m.IsTickled()
}

// IsJoinGroup 判断是否有人加入了群聊
func (m *Message) IsJoinGroup() bool {
	return m.IsSystem() && (strings.Contains(m.Content, "加入了群聊") || strings.Contains(m.Content, "分享的二维码加入群聊")) && m.IsSendByGroup()
}

// IsTickled 判断消息是否为拍一拍
func (m *Message) IsTickled() bool {
	return m.IsSystem() && (strings.Contains(m.Content, "拍了拍") || strings.Contains(m.Content, "拍拍"))
}

// IsTickledMe 判断消息是否拍了拍自己
func (m *Message) IsTickledMe() bool {
	return m.IsSystem() && (strings.Count(m.Content, "拍了拍我") == 1 || strings.Count(m.Content, "拍拍我") == 1)
}

// IsVoipInvite 判断消息是否为语音或视频通话邀请
func (m *Message) IsVoipInvite() bool {
	return m.MsgType == MsgTypeVoipInvite
}

// Bot 返回当前消息所属的Bot
func (m *Message) Bot() *Bot {
	return m.bot
}

// Owner 返回当前消息的拥有者
func (m *Message) Owner() *Self {
	return m.Bot().self
}

func (m *Message) Context() context.Context {
	if m.context == nil {
		return m.Bot().Context()
	}
	return m.context
}

func (m *Message) WithContext(ctx context.Context) {
	if ctx == nil {
		panic("nil context")
	}
	m.context = ctx
}
