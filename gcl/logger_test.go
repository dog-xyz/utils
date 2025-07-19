package gcl

import "testing"


func TestLogger(t *testing.T) {
	logger := NewGCLogger("test", "test", nil)
	logger.Debugf("test", "test debug")
	logger.Infof("test", "test info")
	logger.Errorf("test", "test error")
}