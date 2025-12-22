package room

import (
	"log"
	"sync"
)

type RoomManager struct {
	rooms map[string]*Room
	mu    sync.Mutex
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

func (rm *RoomManager) Get(id string) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if r, ok := rm.rooms[id]; ok {
		return r
	} else {
		r := NewRoom(id)
		log.Printf("Created room %s\n", id)
		go r.SyncState()
		go func() {
			defer rm.Delete(id)
			r.WatchEvents()
		}()
		r.load("Ne7fbb9c-BU")
		rm.rooms[id] = r
		return r
	}
}

func (rm *RoomManager) Delete(id string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.rooms, id)
	log.Printf("Deleted room %s\n", id)
}
