package ui

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SysInfo holds a snapshot of system resource usage for display.
type SysInfo struct {
	// CPUPercent is the CPU usage as a value in [0, 100].
	CPUPercent float64
	// MemUsedBytes is the current resident memory usage in bytes.
	MemUsedBytes uint64
	// MemTotalBytes is the total system memory in bytes.
	MemTotalBytes uint64
	// GPULabel is a human-readable GPU info string, or empty when unavailable.
	GPULabel string
	// ModelStatus is the current model load state string.
	ModelStatus string
	// QueueLen is the number of pending inference requests.
	QueueLen int
	// TokensPerSec is the latest measured inference throughput.
	TokensPerSec float64
}

// MemPercent returns used memory as a percentage of total, or 0.
func (s SysInfo) MemPercent() float64 {
	if s.MemTotalBytes == 0 {
		return 0
	}
	return float64(s.MemUsedBytes) / float64(s.MemTotalBytes) * 100
}

// HealthOK returns true when CPU < 90%, memory < 90%, and queue < 10.
func (s SysInfo) HealthOK() bool {
	return s.CPUPercent < 90 && s.MemPercent() < 90 && s.QueueLen < 10
}

// SysCollector samples system metrics in the background.
// Call Start to begin sampling, Stop to halt.
type SysCollector struct {
	mu       sync.RWMutex
	info     SysInfo
	interval time.Duration
	stop     chan struct{}
	prevIdle uint64
	prevTotal uint64
}

// NewSysCollector creates a collector that samples at the given interval.
func NewSysCollector(interval time.Duration) *SysCollector {
	return &SysCollector{
		interval: interval,
		stop:     make(chan struct{}),
	}
}

// Start launches the background sampling goroutine.
func (sc *SysCollector) Start() {
	go sc.loop()
}

// Stop halts the sampling goroutine.
func (sc *SysCollector) Stop() {
	close(sc.stop)
}

// Info returns the most recent sample.
func (sc *SysCollector) Info() SysInfo {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.info
}

// SetModelStatus updates the model status field.
func (sc *SysCollector) SetModelStatus(s string) {
	sc.mu.Lock()
	sc.info.ModelStatus = s
	sc.mu.Unlock()
}

// SetQueueLen updates the queue length field.
func (sc *SysCollector) SetQueueLen(n int) {
	sc.mu.Lock()
	sc.info.QueueLen = n
	sc.mu.Unlock()
}

// SetTokensPerSec updates the inference throughput field.
func (sc *SysCollector) SetTokensPerSec(tps float64) {
	sc.mu.Lock()
	sc.info.TokensPerSec = tps
	sc.mu.Unlock()
}

// loop periodically samples CPU and memory on Linux; no-ops on other OSes.
func (sc *SysCollector) loop() {
	ticker := time.NewTicker(sc.interval)
	defer ticker.Stop()
	for {
		select {
		case <-sc.stop:
			return
		case <-ticker.C:
			sc.sample()
		}
	}
}

// sample reads /proc/stat and /proc/meminfo (Linux only).
func (sc *SysCollector) sample() {
	if runtime.GOOS != "linux" {
		return
	}
	cpu := sc.sampleCPU()
	memUsed, memTotal := sampleMem()
	sc.mu.Lock()
	sc.info.CPUPercent = cpu
	sc.info.MemUsedBytes = memUsed
	sc.info.MemTotalBytes = memTotal
	sc.mu.Unlock()
}

// sampleCPU reads /proc/stat and returns the CPU usage delta as a percentage.
func (sc *SysCollector) sampleCPU() float64 {
	idle, total, err := readProcStat()
	if err != nil {
		return 0
	}
	idleDelta := idle - sc.prevIdle
	totalDelta := total - sc.prevTotal
	sc.prevIdle = idle
	sc.prevTotal = total
	if totalDelta == 0 {
		return 0
	}
	return (1 - float64(idleDelta)/float64(totalDelta)) * 100
}

// readProcStat parses the first cpu line of /proc/stat and returns (idle, total).
func readProcStat() (idle, total uint64, err error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			break
		}
		var vals [10]uint64
		for i := 1; i < len(fields) && i < 11; i++ {
			vals[i-1], _ = strconv.ParseUint(fields[i], 10, 64)
			total += vals[i-1]
		}
		idle = vals[3] // field 4 is idle
		return idle, total, nil
	}
	return 0, 0, fmt.Errorf("cpu line not found in /proc/stat")
}

// sampleMem parses /proc/meminfo and returns (used, total) in bytes.
func sampleMem() (used, total uint64) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	defer f.Close()

	var memTotal, memFree, memBuff, memCache uint64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		v, _ := strconv.ParseUint(fields[1], 10, 64)
		switch fields[0] {
		case "MemTotal:":
			memTotal = v * 1024
		case "MemFree:":
			memFree = v * 1024
		case "Buffers:":
			memBuff = v * 1024
		case "Cached:":
			memCache = v * 1024
		}
	}
	used = memTotal - memFree - memBuff - memCache
	return used, memTotal
}
