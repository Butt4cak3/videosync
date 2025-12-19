package internal

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type JoinEvent struct {
	user *User
}

type LeaveEvent struct {
	user *User
}

type PlayEvent struct {
	user     *User
	position float32
}

type PauseEvent struct {
	user     *User
	position float32
}

type Room struct {
	Id       string
	users    []*User
	events   chan any
	mu       sync.Mutex
	logger   *log.Logger
	playback Playback
	stopSync chan bool
}

func NewRoom(id string) *Room {
	return &Room{
		Id:       id,
		users:    make([]*User, 0, 2),
		events:   make(chan any, 10),
		logger:   log.New(os.Stdout, fmt.Sprintf("[#%s] ", id), log.LstdFlags),
		stopSync: make(chan bool),
	}
}

func (room *Room) Dispatch(event any) {
	room.events <- event
}

func (room *Room) SyncState() {
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			if room.playback.Position() > 120 {
				room.logger.Println("Reloading video")
				room.load(room.playback.VideoId)
				time.Sleep(time.Second)
				room.handlePlay(nil, 0)
			}
		case <-room.stopSync:
			return
		}
	}
}

func (room *Room) WatchEvents() {
	room.logger.Println("Watching events")

	for event := range room.events {
		room.mu.Lock()
		switch e := event.(type) {
		case JoinEvent:
			room.handleJoin(e.user)
		case LeaveEvent:
			room.handleLeave(e.user)
		case PlayEvent:
			room.handlePlay(e.user, e.position)
		case PauseEvent:
			room.handlePause(e.user, e.position)
		}
		room.mu.Unlock()
	}

	room.logger.Println("Stopped watching events")
}

func (room *Room) close() {
	room.stopSync <- true
	close(room.events)
}

func (room *Room) Join(user *User) {
	room.Dispatch(JoinEvent{user})
}

func (room *Room) handleJoin(user *User) {
	room.users = append(room.users, user)
	user.Conn.WriteJSON(Message{
		Type: Init,
		Payload: InitMessage{
			VideoId:       room.playback.VideoId,
			VideoPos:      room.playback.Position(),
			PlaybackState: int(room.playback.State),
		},
	})
	room.logger.Printf("Client #%d joined\n", user.Id)
}

func (room *Room) Leave(user *User) {
	room.Dispatch(LeaveEvent{user})
}

func (room *Room) handleLeave(user *User) {
	for i := range len(room.users) {
		if room.users[i] == user {
			room.users[i] = room.users[len(room.users)-1]
			room.users = room.users[:len(room.users)-1]
			break
		}
	}
	if len(room.users) == 0 {
		room.close()
	}
	room.logger.Printf("Client #%d left\n", user.Id)
}

func (room *Room) Play(user *User, position float32) {
	room.Dispatch(PlayEvent{user, position})
}

func (room *Room) handlePlay(user *User, position float32) {
	room.logger.Printf("Playing from %f\n", position)
	room.playback.LatestPosition = position
	room.playback.LatestPositionTime = time.Now()
	room.playback.State = Playing
	room.Send(user, Message{Type: Play, Payload: PlayMessage{position}})
}

func (room *Room) Pause(user *User, position float32) {
	room.Dispatch(PauseEvent{user, position})
}

func (room *Room) handlePause(user *User, position float32) {
	room.logger.Printf("Paused at %f\n", position)
	room.playback.LatestPosition = position
	room.playback.LatestPositionTime = time.Now()
	room.playback.State = Paused
	room.Send(user, Message{Type: Pause, Payload: PauseMessage{position}})
}

func (room *Room) load(videoId string) {
	room.playback.VideoId = videoId
	room.playback.LatestPosition = 0
	room.playback.LatestPositionTime = time.Now()
	room.playback.State = Paused
	room.Send(nil, Message{Type: Load, Payload: LoadMessage{VideoId: videoId}})
}

func (room *Room) Kick(user *User) {
	user.Conn.Close()
}

func (room *Room) Send(from *User, message Message) {
	for _, user := range room.users {
		if user != from {
			user.Conn.WriteJSON(message)
		}
	}
}
