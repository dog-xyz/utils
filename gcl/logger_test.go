package gcl

import "testing"


func TestLogger(t *testing.T) {
	logger := NewGCLogger("test")
	logger.Debugf("test debug")
	logger.Infof("test info")
	logger.Errorf("test error")
}