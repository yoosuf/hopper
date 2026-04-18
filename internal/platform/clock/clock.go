package clock

import (
	"time"
)

// Clock provides time-related functionality
type Clock interface {
	Now() time.Time
	UTC() time.Time
	NowInTimezone(loc *time.Location) time.Time
}

// realClock implements Clock using the system time
type realClock struct{}

// New creates a new Clock using real time
func New() Clock {
	return &realClock{}
}

// Now returns the current time in UTC
func (c *realClock) Now() time.Time {
	return time.Now().UTC()
}

// UTC returns the current time in UTC
func (c *realClock) UTC() time.Time {
	return time.Now().UTC()
}

// NowInTimezone returns the current time in the specified timezone
func (c *realClock) NowInTimezone(loc *time.Location) time.Time {
	return time.Now().In(loc)
}

// NewFixedClock creates a Clock that always returns a fixed time (useful for testing)
func NewFixedClock(t time.Time) Clock {
	return &fixedClock{time: t}
}

// fixedClock implements Clock with a fixed time
type fixedClock struct {
	time time.Time
}

func (c *fixedClock) Now() time.Time {
	return c.time
}

func (c *fixedClock) UTC() time.Time {
	return c.time.UTC()
}

func (c *fixedClock) NowInTimezone(loc *time.Location) time.Time {
	return c.time.In(loc)
}
