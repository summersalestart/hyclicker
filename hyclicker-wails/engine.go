package main

import (
	"math/rand"
	"sync"
	"time"
)

const (
	MOUSEEVENTF_LEFTDOWN  = 0x0002
	MOUSEEVENTF_LEFTUP    = 0x0004
	MOUSEEVENTF_RIGHTDOWN = 0x0008
	MOUSEEVENTF_RIGHTUP   = 0x0010
	MOUSEEVENTF_MIDDOWN   = 0x0020
	MOUSEEVENTF_MIDDLEUP  = 0x0040
)

type MOUSEINPUT struct {
	Dx, Dy      int32
	MouseData   uint32
	DwFlags     uint32
	DwTime      uint32
	DwExtraInfo uintptr
}

type INPUT struct {
	Type  uint32
	Mouse MOUSEINPUT
}

type ClickEngine struct {
	mu        sync.Mutex
	running   bool
	stop      chan struct{}
	minCps    float64
	maxCps    float64
	button    string
	onState   func(bool)
	lastPhase float64
}

func NewClickEngine() *ClickEngine { return &ClickEngine{stop: make(chan struct{})} }

func (e *ClickEngine) Configure(minCps, maxCps float64, button string) {
	if minCps <= 0 {
		minCps = 0.1
	}
	if maxCps < minCps {
		maxCps = minCps
	}
	e.mu.Lock()
	e.minCps, e.maxCps, e.button = minCps, maxCps, button
	e.mu.Unlock()
}

func (e *ClickEngine) OnStateChange(fn func(bool)) { e.onState = fn }

func (e *ClickEngine) IsRunning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.running
}

func (e *ClickEngine) Start() {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}
	e.running = true
	e.stop = make(chan struct{})
	go e.loop(e.stop)
	e.mu.Unlock()
	e.notify(true)
}

func (e *ClickEngine) Stop() {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return
	}
	e.running = false
	close(e.stop)
	e.mu.Unlock()
	e.notify(false)
}

func (e *ClickEngine) notify(r bool) {
	if e.onState != nil {
		go e.onState(r)
	}
}

func pickPhaseCps(min, max, avoid float64) float64 {
	v := min + rand.Float64()*(max-min)
	span := max - min
	if span > 0.5 {
		for i := 0; i < 5; i++ {
			if abs(v-avoid) >= span*0.15 {
				break
			}
			v = min + rand.Float64()*(max-min)
		}
	}
	return v
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func (e *ClickEngine) loop(stop chan struct{}) {
	time.Sleep(350 * time.Millisecond) // grace: don't click the Start button itself
	var targetCps, phaseEnd float64
	now := func() float64 { return float64(time.Now().UnixNano()) / 1e6 }

	for {
		select {
		case <-stop:
			return
		default:
		}

		e.mu.Lock()
		minCps, maxCps, button := e.minCps, e.maxCps, e.button
		e.mu.Unlock()

		t := now()
		if targetCps == 0 || t >= phaseEnd {
			targetCps = pickPhaseCps(minCps, maxCps, e.lastPhase)
			e.lastPhase = targetCps
			phaseEnd = t + 500 + rand.Float64()*1500
		}
		ideal := 1000 / max(targetCps, 0.1)
		jitter := 1 + (rand.Float64()*0.3 - 0.15)
		delay := clamp(ideal*jitter, 30, 500)

		clickOnce(button)

		select {
		case <-stop:
			return
		case <-time.After(time.Duration(delay) * time.Millisecond):
		}
	}
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func clickOnce(button string) {
	var down, up uint32 = MOUSEEVENTF_LEFTDOWN, MOUSEEVENTF_LEFTUP
	switch button {
	case "Right":
		down, up = MOUSEEVENTF_RIGHTDOWN, MOUSEEVENTF_RIGHTUP
	case "Middle":
		down, up = MOUSEEVENTF_MIDDOWN, MOUSEEVENTF_MIDDLEUP
	}
	inputs := []INPUT{
		{Type: 0, Mouse: MOUSEINPUT{DwFlags: down}},
		{Type: 0, Mouse: MOUSEINPUT{DwFlags: up}},
	}
	_ = sendInputs(inputs)
}
