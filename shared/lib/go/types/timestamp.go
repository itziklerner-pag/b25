package types

import (
	"time"
)

// Timestamp represents a high-precision timestamp with nanosecond accuracy.
type Timestamp struct {
	Seconds int64
	Nanos   int32
}

// Now returns the current timestamp.
func Now() Timestamp {
	t := time.Now()
	return Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}

// FromTime creates a Timestamp from a time.Time.
func FromTime(t time.Time) Timestamp {
	return Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}

// ToTime converts the Timestamp to a time.Time.
func (ts Timestamp) ToTime() time.Time {
	return time.Unix(ts.Seconds, int64(ts.Nanos))
}

// Unix returns the Unix timestamp in seconds.
func (ts Timestamp) Unix() int64 {
	return ts.Seconds
}

// UnixNano returns the Unix timestamp in nanoseconds.
func (ts Timestamp) UnixNano() int64 {
	return ts.Seconds*1e9 + int64(ts.Nanos)
}

// UnixMicro returns the Unix timestamp in microseconds.
func (ts Timestamp) UnixMicro() int64 {
	return ts.Seconds*1e6 + int64(ts.Nanos)/1e3
}

// UnixMilli returns the Unix timestamp in milliseconds.
func (ts Timestamp) UnixMilli() int64 {
	return ts.Seconds*1e3 + int64(ts.Nanos)/1e6
}

// Before returns true if ts is before other.
func (ts Timestamp) Before(other Timestamp) bool {
	if ts.Seconds < other.Seconds {
		return true
	}
	if ts.Seconds == other.Seconds {
		return ts.Nanos < other.Nanos
	}
	return false
}

// After returns true if ts is after other.
func (ts Timestamp) After(other Timestamp) bool {
	if ts.Seconds > other.Seconds {
		return true
	}
	if ts.Seconds == other.Seconds {
		return ts.Nanos > other.Nanos
	}
	return false
}

// Equal returns true if ts equals other.
func (ts Timestamp) Equal(other Timestamp) bool {
	return ts.Seconds == other.Seconds && ts.Nanos == other.Nanos
}

// Sub returns the duration ts - other.
func (ts Timestamp) Sub(other Timestamp) time.Duration {
	return ts.ToTime().Sub(other.ToTime())
}

// Add returns the timestamp ts + duration.
func (ts Timestamp) Add(d time.Duration) Timestamp {
	return FromTime(ts.ToTime().Add(d))
}

// String returns the string representation.
func (ts Timestamp) String() string {
	return ts.ToTime().Format(time.RFC3339Nano)
}

// IsZero returns true if the timestamp is the zero value.
func (ts Timestamp) IsZero() bool {
	return ts.Seconds == 0 && ts.Nanos == 0
}
