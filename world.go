package main

import (
	"time"
)

type World struct {
	time     time.Time
	ticker   *time.Ticker
	stopChan chan bool
}

func NewWorld(startTime time.Time) *World {
	w := &World{
		time:     startTime,
		ticker:   time.NewTicker(time.Millisecond),
		stopChan: make(chan bool),
	}
	go w.startWorldTime()

	return w
}

func (w *World) startWorldTime() {
	for {
		select {
		case <-w.ticker.C:
			w.time = w.time.Add(time.Millisecond)
		case <-w.stopChan:
			w.ticker.Stop()
			return
		}
	}
}

func (w *World) stopWorldTime() {
	w.stopChan <- true
}
