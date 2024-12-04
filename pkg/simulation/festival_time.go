// festival_time.go
// Create this file in the pkg/simulation directory

package simulation

import (
	"sync"
	"time"
)

type FestivalTime struct {
	startTime   time.Time
	currentTime time.Time
	timeScale   float64 // How much faster simulation time moves compared to real time
	isRunning   bool
	mu          sync.RWMutex

	// Festival schedule
	gateOpenTime  time.Time
	gateCloseTime time.Time
	eventEndTime  time.Time
}

func NewFestivalTime() *FestivalTime {
	now := time.Now()
	return &FestivalTime{
		startTime:     now,
		currentTime:   now,
		timeScale:     60.0, // 1 second real time = 1 minute simulation time
		isRunning:     false,
		gateOpenTime:  now.Add(30 * time.Minute), // Gates open 30 minutes after sim start
		gateCloseTime: now.Add(4 * time.Hour),    // Gates close after 4 hours
		eventEndTime:  now.Add(6 * time.Hour),    // Event ends after 6 hours
	}
}

func (ft *FestivalTime) Start() {
	ft.mu.Lock()
	if !ft.isRunning {
		ft.isRunning = true
		go ft.updateTime()
	}
	ft.mu.Unlock()
}

func (ft *FestivalTime) Stop() {
	ft.mu.Lock()
	ft.isRunning = false
	ft.mu.Unlock()
}

func (ft *FestivalTime) updateTime() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ft.mu.Lock()
		if !ft.isRunning {
			ft.mu.Unlock()
			return
		}

		// Update simulation time based on time scale
		ft.currentTime = ft.currentTime.Add(time.Duration(ft.timeScale * float64(time.Second)))
		ft.mu.Unlock()
	}
}

func (ft *FestivalTime) GetCurrentTime() time.Time {
	ft.mu.RLock()
	defer ft.mu.RUnlock()
	return ft.currentTime
}

func (ft *FestivalTime) IsGateOpen() bool {
	ft.mu.RLock()
	defer ft.mu.RUnlock()
	return ft.currentTime.After(ft.gateOpenTime) && ft.currentTime.Before(ft.gateCloseTime)
}

func (ft *FestivalTime) IsEventEnded() bool {
	ft.mu.RLock()
	defer ft.mu.RUnlock()
	return ft.currentTime.After(ft.eventEndTime)
}

func (ft *FestivalTime) GetTimeScale() float64 {
	ft.mu.RLock()
	defer ft.mu.RUnlock()
	return ft.timeScale
}

func (ft *FestivalTime) SetTimeScale(scale float64) {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	if scale > 0 {
		ft.timeScale = scale
	}
}

func (ft *FestivalTime) GetElapsedTime() time.Duration {
	ft.mu.RLock()
	defer ft.mu.RUnlock()
	return ft.currentTime.Sub(ft.startTime)
}

func (ft *FestivalTime) GetRemainingTime() time.Duration {
	ft.mu.RLock()
	defer ft.mu.RUnlock()
	return ft.eventEndTime.Sub(ft.currentTime)
}

func (ft *FestivalTime) GetPhase() string {
	ft.mu.RLock()
	defer ft.mu.RUnlock()

	if ft.currentTime.Before(ft.gateOpenTime) {
		return "setup"
	} else if ft.currentTime.Before(ft.gateCloseTime) {
		return "active"
	} else if ft.currentTime.Before(ft.eventEndTime) {
		return "closing"
	}
	return "ended"
}
