package kv

import (
	"context"

	"github.com/go-redis/redis/v8"
)



func New(client *redis.Client) *KV {
    return &KV{client}
}


type KV struct{
    client *redis.Client
}

// store the client with the server id
func (k *KV) Set(key, value string) error {
    return k.client.Set(context.TODO(), key, value, 0).Err()
} 

// get the client with the server id
func (k *KV) Get(key string)  string {
    return k.client.Get(context.TODO(), key).Val()
}
