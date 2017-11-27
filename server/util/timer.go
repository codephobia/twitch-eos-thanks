package util

import (
    "time"
)

type Timer struct {
    timer    *time.Timer

    duration time.Duration
    end      time.Time
    again    bool
    
    callback func()
}

// create new timer
func NewTimer(t time.Duration, again bool, callback func()) *Timer {
    timer := &Timer{
        timer: time.NewTimer(t),
        duration: t,
        end: time.Now().Add(t),
        again: again,
        callback: callback,
    }
    
    go func() {
        <-timer.timer.C
        timer.callback()
        
        if timer.again {
            timer.Reset()
        }
    }()
    
    return timer
}

// reset timer
func (t *Timer) Reset() {
    t.timer = time.NewTimer(t.duration)
    t.end = time.Now().Add(t.duration)
    
    go func() {
        <-t.timer.C
        t.callback()
        
        if t.again {
            t.Reset()
        }
    }()
}

// stop timer
func (t *Timer) Stop() bool {
    stop := t.timer.Stop()
    if stop {
        t.end = time.Now()
    }
    return stop
}