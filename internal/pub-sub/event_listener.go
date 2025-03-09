package pubsub

import (
	"github.com/go-redis/redis/v8"
)


//  listens  and pushes messages to peer servers
type EventPubSub struct{
    Pub Publisher
    Sub Subcriber
}

func New(client *redis.Client , listenChannel string) *EventPubSub {
    pubSub := client.Subscribe(client.Context(), listenChannel)
    pub := NewPub(client)
    sub := NewSub(pubSub)
    return &EventPubSub{
        Pub: pub,
        Sub: sub,
    }
}
