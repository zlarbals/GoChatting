package main

import (
	"github.com/gorilla/websocket"
	"log"
	"time"
)

var clients []*Client

const messageBufferSize = 256

type Client struct {
	conn *websocket.Conn //웹소켓 커넥션
	send chan *Message   //메시지 전송용 채널

	roomId string //현재 접속한 채팅방 아이디
	user   *User  //현재 접속한 사용자 정보
}

func newClient(conn *websocket.Conn, roomId string, u *User) {
	//새로운 클라이언트 생성
	c := &Client{
		conn:   conn,
		send:   make(chan *Message, messageBufferSize),
		roomId: roomId,
		user:   u,
	}

	//clients 목록에 새로 생성한 클라이언트 추가
	clients = append(clients, c)

	//메시지 수신/전송 대기
	go c.readLoop()
	go c.writeLoop()
}

func (c *Client) Close() {
	//clients 목록에서 종료된 클라이언트 제거
	for i, client := range clients {
		if client == c {
			clients = append(clients[:i], clients[i+1:]...)
			break
		}
	}

	//send 채널 닫음
	close(c.send)

	//웹소켓 커넥션 종료
	c.conn.Close()
	log.Printf("close connection. addr: %s", c.conn.RemoteAddr())
}

func (c *Client) readLoop() {
	//메시지 수신 대기
	for {
		m, err := c.read()
		if err != nil {
			//오류가 발생하면 메시지 수신 루프 종료
			log.Println("read message error: ", err)
			break
		}

		//메시지가 수신되면 수신된 메시지를 몽고DB에 생성하고 모든 clients에 전달
		m.create()
		broadcast(m)
	}
	c.Close()
}

func (c *Client) writeLoop() {
	//클라이언트의 send 채널 메시지 수신 대기
	for msg := range c.send {
		//클라이언트의 채팅방 아이디와 전달된 메시지의 채팅방 아이디가 일치하면 웹소켓에 메시지 전달
		if c.roomId == msg.RoomId.Hex() {
			c.write(msg)
		}
	}
}

func broadcast(m *Message){
	//모든 클라이언트의 send 채널에 메시지 전달
	for _, client := range clients{
		client.send<-m
	}
}

func (c *Client) read() (*Message, error){
	var msg *Message

	//웹소켓 커녁션에 JSON 형태의 메시지가 전달되면 Message 타입으로 메시지를 읽음
	if err:=c.conn.ReadJSON(&msg); err!=nil{
		return nil,err
	}

	//Message에 현재 시간과 사용자 정보 세팅
	msg.CreatedAt=time.Now()
	msg.User=c.user

	log.Println("read from websocket:",msg)

	//메시지 정보 반환
	return msg,nil
}

func (c *Client) write(m *Message) error{
	log.Println("write to websocket:",m)

	//웹소켓 커넥션에 JSON 형태로 메시지 전달
	return c.conn.WriteJSON(m)
}