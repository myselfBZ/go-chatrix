package distribution

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)


type MessageHandler func(msg *redis.Message)



func NewPubSub(client *redis.Client, handler MessageHandler, channel string) *PubSub {
    pubsub := client.Subscribe(context.TODO(), channel)
    ps := &PubSub{
        client: client,
        pubsub: pubsub,
        handler: handler,
        listenChannel:channel,
        done: make(chan struct{}),
    }
    return ps
}

type PubSub struct{
    client *redis.Client
    handler MessageHandler
    pubsub  *redis.PubSub
    listenChannel string

    done chan struct{}
}


func (ps *PubSub) Publish(channel string, msg interface{}) error {
    data, err := json.Marshal(msg)
    if err != nil{
        return err
    }
    ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)
    defer cancel()
    return ps.client.Publish(ctx ,channel, data).Err()
} 

func (ps *PubSub) Start() {

    go func(){
        for{
            select{
            case <- ps.done:
                return
            case msg := <- ps.pubsub.Channel():
                ps.handler(msg)
            }
        }
    }()
}
