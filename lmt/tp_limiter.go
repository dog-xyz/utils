package lmt

import (
	"context"
	"fmt"
	"time"
	"github.com/redis/go-redis/v9"
)

const (
	tokenListKey = "TP_LIMITER_TOKEN_LIST:%s"
	genTokenLock = "TP_LIMITER_GEN_TOKEN_LOCK:%s:%d"
	rateKey = "TP_LIMITER_RATE:%s"
)

type LoggerFunc func(quickId string, format string, args ...interface{})

func defaultLogger(quickId string, format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

type TpLimiter struct {
	name string
	capacity int
	rate float64
	redis *redis.Client
	done chan struct{}
	rateChan chan float64
	logger LoggerFunc
}

func NewTpLimiter(name string, rate float64, capacity int, redis *redis.Client, logger LoggerFunc) *TpLimiter {
	if logger == nil {
		logger = defaultLogger
	}
	if rate <= 0 {
		return nil
	}
	tp := &TpLimiter{
		name: name,
		capacity: capacity,
		rate: rate,
		redis: redis,
		done: make(chan struct{}),
		rateChan: make(chan float64, 1),
		logger: logger,
	}
	err := tp.redis.Set(context.Background(), fmt.Sprintf(rateKey, tp.name), rate, 0).Err()
	if err != nil {
		tp.logger(fmt.Sprintf("limiter-%s", tp.name), "set rate failed, err: %s", err.Error())
		return nil
	}
	go tp.genToken()
	go tp.checkRate()
	return tp
}

func (tp *TpLimiter) Stop() {
	close(tp.done)
}

func (tp *TpLimiter) SetRate(newRate float64) {
	if newRate <= 0 {
		tp.logger(fmt.Sprintf("limiter-%s", tp.name), "invalid rate: %f", newRate)
		return
	}

	// Update rate in Redis
	err := tp.redis.Set(context.Background(), fmt.Sprintf(rateKey, tp.name), newRate, 0).Err()
	if err != nil {
		tp.logger(fmt.Sprintf("limiter-%s", tp.name), "set rate failed, err: %s", err.Error())
		return
	}

	// Try to send new rate to channel, if full then drain it first
	select {
	case tp.rateChan <- newRate:
		// Successfully sent
	default:
		// Channel is full, drain it and send new rate
		for len(tp.rateChan) > 0 {
			<-tp.rateChan
		}
		tp.rateChan <- newRate
	}
}

func (tp *TpLimiter) checkRate() {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	quickId := fmt.Sprintf("limiter-%s", tp.name)
	
	for {
		select {
		case <-t.C:
			rate, err := tp.redis.Get(context.Background(), fmt.Sprintf(rateKey, tp.name)).Float64()
			if err != nil {
				if err != redis.Nil {
					tp.logger(quickId, "get rate failed, err: %s", err.Error())
				}
				continue
			}
			
			if rate != tp.rate {
				tp.SetRate(rate)
			}
		case <-tp.done:
			return
		}
	}
}

func (tp *TpLimiter) genToken() {
	interval := time.Duration(float64(time.Second) / tp.rate)
	t := time.NewTicker(interval)
	defer t.Stop()
	quickId := fmt.Sprintf("limiter-%s", tp.name)
	for {
		select {
		case newRate := <-tp.rateChan:
			tp.rate = newRate
			interval = time.Duration(float64(time.Second) / tp.rate)
			t.Reset(interval)
		case <-t.C:
			lockKey := fmt.Sprintf(genTokenLock, tp.name, time.Now().UnixNano()/int64(interval))
			lock := tp.redis.SetNX(context.Background(), lockKey, 1, interval)
			if lock.Err() != nil {
				tp.logger(quickId, "setnx failed, err: %s", lock.Err().Error())
				continue
			}
			if !lock.Val() {
				continue
			}
			tpKey := fmt.Sprintf(tokenListKey, tp.name)
			llenRet, err := tp.redis.LLen(context.Background(), tpKey).Result()
			if err != nil {
				continue
			}
			if llenRet < int64(tp.capacity) {
				err = tp.redis.LPush(context.Background(), tpKey, 1).Err()
				if err != nil {
					tp.logger(quickId, "lpush failed, err: %s", err.Error())
					continue
				}
			}
		case <-tp.done:
			return
		}
	}
}

func (tp *TpLimiter) Allow() bool {
	tpKey := fmt.Sprintf(tokenListKey, tp.name)
	err := tp.redis.LPop(context.Background(), tpKey).Err()
	return err == nil
}

func (tp *TpLimiter) Len() int {
	quickId := fmt.Sprintf("limiter-%s", tp.name)
	tpKey := fmt.Sprintf(tokenListKey, tp.name)
	llenRet, err := tp.redis.LLen(context.Background(), tpKey).Result()
	if err != nil {
		tp.logger(quickId, "get llen failed, err: %s", err.Error())
		return 0
	}
	return int(llenRet)
}

const (
	checkInterval = 20 * time.Millisecond
	waitTimeout = 120 * time.Second
)

func (tp *TpLimiter) WaitAllow() bool {
	t := time.NewTicker(checkInterval)
	defer t.Stop()
	timeout := time.After(waitTimeout)
	for {
		select {
		case <-t.C:
			if tp.Allow() {
				return true
			}
		case <-timeout:
			return false
		}
	}
}

func (tp *TpLimiter) GetRate() float64 {
	return tp.rate
}
