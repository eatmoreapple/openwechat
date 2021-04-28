package openwechat

import (
    "testing"
    "time"
)

func defaultBot(modes ...mode) *Bot {
    bot := DefaultBot(modes...)
    bot.UUIDCallback = PrintlnQrcodeUrl
    return bot
}

func getSelf(modes ...mode) (*Self, error) {
    bot := defaultBot(modes...)
    if err := bot.Login(); err != nil {
        return nil, err
    }
    return bot.GetCurrentUser()
}

func TestBotLogin(t *testing.T) {
    bot := defaultBot()
    if err := bot.Login(); err != nil {
        t.Error(err)
        return
    }
    self, err := bot.GetCurrentUser()
    if err != nil {
        t.Error(err)
        return
    }
    t.Log(self.NickName)
}

func TestMessage(t *testing.T) {
    bot := defaultBot()
    bot.MessageHandler = func(msg *Message) {
        t.Log(msg.MsgType)
        t.Log(msg.Content)
        if msg.IsMedia() {
            t.Log(msg.Content)
            t.Log(msg.FileName)
        }
        if msg.IsCard() {
            c, _ := msg.Card()
            t.Log(c.Alias)
        }
        if msg.IsSystem() {
            t.Log(msg.Content)
        }
        if msg.IsRecalled() {
            t.Log(msg.Content)
        }
    }
    if err := bot.Login(); err != nil {
        t.Error(err)
        return
    }
    bot.Block()
}

func TestFriend(t *testing.T) {
    self, err := getSelf()
    if err != nil {
        t.Error(err)
        return
    }
    friends, err := self.Friends()
    if err != nil {
        t.Error(err)
        return
    }
    t.Log(friends)
}

func TestGroup(t *testing.T) {
    self, err := getSelf()
    if err != nil {
        t.Error(err)
        return
    }
    group, err := self.Groups()
    if err != nil {
        t.Error(err)
        return
    }
    t.Log(group)
    g := group.SearchByNickName(1, "杭州Gopher群组")
    if g.First() != nil {
        members, err := g.First().Members()
        if err != nil {
            t.Error(err)
            return
        }
        t.Log(members.Count())
    }
}

func TestMps(t *testing.T) {
    self, err := getSelf()
    if err != nil {
        t.Error(err)
        return
    }
    mps, err := self.Mps()
    if err != nil {
        t.Error(err)
        return
    }
    t.Log(mps)
}

func TestAddFriendIntoChatRoom(t *testing.T) {
    self, err := getSelf(Desktop)
    if err != nil {
        t.Error(err)
        return
    }
    groups, err := self.Groups()
    if err != nil {
        t.Error(err)
        return
    }
    friends, err := self.Friends()
    if err != nil {
        t.Error(err)
        return
    }
    searchGroups := groups.SearchByNickName(1, "厉害了")
    if g := searchGroups.First(); g != nil {
        addFriends := friends.SearchByRemarkName(1, "1")
        if err := g.AddFriendsIn(addFriends...); err != nil {
            t.Error(err)
        }
    }
}

func TestRemoveFriendIntoChatRoom(t *testing.T) {
    self, err := getSelf()
    if err != nil {
        t.Error(err)
        return
    }
    groups, err := self.Groups()
    if err != nil {
        t.Error(err)
        return
    }
    friends, err := self.Friends()
    if err != nil {
        t.Error(err)
        return
    }
    searchGroups := groups.SearchByNickName(1, "厉害了")
    if g := searchGroups.First(); g != nil {
        addFriends := friends.SearchByRemarkName(1, "大爷")
        if f := addFriends.First(); f != nil {
            if err := g.RemoveMembers(Members{f.User}); err != nil {
                t.Error(err)
            }
        }
    }
}

func TestLogout(t *testing.T) {
    bot := defaultBot()
    bot.MessageHandler = func(msg *Message) {
        if msg.Content == "logout" {
            msg.Bot.Logout()
        }
    }
    if err := bot.Login(); err != nil {
        t.Error(err)
        return
    }
    bot.Block()
}

func TestSendMessage(t *testing.T) {
    bot := defaultBot()
    if err := bot.Login(); err != nil {
        t.Error(err)
        return
    }
    self, err := bot.GetCurrentUser()
    if err != nil {
        t.Error(err)
        return
    }
    helper, err := self.FileHelper()
    if err != nil {
        t.Error(err)
        return
    }
    if err = helper.SendText("test message! received ?"); err != nil {
        t.Error(err)
        return
    }
    time.Sleep(time.Second)
    if err = self.SendTextToFriend(helper, "send test message twice ! received?"); err != nil {
        t.Error(err)
        return
    }
}

func TestAgreeFriendsAdd(t *testing.T) {
    bot := defaultBot()
    bot.MessageHandler = func(msg *Message) {
        if msg.IsFriendAdd() {
            if err := msg.Agree(); err != nil {
                t.Error(err)
            }
            bot.Logout()
        }
    }
    if err := bot.Login(); err != nil {
        t.Error(err)
        return
    }
    bot.Block()
}

func TestHotLogin(t *testing.T) {
    filename := "test.json"
    bot := defaultBot()
    s := NewJsonFileHotReloadStorage(filename)
    if err := bot.HotLogin(s); err != nil {
        t.Error(err)
        return
    }
    self, err := bot.GetCurrentUser()
    if err != nil {
        t.Error(err)
        return
    }
    t.Log(self.NickName)
}

func TestFriendHelper(t *testing.T) {
    bot := defaultBot(Desktop)
    if err := bot.Login(); err != nil {
        t.Error(err)
        return
    }
    self, err := bot.GetCurrentUser()
    if err != nil {
        t.Error(err)
        return
    }
    fh, err := self.FileHelper()
    if err != nil {
        t.Error(err)
        return
    }
    fh.SendText("test message")
}
