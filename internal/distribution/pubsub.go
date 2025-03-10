package distribution

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)


type MessageHandler func(msg *redis.Message)

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
    // subscribe
    err := ps.pubsub.Subscribe(context.Background(), ps.listenChannel)
    if err != nil{
        log.Fatal("FATAL ERR: ", err)
    }

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



