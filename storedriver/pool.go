package storedriver

import (
	"sync"
	"io"
	"errors"
)
type Pool struct {
	m sync.Mutex
	resources chan io.Closer
	factory Connecter
	closed bool
}

type Connecter interface {
	Connect() (io.Closer,error)
}
var ErrPoolClosed = errors.New("Pool has been closed.")
func NewPool(cn Connecter, size int)(*Pool,error){
	if size <= 0 {
		return  nil, errors.New("Size value too small.")
	}
	return &Pool{resources: make(chan io.Closer, size),factory: cn},nil
}

func (p *Pool) Acquire()(io.Closer, error){
	select {
	case res, ok := <- p.resources:
		if !ok{
			return nil, ErrPoolClosed
		}
		return res,nil
	default:
		return p.factory.Connect()

	}
}

func (p *Pool)Release(r io.Closer){
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

func (p *Pool)Close(){
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
