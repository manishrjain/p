package conn

// In package x: func Log(p string) *logrus.Entry
import (
	"log"
	"net"
	"net/rpc"
	"strings"
	"time"
)

type Pool struct {
	clients chan *rpc.Client
	Addr    string
}

func NewPool(addr string, maxCap int) *Pool {
	p := new(Pool)
	p.Addr = addr
	p.clients = make(chan *rpc.Client, maxCap)
	client, err := p.dialNew()
	if err != nil {
		log.Fatal(err)
		return nil
	}
	p.clients <- client
	return p
}

func (p *Pool) dialNew() (*rpc.Client, error) {
	d := &net.Dialer{Timeout: 3 * time.Minute}
	var nconn net.Conn
	var err error
	for i := 0; i < 60; i++ { // This loop will retry for 10 minutes before giving up.
		nconn, err = d.Dial("tcp", p.Addr)
		if err == nil {
			break
		}
		if !strings.Contains(err.Error(), "refused") {
			break
		}
		log.Print("Retrying connection...")
		time.Sleep(10 * time.Second)
	}
	if err != nil {
		return nil, err
	}
	cc := &ClientCodec{
		Rwc: nconn,
	}
	return rpc.NewClientWithCodec(cc), nil
}

func (p *Pool) Call(serviceMethod string, args interface{}, reply interface{}) error {
	client, err := p.get()
	if err != nil {
		return err
	}
	if err = client.Call(serviceMethod, args, reply); err != nil {
		return err
	}

	select {
	case p.clients <- client:
		return nil
	default:
		return client.Close()
	}
}

func (p *Pool) get() (*rpc.Client, error) {
	select {
	case client := <-p.clients:
		return client, nil
	default:
		return p.dialNew()
	}
}

func (p *Pool) Close() error {
	// We're not doing a clean exit here. A clean exit here would require
	// synchronization, which seems unnecessary for now. But, we should
	// add one if required later.
	return nil
}
