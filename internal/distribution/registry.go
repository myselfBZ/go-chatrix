package distribution

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)


type Registry struct{
    client *redis.Client
    ttl time.Duration
    serverAddr string
}

func (r *Registry) Lookup(username string) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second * 2)
    defer cancel()
    return r.client.Get(ctx, username).Result()
}

func (r *Registry) Register(username string) error{
    ctx, cancel := context.WithTimeout(context.Background(), time.Second * 2)
    defer cancel()
    return r.client.Set(ctx, r.userKey(username), r.serverAddr, r.ttl).Err()
}

func (r *Registry) userKey(username string) string{
    return fmt.Sprintf("user:%s", username)
}
