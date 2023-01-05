package openwechat

import (
	"context"
)

// BotOptionFunc 用于设置Bot的选项
type BotOptionFunc func(*Bot)

// Normal 网页版微信
func Normal(b *Bot) {
	b.Caller.Client.SetMode(normal)
}

// Desktop 模式
func Desktop(b *Bot) {
	b.Caller.Client.SetMode(desktop)
}

func WithContext(ctx context.Context) BotOptionFunc {
	return func(b *Bot) {
		b.context = ctx
	}
}

func WithDeviceID(deviceID string) BotOptionFunc {
	return func(b *Bot) {
		b.SetDeviceId(deviceID)
	}
}
