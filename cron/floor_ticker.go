package util

import (
	"github.com/robfig/cron/v3"
	"time"
)

type FloorTicker struct {
	C      chan bool
	c      *cron.Cron
	CurrTs int64
	LastTs int64
}

func NewFloorTicker(cronArgs string) (ft *FloorTicker, err error) {
	ft = &FloorTicker{
		C:      make(chan bool),
		c:      cron.New(cron.WithSeconds()),
		CurrTs: time.Now().Unix(),
		LastTs: time.Now().Unix(),
	}
	//add cron job
	_, err = ft.c.AddJob(cronArgs, cron.NewChain(cron.SkipIfStillRunning(cron.DefaultLogger)).Then(ft))
	if err != nil {
		return
	}
	ft.c.Start()

	return
}

func (ft *FloorTicker) Stop() {
	ft.c.Stop()
}

func (ft *FloorTicker) Run() {
	ft.LastTs = ft.CurrTs
	ft.CurrTs = time.Now().Unix()
	select {
	case ft.C <- true:
	default:
	}
}
