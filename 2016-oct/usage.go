package conn

import (
	"context"
	"fmt"

	"github.com/dgraph-io/dgraph/x"
)

func send(ctx context.Context, pl *Pool, data []byte) (rd []byte, rerr error) {
	conn, err := pl.Get()
	if err != nil {
		x.TraceError(ctx, err)
		return rd, err
	}
	defer pl.Put(conn)
	query := new(Payload)
	query.Data = data

	c := NewWorkerClient(conn)
	reply, err := c.Mutate(ctx, query)
	if reply == nil {
		return rd, err
	}
	return reply.Data, err
}

func main() {
	fmt.Println("vim-go")
}
