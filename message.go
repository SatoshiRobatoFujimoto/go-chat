package main

import (
	"time"
)

// １つのメッセージの構造体
type message struct {
	Name      string
	Message   string
	When      time.Time
	AvatarURL string
}
