package lmt

import (
	"testing"

	"github.com/redis/go-redis/v9"
	"fmt"
	"time"
)


func TestTpLimiter(t *testing.T) {
	redis := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})
	limiter := NewTpLimiter("test", 0.5, 1, redis, nil)
	go func() {
		for {
			limiter.WaitAllow()
			fmt.Printf("allow at %s\n", time.Now().Format("2006-01-02 15:04:05"))
		}
	}()
	time.Sleep(time.Second * 10)
	limiter.SetRate(2)
	time.Sleep(time.Second * 10)
}