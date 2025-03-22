package service

import (
	"errors"
	"sync"
	"time"
)

const (
	epoch          = 1609459200000 // 2021-01-01 00:00:00 UTC 的毫秒时间戳（自定义起始时间）
	timestampBits  = 41            // 时间戳位数（最多支持 69 年）
	datacenterBits = 5             // 数据中心 ID 位数（0-31）
	machineIDBits  = 5             // 机器 ID 位数（0-31）
	sequenceBits   = 12            // 序列号位数（每毫秒最多生成 4096 个 ID）

	maxDatacenterID = 1<<datacenterBits - 1
	maxMachineID    = 1<<machineIDBits - 1
	maxSequence     = 1<<sequenceBits - 1

	timestampShift  = datacenterBits + machineIDBits + sequenceBits
	datacenterShift = machineIDBits + sequenceBits
	machineShift    = sequenceBits
)

type snowflake struct {
	mu            sync.Mutex
	lastTimestamp int64
	datacenterID  int64
	machineID     int64
	sequence      int64
}

// NewSnowflake 根据 datacenterID, machineID，返回一个 Snowflake 对象，
// datacenterID 和 machineID 分别代表数据中心 ID 和机器 ID，均可在 0-31 之间选择。如果超过范围，返回错误。
//
// NewSnowflake returns a Snowflake object according to datacenterID and machineID,
// datacenterID and machineID represent the data center ID and machine ID, respectively,
// both of which can be selected between 0 and 31. If out of range, an error is returned.
func newSnowflake(datacenterID, machineID int64) (*snowflake, error) {
	if datacenterID < 0 || datacenterID > maxDatacenterID {
		return nil, errors.New("invalid datacenter ID")
	}
	if machineID < 0 || machineID > maxMachineID {
		return nil, errors.New("invalid machine ID")
	}
	return &snowflake{
		// lastTimestamp: time.Now().UnixNano() / 1e6,
		datacenterID: datacenterID,
		machineID:    machineID,
	}, nil
}

func (s *snowflake) Generate() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current := time.Now().UnixNano() / 1e6 // 当前毫秒时间戳
	if current < s.lastTimestamp {
		return 0, errors.New("time is moving backwards")
	}

	if current == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			current = s.waitNextMillis(current)
		}
	} else {
		s.sequence = 0
	}

	s.lastTimestamp = current

	id := (current-epoch)<<timestampShift |
		(s.datacenterID << datacenterShift) |
		(s.machineID << machineShift) |
		s.sequence

	return id, nil
}

func (s *snowflake) waitNextMillis(current int64) int64 {
	for current <= s.lastTimestamp {
		current = time.Now().UnixNano() / 1e6
	}
	return current
}
