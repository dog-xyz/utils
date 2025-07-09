package logger

import "testing"


func TestLogger(t *testing.T) {
	lg := NewDefaultLogger(InitArgs{
		SrvName:    "test",
		SrvHost:    "127.0.0.1",
	})

	lg.Infof("test")
	lg.Errorf("test")
	lg.Warnf("test")
	lg.Debugf("test")

	lg.Stop()
}