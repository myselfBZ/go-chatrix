package pubsub

import (
	"encoding/json"
	"log"

	"gioui.org/io/event"
	"github.com/go-redis/redis/v8"
	"github.com/myselfBZ/chatrix/internal/events"
)

type Subcriber interface {
    Channel() <- chan *events.Event
    Run() error
}


func NewSub(sub *redis.PubSub) Subcriber {
    return &RedisSub{
        sub: sub,
        eventChan: make(chan *events.Event),
    }
}


type RedisSub struct{
    sub *redis.PubSub
    eventChan chan *events.Event
}


func (s *RedisSub) Run() error {
    return s.listenForEvents()    
}

func (s *RedisSub) Channel() <- chan *events.Event {
    return s.eventChan
}

func (s *RedisSub) listenForEvents() error {
    for msg := range s.sub.Channel() {
        var event event.Event
        if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil{
            log.Fatal("why not?", err)
        }
        s.eventChan <- &event
    }
    return nil
}
