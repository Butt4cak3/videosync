package internal

import "github.com/gorilla/websocket"

type User struct {
	Id   int
	Conn *websocket.Conn
}
