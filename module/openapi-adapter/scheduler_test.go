package main

import (
	"log"
	"testing"
	"time"
)

type mockupClock struct{}

func (m *mockupClock) NextRun() time.Duration {
	return time.Duration(1) * time.Second
}
func TestScheduler(t *testing.T) {
	var s = 0
	NewScheduler(func() error {
		s++
		log.Println("fired")
		return nil
	}).Start(&mockupClock{})
	time.Sleep(time.Duration(3500) * time.Millisecond)
	if s == 0 {
		t.Fatal("task has not been fired")
	} else if s < 3 {
		t.Fatal("task should have fired at least three time, but got ", s)
	}
}
