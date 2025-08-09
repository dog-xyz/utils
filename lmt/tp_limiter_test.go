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


func TestTpLimiter_SetRate(t *testing.T) {
	redis := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})
	limiter1 := NewTpLimiter("test", 0.25, 1, redis, nil)
	limiter2 := NewTpLimiter("test", 0.25, 1, redis, nil)

	go func() {
		for {
			limiter1.WaitAllow()
			fmt.Printf("1 allow at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		}
	}()
	go func() {
		for {
			limiter2.WaitAllow()
			fmt.Printf("2 allow at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		}
	}()

	fmt.Printf("1 rate: %v, 2 rate: %v\n", limiter1.GetRate(), limiter2.GetRate())

	time.Sleep(time.Second * 16)
	limiter1.SetRate(3)
	time.Sleep(time.Second * 1)
	fmt.Printf("limit2 rate: %v\n", limiter2.GetRate())
	time.Sleep(time.Second * 12)
}