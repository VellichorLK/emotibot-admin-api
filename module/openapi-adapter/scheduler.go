package main

import (
	"log"
	"time"
)

type scheduler struct {
	Task   func() error
	timer  *time.Timer
	doneCh chan struct{}
}

type Clock interface {
	NextRun() time.Duration
}

type daily struct{}

func (c *daily) NextRun() time.Duration {
	now := time.Now()
	nextday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
	return nextday.Sub(now)
}

func NewScheduler(task func() error) *scheduler {
	return &scheduler{
		Task:   task,
		doneCh: make(chan struct{}),
	}
}

func (s *scheduler) Start(c Clock) {
	go func() {
		for {
			s.timer = time.NewTimer(c.NextRun())
			select {
			case <-s.timer.C:
				err := s.Task()
				if err != nil {
					log.Println("task failed, ", err)
				}
				s.timer.Reset(c.NextRun())
			case <-s.doneCh:
				s.timer.Stop()
				return
			}
		}
	}()

}

func (s *scheduler) Stop() bool {
	s.doneCh <- struct{}{}
	return true
}
