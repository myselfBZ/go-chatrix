package messaging

import (
	"log"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/myselfBZ/chatrix/internal/distribution"
)

func NewPool(serverAddr string, client *redis.Client) ConnectionManager {
    return &Pool{
        conns: sync.Map{},
        registry: distribution.NewRegistry(client, serverAddr),
    }
}

type ConnectionManager interface{
    GetServerAddr(username string) (string, error)
    Get(username string) *Client
    Add(*Client)
    Remove(username string)
}

type Pool struct{
    conns sync.Map

    registry *distribution.Registry
}

func (p *Pool) GetServerAddr(username string) (string, error) {
    return p.registry.Lookup(username)
}


func (p *Pool) Get(username string) (*Client) {
    v, ok := p.conns.Load(username)
    if !ok{
        return nil
    }

    return v.(*Client) 
}

func (p *Pool) Remove(username string) {
    p.conns.Delete(username)

    // delete it from redis registry
    err := p.registry.Del(username)
    if err != nil{
        log.Println("REDIS ERR: ", err)
    }
}

func (p *Pool) Add(client *Client) {
    p.registry.Register(client.User.Username)
    p.conns.Store(client.User.Username, client)
}
