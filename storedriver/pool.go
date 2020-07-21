package storedriver

import (
	"errors"
	"sync"

	"github.com/go-redis/redis"
)

type Pool struct {
	m         sync.Mutex
	resources chan *redis.Client
	factory   *RedisClient
	closed    bool
}

var ErrPoolClosed = errors.New("Pool has been closed.")

func NewPool(cn *RedisClient, size int) (*Pool, error) {
	if size <= 0 {
		return nil, errors.New("Size value too small.")
	}
	return &Pool{resources: make(chan *redis.Client, size), factory: cn}, nil
}

func (p *Pool) Acquire() (*redis.Client, error) {
	select {
	case res, ok := <-p.resources:
		if !ok {
			return nil, ErrPoolClosed
		}
		return res, nil
	default:
		return p.factory.Connect()

	}
}

func (p *Pool) Release(r *redis.Client) {
	p.m.Lock()
	defer p.m.Unlock()
	if p.closed {
		r.Close()
		return
	}
	select {
	case p.resources <- r:
		return
	default:
		r.Close()
	}

}

func (p *Pool) Close() {
	p.m.Lock()
	defer p.m.Unlock()
	if p.closed {
		return
	}
	p.closed = true
	close(p.resources)
	for r := range p.resources {
		r.Close()
	}

}
