// internal/cache/redis.go
//
// A minimal Redis cache adapter. It uses raw TCP connections and
// RESP protocol for GET/SET with EX.
//
// This avoids external dependencies while remaining fully functional.

package cache

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"time"
)

type redisCache struct {
	addr string
	ttl  time.Duration
}

func NewRedis(addr string, ttl time.Duration) Cache {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	return &redisCache{
		addr: addr,
		ttl:  ttl,
	}
}

func (r *redisCache) Get(key string) ([]byte, bool) {
	conn, err := net.Dial("tcp", r.addr)
	if err != nil {
		return nil, false
	}
	defer conn.Close()

	fmt.Fprintf(conn, "*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(key), key)

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil || len(line) == 0 {
		return nil, false
	}

	if line[0] == '$' {
		var length int
		fmt.Sscanf(line, "$%d", &length)
		if length < 0 {
			return nil, false
		}
		buf := make([]byte, length)
		io.ReadFull(reader, buf)
		reader.ReadString('\n') // CRLF
		return buf, true
	}

	return nil, false
}

func (r *redisCache) Set(key string, value []byte, ttl time.Duration) {
	if ttl <= 0 {
		ttl = r.ttl
	}

	conn, err := net.Dial("tcp", r.addr)
	if err != nil {
		return
	}
	defer conn.Close()

	fmt.Fprintf(conn,
		"*5\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n$2\r\nPX\r\n$%d\r\n%d\r\n",
		len(key), key,
		len(value), value,
		len(fmt.Sprint(ttl.Milliseconds())), ttl.Milliseconds(),
	)
}
