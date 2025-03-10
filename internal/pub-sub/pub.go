package pubsub

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"github.com/myselfBZ/chatrix/internal/events"
)



type Publisher interface{
    Publish(context.Context, string, *events.Event) error
}


func NewPub(client *redis.Client) Publisher{
    return &RedisPublisher{
        client,
    }
}



type RedisPublisher struct{
    client *redis.Client
}


func (p *RedisPublisher) Publish(ctx context.Context, channelName string, e *events.Event) error {
    jsonvalue, _ := json.Marshal(e)
    return p.client.Publish(ctx, channelName, string(jsonvalue)).Err()
}

