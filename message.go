package xmpush

import (
	"fmt"
	"strings"
	"time"
)

var (
	MaxTimeToSend = time.Hour * 24 * 7
	MaxTimeToLive = int64(3600 * 1000 * 24 * 7 * 2)
)

func NewMessage(title string, description string) *Message {
	return &Message{
		Payload:     "",
		Title:       title,
		Description: description,
		PassThrough: 0,
		NotifyType:  1,
		Extra:       make(map[string]string),
	}
}

type Message struct {
	RestrictedPackageName string            `json:"restricted_package_name,omitempty"`
	Payload               string            `json:"payload,omitempty"`
	Title                 string            `json:"title"`
	Description           string            `json:"description"`
	PassThrough           int32             `json:"pass_through"`          // 0 通知栏消息, 1 透传消息
	NotifyType            int32             `json:"notify_type,omitempty"` // -1: DEFAULT_ALL 1: 使用默认提示音提示, 2: 使用默认震动提示, 4: 使用默认led灯光提示
	TimeToLive            int64             `json:"time_to_live,omitempty"`
	TimeToSend            int64             `json:"time_to_send,omitempty"`
	NotifyID              int64             `json:"notify_id,omitempty"`
	Extra                 map[string]string `json:"extra,omitempty"`
}

func (m *Message) SetRestrictedPackageName(packageName ...string) *Message {
	m.RestrictedPackageName = strings.Join(packageName, ",")
	return m
}

func (m *Message) SetPayload(payload string) *Message {
	m.Payload = payload
	return m
}

func (m *Message) SetTitle(title string) *Message {
	m.Title = title
	return m
}

func (m *Message) SetDescription(description string) *Message {
	m.Description = description
	return m
}

func (m *Message) EnablePassThrough() *Message {
	m.PassThrough = 1
	return m
}

func (m *Message) DisablePassThrough() *Message {
	m.PassThrough = 0
	return m
}

func (m *Message) SetNotifyType(notifyType int32) *Message {
	m.NotifyType = notifyType
	return m
}

func (m *Message) SetTimeToLive(ttl int64) *Message {
	if ttl <= 0 || ttl > MaxTimeToLive {
		m.TimeToLive = MaxTimeToLive
	} else {
		m.TimeToLive = ttl
	}
	return m
}

func (m *Message) SetTimeToSend(timeToSend int64) *Message {
	sc := time.Unix(0, timeToSend*int64(time.Millisecond))
	max := time.Now().Add(MaxTimeToSend)
	if sc.After(max) {
		m.TimeToSend = max.UnixNano() / int64(time.Millisecond)
	} else if sc.Before(time.Now()) {
		m.TimeToSend = 0
	} else {
		m.TimeToSend = timeToSend
	}
	return m
}

func (m *Message) SetNotifyID(notifyID int64) *Message {
	m.NotifyID = notifyID
	return m
}

func (m *Message) AddExtra(key string, value string) *Message {
	if m.Extra == nil {
		m.Extra = make(map[string]string)
	}
	m.Extra[key] = value
	return m
}

func (m *Message) RemoveExtra(key string) *Message {
	if m.Extra != nil {
		delete(m.Extra, key)
	}
	return m
}

func (m *Message) EnableFlowControl() *Message {
	m.AddExtra("flow_control", "1")
	return m
}

func (m *Message) DisableFlowControl() *Message {
	m.RemoveExtra("flow_control")
	return m
}

func (m *Message) SetBadge(badge int64) *Message {
	m.AddExtra("badge", fmt.Sprintf("%d", badge))
	return m
}

func (m *Message) SetJobKey(jobKey string) *Message {
	m.AddExtra("jobkey", jobKey)
	return m
}

func (m *Message) SetCallback(callbackURL string) *Message {
	m.AddExtra("callback", callbackURL)
	// 1:送达回执, 2:点击回执, 3:送达和点击回执
	m.AddExtra("callback.type", "3")
	return m
}

func (m *Message) SetLauncherActivity() *Message {
	m.AddExtra("notify_effect", "1")
	return m
}

func (m *Message) SetOpenActivity(intentURI string) *Message {
	m.AddExtra("notify_effect", "2")
	m.AddExtra("intent_uri", intentURI)
	return m
}

func (m *Message) SetOpenWebURI(url string) *Message {
	m.AddExtra("notify_effect", "3")
	m.AddExtra("web_uri", url)
	return m
}

func (m *Message) SetTicker(ticker string) *Message {
	m.AddExtra("ticker", ticker)
	return m
}

func (m *Message) EnableNotifyForeground() *Message {
	m.AddExtra("notify_foreground", "1")
	return m
}

func (m *Message) DisableNotifyForeground() *Message {
	m.RemoveExtra("notify_foreground")
	return m
}

func (m *Message) SetNotificationGroup(group string) *Message {
	m.AddExtra("notification_group", group)
	m.AddExtra("notification_is_summary", "true")
	return m
}

type TargetType int8

const (
	TargetRegId   TargetType = 1
	TargetAlias   TargetType = 2
	TargetAccount TargetType = 3
)

type TargetedMessage struct {
	message    *Message
	target     string
	targetType TargetType
}

func NewTargetedMessage(message *Message, target string, targetType TargetType) *TargetedMessage {
	return &TargetedMessage{
		message:    message,
		target:     target,
		targetType: targetType,
	}
}
