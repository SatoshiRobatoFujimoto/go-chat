package main

import (
	"time"

	"github.com/gorilla/websocket"
)

// クライアントの構造体
type client struct {
	// このクライアントのためのWebSocket
	socket *websocket.Conn

	// メッセージが送られるチャネル
	send chan *message

	// クライアントが参加しているチャットルーム
	room *room

	// ユーザ情報
	userData map[string]interface{}
}

// メッセージの読み込み
func (c *client) read() {
	for {
		var msg *message
		if err := c.socket.ReadJSON(&msg); err == nil {
			// チャットルームからのメッセージを受信
			// 名前と投稿時刻の取得
			msg.When = time.Now()
			msg.Name = c.userData["name"].(string)

			// アバターの取得
			if avatarURL, ok := c.userData["avatar_url"]; ok {
				msg.AvatarURL = avatarURL.(string)
			}

			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}

// メッセージの書き込み
func (c *client) write() {

	for msg := range c.send {
		if err := c.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	c.socket.Close()
}
