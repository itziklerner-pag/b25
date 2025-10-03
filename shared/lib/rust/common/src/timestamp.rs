use chrono::{DateTime, NaiveDateTime, Utc};
use serde::{Deserialize, Serialize};
use std::fmt;
use std::time::{Duration, SystemTime, UNIX_EPOCH};

/// High-precision timestamp with nanosecond accuracy.
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize)]
pub struct Timestamp {
    pub seconds: i64,
    pub nanos: i32,
}

impl Timestamp {
    /// Creates a new timestamp from seconds and nanoseconds.
    pub fn new(seconds: i64, nanos: i32) -> Self {
        Timestamp { seconds, nanos }
    }

    /// Returns the current timestamp.
    pub fn now() -> Self {
        let now = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .expect("Time went backwards");

        Timestamp {
            seconds: now.as_secs() as i64,
            nanos: now.subsec_nanos() as i32,
        }
    }

    /// Creates a timestamp from a SystemTime.
    pub fn from_system_time(time: SystemTime) -> Self {
        let duration = time
            .duration_since(UNIX_EPOCH)
            .expect("Time went backwards");

        Timestamp {
            seconds: duration.as_secs() as i64,
            nanos: duration.subsec_nanos() as i32,
        }
    }

    /// Creates a timestamp from a DateTime<Utc>.
    pub fn from_datetime(dt: DateTime<Utc>) -> Self {
        Timestamp {
            seconds: dt.timestamp(),
            nanos: dt.timestamp_subsec_nanos() as i32,
        }
    }

    /// Converts the timestamp to a SystemTime.
    pub fn to_system_time(&self) -> SystemTime {
        UNIX_EPOCH + Duration::new(self.seconds as u64, self.nanos as u32)
    }

    /// Converts the timestamp to a DateTime<Utc>.
    pub fn to_datetime(&self) -> DateTime<Utc> {
        DateTime::<Utc>::from_utc(
            NaiveDateTime::from_timestamp_opt(self.seconds, self.nanos as u32)
                .expect("Invalid timestamp"),
            Utc,
        )
    }

    /// Returns the Unix timestamp in seconds.
    pub fn unix(&self) -> i64 {
        self.seconds
    }

    /// Returns the Unix timestamp in nanoseconds.
    pub fn unix_nano(&self) -> i64 {
        self.seconds * 1_000_000_000 + self.nanos as i64
    }

    /// Returns the Unix timestamp in microseconds.
    pub fn unix_micro(&self) -> i64 {
        self.seconds * 1_000_000 + (self.nanos as i64) / 1_000
    }

    /// Returns the Unix timestamp in milliseconds.
    pub fn unix_milli(&self) -> i64 {
        self.seconds * 1_000 + (self.nanos as i64) / 1_000_000
    }

    /// Returns true if this timestamp is before the other.
    pub fn before(&self, other: &Timestamp) -> bool {
        self < other
    }

    /// Returns true if this timestamp is after the other.
    pub fn after(&self, other: &Timestamp) -> bool {
        self > other
    }

    /// Returns the duration between this timestamp and the other.
    pub fn duration_since(&self, other: &Timestamp) -> Duration {
        let secs = (self.seconds - other.seconds) as u64;
        let nanos = (self.nanos - other.nanos) as u32;
        Duration::new(secs, nanos)
    }

    /// Adds a duration to this timestamp.
    pub fn add_duration(&self, duration: Duration) -> Timestamp {
        let total_nanos = self.nanos as i64 + duration.subsec_nanos() as i64;
        let seconds = self.seconds + duration.as_secs() as i64 + total_nanos / 1_000_000_000;
        let nanos = (total_nanos % 1_000_000_000) as i32;

        Timestamp { seconds, nanos }
    }

    /// Subtracts a duration from this timestamp.
    pub fn sub_duration(&self, duration: Duration) -> Timestamp {
        let total_nanos = self.nanos as i64 - duration.subsec_nanos() as i64;
        let mut seconds = self.seconds - duration.as_secs() as i64;
        let mut nanos = total_nanos;

        if nanos < 0 {
            seconds -= 1;
            nanos += 1_000_000_000;
        }

        Timestamp {
            seconds,
            nanos: nanos as i32,
        }
    }

    /// Returns true if the timestamp is zero.
    pub fn is_zero(&self) -> bool {
        self.seconds == 0 && self.nanos == 0
    }
}

impl fmt::Display for Timestamp {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.to_datetime().to_rfc3339())
    }
}

impl From<SystemTime> for Timestamp {
    fn from(time: SystemTime) -> Self {
        Timestamp::from_system_time(time)
    }
}

impl From<DateTime<Utc>> for Timestamp {
    fn from(dt: DateTime<Utc>) -> Self {
        Timestamp::from_datetime(dt)
    }
}

impl From<Timestamp> for SystemTime {
    fn from(ts: Timestamp) -> Self {
        ts.to_system_time()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_timestamp_now() {
        let ts = Timestamp::now();
        assert!(ts.seconds > 0);
        assert!(ts.nanos >= 0 && ts.nanos < 1_000_000_000);
    }

    #[test]
    fn test_timestamp_conversions() {
        let ts = Timestamp::new(1234567890, 123456789);
        assert_eq!(ts.unix(), 1234567890);
        assert_eq!(ts.unix_milli(), 1234567890123);
        assert_eq!(ts.unix_micro(), 1234567890123456);
    }
}
