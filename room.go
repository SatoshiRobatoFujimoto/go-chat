package main

import (
	"go-chat/trace"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/stretchr/objx"
)

// チャットルームの構造体
type room struct {
	// forwardは他のクライアントに転送するためのメッセージを保持
	forward chan *message

	// joinはチャットに参加しようとしているクライアントのためのチャネル
	join chan *client

	// leaveはチャットから退出しようとしているクライアントのためのチャネル
	leave chan *client

	// clientsは在室している全てのクライアントを保持
	clients map[*client]bool

	// tracerはチャットルームで行われた操作ログを受け取ります
	tracer trace.Tracer
}

func (r *room) run() {
	for {
		select {

		// チャットルーム参加
		case client := <-r.join:
			r.clients[client] = true
			r.tracer.Trace("新しいクライアントが参加しました")

		// チャットルーム退出
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("クライアントが退出しました")

		// 全てのクライアントにメッセージ転送
		case msg := <-r.forward:
			r.tracer.Trace("メッセージを受信しました：", msg.Message)
			for client := range r.clients {
				select {
				// メッセージ送信
				case client.send <- msg:
					r.tracer.Trace("クライアントに送信されました")

				// 送信失敗
				default:
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace("送信失敗。クライアントをクリーンアップします")

				}

			}
		}
	}

}

// 新しいチャットルームの生成
func newRoom() *room {
	return &room{
		forward: make(chan *message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	authCookie, err := req.Cookie("auth")
	if err != nil {
		log.Fatal("クッキーの取得に失敗しました：", err)
		return
	}

	client := &client{
		socket:   socket,
		send:     make(chan *message, messageBufferSize),
		room:     r,
		userData: objx.MustFromBase64(authCookie.Value),
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
