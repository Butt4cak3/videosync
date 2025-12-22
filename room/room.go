package room

import (
	"sync"
	"time"
	"videosync/message"
	"videosync/youtube"
)

type Room struct {
	Id       string
	users    []*User
	mu       sync.Mutex
	playback Playback
	stopSync chan bool
}

func NewRoom(id string) *Room {
	return &Room{
		Id:       id,
		users:    make([]*User, 0, 2),
		stopSync: make(chan bool),
	}
}

func (room *Room) SyncState() {
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			if room.playback.Position() > room.playback.Duration {
				room.Load(room.playback.VideoId)
				time.Sleep(time.Second)
				room.Play(nil, 0)
			}
		case <-room.stopSync:
			return
		}
	}
}

func (room *Room) close() {
	room.stopSync <- true
}

func (room *Room) Join(user *User) {
	room.mu.Lock()
	defer room.mu.Unlock()
	room.users = append(room.users, user)
	users := make([]string, len(room.users))

	for i, user := range room.users {
		users[i] = user.Name
	}

	user.Conn.WriteJSON(message.Message{
		Type: message.Init,
		Payload: message.InitMessage{
			VideoId:       room.playback.VideoId,
			VideoPos:      room.playback.Position(),
			PlaybackState: int(room.playback.State),
			Users:         users,
		},
	})
	room.Send(user, message.Message{
		Type:    message.Join,
		Payload: message.JoinMessage{UserName: user.Name},
	})
}

func (room *Room) Leave(user *User) {
	room.mu.Lock()
	defer room.mu.Unlock()
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
	room.Send(user, message.Message{
		Type:    message.Leave,
		Payload: message.LeaveMessage{UserName: user.Name},
	})
}

func (room *Room) Play(user *User, position float32) {
	room.mu.Lock()
	defer room.mu.Unlock()
	room.playback.LatestPosition = position
	room.playback.LatestPositionTime = time.Now()
	room.playback.State = Playing
	room.Send(user, message.Message{Type: message.Play, Payload: message.PlayMessage{Position: position}})
}

func (room *Room) Pause(user *User, position float32) {
	room.mu.Lock()
	defer room.mu.Unlock()
	room.playback.LatestPosition = position
	room.playback.LatestPositionTime = time.Now()
	room.playback.State = Paused
	room.Send(user, message.Message{Type: message.Pause, Payload: message.PauseMessage{Position: position}})
}

func (room *Room) Load(videoId string) {
	room.mu.Lock()
	defer room.mu.Unlock()
	duration, err := youtube.GetVideoDuration(videoId)
	if err != nil {
		return
	}

	room.playback.VideoId = videoId
	room.playback.LatestPosition = 0
	room.playback.LatestPositionTime = time.Now()
	room.playback.State = Paused
	room.playback.Duration = float32(duration.Seconds())

	room.Send(nil, message.Message{Type: message.Load, Payload: message.LoadMessage{VideoId: videoId}})
}

func (room *Room) Kick(user *User) {
	user.Conn.Close()
}

func (room *Room) Send(from *User, message message.Message) {
	for _, user := range room.users {
		if user != from {
			user.Conn.WriteJSON(message)
		}
	}
}
