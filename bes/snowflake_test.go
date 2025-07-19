package bes

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSnowflakeBasic(t *testing.T) {
	// 创建 Snowflake 实例
	snowflake, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("Failed to create snowflake: %v", err)
	}

	// 生成多个 ID
	ids := make([]int64, 10)
	for i := 0; i < 10; i++ {
		id, err := snowflake.NextID()
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}
		ids[i] = id
		fmt.Printf("Generated ID: %d\n", id)
	}

	// 验证 ID 是递增的
	for i := 1; i < len(ids); i++ {
		if ids[i] <= ids[i-1] {
			t.Errorf("ID should be increasing: %d <= %d", ids[i], ids[i-1])
		}
	}

	// 解析 ID
	for i, id := range ids {
		parsed := snowflake.ParseID(id)
		fmt.Printf("ID %d parsed: timestamp=%d, datacenter_id=%d, machine_id=%d, sequence=%d\n",
			i+1, parsed["timestamp"], parsed["datacenter_id"], parsed["machine_id"], parsed["sequence"])
	}
}

func TestSnowflakeConcurrent(t *testing.T) {
	// 创建 Snowflake 实例
	snowflake, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("Failed to create snowflake: %v", err)
	}

	// 并发生成 ID
	const numGoroutines = 10
	const idsPerGoroutine = 100
	var wg sync.WaitGroup
	idChan := make(chan int64, numGoroutines*idsPerGoroutine)

	// 启动多个 goroutine 并发生成 ID
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				id, err := snowflake.NextID()
				if err != nil {
					t.Errorf("Failed to generate ID: %v", err)
					return
				}
				idChan <- id
			}
		}()
	}

	// 等待所有 goroutine 完成
	wg.Wait()
	close(idChan)

	// 收集所有 ID
	ids := make([]int64, 0, numGoroutines*idsPerGoroutine)
	for id := range idChan {
		ids = append(ids, id)
	}

	// 验证 ID 数量
	if len(ids) != numGoroutines*idsPerGoroutine {
		t.Errorf("Expected %d IDs, got %d", numGoroutines*idsPerGoroutine, len(ids))
	}

	// 验证 ID 唯一性
	idSet := make(map[int64]bool)
	for _, id := range ids {
		if idSet[id] {
			t.Errorf("Duplicate ID found: %d", id)
		}
		idSet[id] = true
	}

	fmt.Printf("Generated %d unique IDs concurrently\n", len(ids))
}

func TestSnowflakePerformance(t *testing.T) {
	// 创建 Snowflake 实例
	snowflake, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("Failed to create snowflake: %v", err)
	}

	// 性能测试
	const numIDs = 100000
	start := time.Now()

	for i := 0; i < numIDs; i++ {
		_, err := snowflake.NextID()
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}
	}

	duration := time.Since(start)
	rate := float64(numIDs) / duration.Seconds()

	fmt.Printf("Generated %d IDs in %v (%.2f IDs/sec)\n", numIDs, duration, rate)
}

func TestGlobalSnowflake(t *testing.T) {
	// 初始化全局 Snowflake
	err := InitGlobalSnowflake(2, 2)
	if err != nil {
		t.Fatalf("Failed to init global snowflake: %v", err)
	}

	// 使用全局实例生成 ID
	ids := make([]int64, 5)
	for i := 0; i < 5; i++ {
		id, err := GenerateID()
		if err != nil {
			t.Fatalf("Failed to generate global ID: %v", err)
		}
		ids[i] = id
		fmt.Printf("Global ID %d: %d\n", i+1, id)
	}

	// 生成字符串格式的 ID
	idStr, err := GenerateIDString()
	if err != nil {
		t.Fatalf("Failed to generate ID string: %v", err)
	}
	fmt.Printf("ID string: %s\n", idStr)

	// 解析全局生成的 ID
	for i, id := range ids {
		parsed := ParseGlobalID(id)
		if parsed == nil {
			t.Errorf("Failed to parse global ID: %d", id)
			continue
		}
		fmt.Printf("Global ID %d parsed: timestamp=%d, datacenter_id=%d, machine_id=%d, sequence=%d\n",
			i+1, parsed["timestamp"], parsed["datacenter_id"], parsed["machine_id"], parsed["sequence"])
	}
}

func TestSnowflakeIDComponents(t *testing.T) {
	// 创建 Snowflake 实例
	snowflake, err := NewSnowflake(3, 4)
	if err != nil {
		t.Fatalf("Failed to create snowflake: %v", err)
	}

	// 生成 ID
	id, err := snowflake.NextID()
	if err != nil {
		t.Fatalf("Failed to generate ID: %v", err)
	}

	// 测试各个组件提取函数
	timestamp := GetTimestampFromID(id)
	datacenterID := GetDatacenterIDFromID(id)
	machineID := GetMachineIDFromID(id)
	sequence := GetSequenceFromID(id)

	fmt.Printf("ID: %d\n", id)
	fmt.Printf("Timestamp: %d (%s)\n", timestamp, time.Unix(timestamp/1000, 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("Datacenter ID: %d\n", datacenterID)
	fmt.Printf("Machine ID: %d\n", machineID)
	fmt.Printf("Sequence: %d\n", sequence)

	// 验证组件值
	if datacenterID != 3 {
		t.Errorf("Expected datacenter ID 3, got %d", datacenterID)
	}
	if machineID != 4 {
		t.Errorf("Expected machine ID 4, got %d", machineID)
	}
}

func TestSnowflakeValidation(t *testing.T) {
	// 测试无效的数据中心 ID
	_, err := NewSnowflake(32, 1) // 超过最大值 31
	if err == nil {
		t.Error("Expected error for invalid datacenter ID")
	}

	// 测试无效的机器 ID
	_, err = NewSnowflake(1, 32) // 超过最大值 31
	if err == nil {
		t.Error("Expected error for invalid machine ID")
	}

	// 测试负数
	_, err = NewSnowflake(-1, 1)
	if err == nil {
		t.Error("Expected error for negative datacenter ID")
	}

	_, err = NewSnowflake(1, -1)
	if err == nil {
		t.Error("Expected error for negative machine ID")
	}

	fmt.Println("Validation tests passed")
}
