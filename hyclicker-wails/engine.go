package main

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

type MouseButton string

const (
	ButtonLeft   MouseButton = "Left"
	ButtonRight  MouseButton = "Right"
	ButtonMiddle MouseButton = "Middle"
)

type buttonPair struct{ down, up uint32 }

var buttonPairs = map[MouseButton]buttonPair{
	ButtonLeft:   {down: 0x0002, up: 0x0004},
	ButtonRight:  {down: 0x0008, up: 0x0010},
	ButtonMiddle: {down: 0x0020, up: 0x0040},
}

func resolveButton(name MouseButton) buttonPair {
	if p, ok := buttonPairs[name]; ok {
		return p
	}
	return buttonPairs[ButtonLeft]
}

type engineConfig struct {
	minCps float64
	maxCps float64
	button MouseButton
	rocket bool
}

type ClickEngine struct {
	mu        sync.Mutex
	running   bool
	stop      chan struct{}
	done      chan struct{}
	cfg       engineConfig
	stateFn   func(bool)
	lastPhase float64
}

func NewClickEngine() *ClickEngine {
	return &ClickEngine{
		cfg: engineConfig{minCps: 8, maxCps: 12, button: ButtonLeft},
	}
}

var ErrEngineRunning = errors.New("engine already running")

func (e *ClickEngine) SetConfig(minCps, maxCps float64, button MouseButton, rocket bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.cfg = engineConfig{
		minCps: clampFloat(minCps, 0.1, maxCps),
		maxCps: clampFloat(maxCps, 0.1, maxCps),
		button: button,
		rocket: rocket,
	}
}

func (e *ClickEngine) current() engineConfig {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.cfg
}

func (e *ClickEngine) CurrentConfig() engineConfig {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.cfg
}

func (e *ClickEngine) OnStateChange(fn func(bool)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.stateFn = fn
}

func (e *ClickEngine) Running() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.running
}

func (e *ClickEngine) Start() error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return ErrEngineRunning
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	e.stop, e.done, e.running = stop, done, true
	e.mu.Unlock()

	go func() {
		runLoop(stop, e, &e.lastPhase)
		close(done)
		e.mu.Lock()
		e.running = false
		e.mu.Unlock()
		e.emit(false)
	}()
	e.emit(true)
	return nil
}

func (e *ClickEngine) Stop() {
	e.mu.Lock()
	stop, done, running := e.stop, e.done, e.running
	e.mu.Unlock()
	if !running {
		return
	}
	close(stop)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}
}

func (e *ClickEngine) emit(running bool) {
	e.mu.Lock()
	fn := e.stateFn
	e.mu.Unlock()
	if fn != nil {
		go fn(running)
	}
}

const (
	phaseMinMillis   = 500.0
	phaseMaxMillis   = 2000.0
	delayMinMillis   = 30.0
	delayMaxMillis   = 500.0
	jitterFraction   = 0.15
	startDelayMillis = 350.0
)

func runLoop(stop <-chan struct{}, e *ClickEngine, lastPhase *float64) {
	cfg := e.current()
	if !cfg.rocket {
		select {
		case <-stop:
			return
		case <-time.After(time.Duration(startDelayMillis) * time.Millisecond):
		}
	}

	var targetCps, phaseDeadline float64
	for {
		cfg = e.current()
		now := monotonicMillis()
		if cfg.rocket {
			targetCps = cfg.maxCps
		} else if targetCps == 0 || now >= phaseDeadline {
			targetCps = nextPhaseCps(cfg.minCps, cfg.maxCps, *lastPhase)
			*lastPhase = targetCps
			phaseDeadline = now + (phaseMinMillis + rand.Float64()*(phaseMaxMillis-phaseMinMillis))
		}
		sendClickPair(resolveButton(cfg.button))

		select {
		case <-stop:
			return
		case <-time.After(clickDelay(cfg, targetCps)):
		}
	}
}

func monotonicMillis() float64 {
	return float64(time.Now().UnixNano()) / 1e6
}

func nextPhaseCps(min, max, avoid float64) float64 {
	span := max - min
	v := min + rand.Float64()*span
	if span > 0.5 {
		for i := 0; i < 5; i++ {
			if abs(v-avoid) >= span*0.15 {
				break
			}
			v = min + rand.Float64()*span
		}
	}
	return v
}

func clickDelay(cfg engineConfig, targetCps float64) time.Duration {
	ideal := 1000 / clampFloat(targetCps, 0.1, 1e6)
	if cfg.rocket {
		return time.Duration(clampFloat(ideal, 2, delayMaxMillis)) * time.Millisecond
	}
	jitter := 1 + (rand.Float64()*2*jitterFraction - jitterFraction)
	ms := clampFloat(ideal*jitter, delayMinMillis, delayMaxMillis)
	return time.Duration(ms) * time.Millisecond
}

func sendClickPair(p buttonPair) {
	sendInput(clickInput(p.down, p.up))
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
