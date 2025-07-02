package util

import (
	"testing"
	"time"
)

func TestFloorTicker(t *testing.T) {
	ft, err := NewFloorTicker("0 * * * * *")
	if err != nil {
		t.Error(err)
	}
	defer ft.Stop()

	go func() {
		for {
			select {
			case <-ft.C:
				t.Logf("tick at: %s, last: %s\n", time.Unix(ft.CurrTs, 0).Format("2006-01-02 15:04:05"), time.Unix(ft.LastTs, 0).Format("2006-01-02 15:04:05"))
			}
		}
	}()

	time.Sleep(300 * time.Second)
}
