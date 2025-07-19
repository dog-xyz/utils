package bes

import (
	"fmt"
	"sync"
	"time"
)

const (
	// 时间戳位数 (41位，可以使用69年)
	timestampBits = 41
	// 数据中心ID位数 (5位，最多32个数据中心)
	datacenterBits = 5
	// 机器ID位数 (5位，每个数据中心最多32台机器)
	machineBits = 5
	// 序列号位数 (12位，每毫秒最多4096个ID)
	sequenceBits = 12

	// 最大值
	maxDatacenterID = (1 << datacenterBits) - 1
	maxMachineID    = (1 << machineBits) - 1
	maxSequence     = (1 << sequenceBits) - 1

	// 偏移量
	machineIDShift    = sequenceBits
	datacenterIDShift = sequenceBits + machineBits
	timestampShift    = sequenceBits + machineBits + datacenterBits

	// 时间戳起始点 (2020-01-01 00:00:00 UTC)
	epoch = int64(1577836800000)
)

// Snowflake Snowflake ID 生成器
type Snowflake struct {
	mutex         sync.Mutex
	lastTimestamp int64
	datacenterID  int64
	machineID     int64
	sequence      int64
}

// NewSnowflake 创建新的 Snowflake 实例
func NewSnowflake(datacenterID, machineID int64) (*Snowflake, error) {
	if datacenterID < 0 || datacenterID > maxDatacenterID {
		return nil, fmt.Errorf("datacenter ID must be between 0 and %d", maxDatacenterID)
	}
	if machineID < 0 || machineID > maxMachineID {
		return nil, fmt.Errorf("machine ID must be between 0 and %d", maxMachineID)
	}

	return &Snowflake{
		lastTimestamp: 0,
		datacenterID:  datacenterID,
		machineID:     machineID,
		sequence:      0,
	}, nil
}

// NextID 生成下一个 ID
func (s *Snowflake) NextID() (int64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := s.getCurrentTimestamp()

	// 如果当前时间小于上次生成时间，说明系统时钟回退了
	if now < s.lastTimestamp {
		return 0, fmt.Errorf("clock moved backwards, refusing to generate ID for %d milliseconds", s.lastTimestamp-now)
	}

	// 如果是同一毫秒内，增加序列号
	if now == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		// 如果序列号溢出，等待下一毫秒
		if s.sequence == 0 {
			now = s.waitNextMillis(s.lastTimestamp)
		}
	} else {
		// 不同毫秒，序列号重置为0
		s.sequence = 0
	}

	s.lastTimestamp = now

	// 生成ID
	id := ((now - epoch) << timestampShift) |
		(s.datacenterID << datacenterIDShift) |
		(s.machineID << machineIDShift) |
		s.sequence

	return id, nil
}

// getCurrentTimestamp 获取当前时间戳（毫秒）
func (s *Snowflake) getCurrentTimestamp() int64 {
	return time.Now().UnixNano() / 1e6
}

// waitNextMillis 等待下一毫秒
func (s *Snowflake) waitNextMillis(lastTimestamp int64) int64 {
	timestamp := s.getCurrentTimestamp()
	for timestamp <= lastTimestamp {
		timestamp = s.getCurrentTimestamp()
	}
	return timestamp
}

// ParseID 解析 Snowflake ID
func (s *Snowflake) ParseID(id int64) map[string]int64 {
	timestamp := (id >> timestampShift) + epoch
	datacenterID := (id >> datacenterIDShift) & maxDatacenterID
	machineID := (id >> machineIDShift) & maxMachineID
	sequence := id & maxSequence

	return map[string]int64{
		"timestamp":     timestamp,
		"datacenter_id": datacenterID,
		"machine_id":    machineID,
		"sequence":      sequence,
	}
}

// GetTimestampFromID 从 ID 中提取时间戳
func GetTimestampFromID(id int64) int64 {
	return (id >> timestampShift) + epoch
}

// GetDatacenterIDFromID 从 ID 中提取数据中心ID
func GetDatacenterIDFromID(id int64) int64 {
	return (id >> datacenterIDShift) & maxDatacenterID
}

// GetMachineIDFromID 从 ID 中提取机器ID
func GetMachineIDFromID(id int64) int64 {
	return (id >> machineIDShift) & maxMachineID
}

// GetSequenceFromID 从 ID 中提取序列号
func GetSequenceFromID(id int64) int64 {
	return id & maxSequence
}

// 全局 Snowflake 实例
var (
	globalSnowflake *Snowflake
	snowflakeOnce   sync.Once
)

// InitGlobalSnowflake 初始化全局 Snowflake 实例
func InitGlobalSnowflake(datacenterID, machineID int64) error {
	var err error
	snowflakeOnce.Do(func() {
		globalSnowflake, err = NewSnowflake(datacenterID, machineID)
	})
	return err
}

// GenerateID 使用全局实例生成 ID
func GenerateID() (int64, error) {
	if globalSnowflake == nil {
		return 0, fmt.Errorf("global snowflake not initialized, call InitGlobalSnowflake first")
	}
	return globalSnowflake.NextID()
}

// GenerateIDString 生成字符串格式的 ID
func GenerateIDString() (string, error) {
	id, err := GenerateID()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", id), nil
}

// ParseGlobalID 解析全局生成的 ID
func ParseGlobalID(id int64) map[string]int64 {
	if globalSnowflake == nil {
		return nil
	}
	return globalSnowflake.ParseID(id)
}
