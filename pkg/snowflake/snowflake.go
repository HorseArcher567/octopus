package snowflake

import (
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var defaultIDMaker = NewSnowflake()

func NextID() int64 {
	return defaultIDMaker.NextId()
}

// These constants are the bit lengths of Snowflake id parts.
const (
	BitLenTime      = 39                               // bit length of time
	BitLenSequence  = 8                                // bit length of sequence number
	BitLenMachineID = 63 - BitLenTime - BitLenSequence // bit length of machine id
)

// Snowflake is a distributed unique id generator.
type Snowflake struct {
	mutex       sync.Mutex
	machineId   int16
	startTime   int64
	elapsedTime int64
	sequence    int16
}

type MachineID func() int16

// sequenceRefreshPeriod is sequence refresh period, 1e7 means 10ms refresh sequence.
const sequenceRefreshPeriod = 1e7

var startTime = time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC)

func NewSnowflake() *Snowflake {
	snowflake := &Snowflake{
		startTime: startTime.UnixNano() / sequenceRefreshPeriod,
		machineId: KubeCIDRMachineID(),
	}

	return snowflake
}

func (snowflake *Snowflake) NextId() int64 {
	const maskSequence = int16(1<<BitLenSequence - 1)

	snowflake.mutex.Lock()
	defer snowflake.mutex.Unlock()

	curElapsedTime := currentElapsedTime(snowflake.startTime)
	if snowflake.elapsedTime < curElapsedTime {
		snowflake.elapsedTime = curElapsedTime
		snowflake.sequence = 0
	} else {
		snowflake.sequence = (snowflake.sequence + 1) & maskSequence
		if snowflake.sequence == 0 {
			snowflake.elapsedTime++
			overtime := snowflake.elapsedTime - curElapsedTime
			time.Sleep(sleepTime(overtime))
		}
	}

	return snowflake.toID()
}

func (snowflake *Snowflake) toID() int64 {
	return snowflake.elapsedTime<<(BitLenSequence+BitLenMachineID) |
		int64(snowflake.sequence)<<BitLenMachineID |
		int64(snowflake.machineId)
}

func toSequenceRefreshTime(t time.Time) int64 {
	return t.UnixNano() / sequenceRefreshPeriod
}

func currentElapsedTime(startTime int64) int64 {
	return toSequenceRefreshTime(time.Now()) - startTime
}

func sleepTime(overtime int64) time.Duration {
	return time.Duration(overtime)*10*time.Millisecond -
		time.Duration(time.Now().UnixNano()%sequenceRefreshPeriod)*time.Nanosecond
}

// KubeCIDRMachineID base CIDR network prefix length.
// CIDR prefix length must greater than 16.
func KubeCIDRMachineID() int16 {
	podIP := os.Getenv("POD_IP")
	if ip := net.ParseIP(podIP); ip != nil {
		return lower16BitPrivateIP(ip)
	}

	log.Fatalln("env ${POD_IP} error", podIP)
	return 0
}

func lower16BitPrivateIP(ip net.IP) int16 {
	return int16(ip[2])<<8 + int16(ip[3])
}
