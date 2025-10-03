use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::Mutex;
use tokio::time::sleep;

/// Token bucket rate limiter.
pub struct RateLimiter {
    inner: Arc<Mutex<RateLimiterInner>>,
}

struct RateLimiterInner {
    rate: f64,           // Tokens per second
    burst: f64,          // Max tokens in bucket
    tokens: f64,         // Current tokens
    last_update: Instant, // Last token update time
}

impl RateLimiter {
    /// Creates a new rate limiter.
    /// rate: tokens per second
    /// burst: maximum tokens that can be accumulated
    pub fn new(rate: usize, burst: usize) -> Self {
        let burst = burst.max(rate);

        RateLimiter {
            inner: Arc::new(Mutex::new(RateLimiterInner {
                rate: rate as f64,
                burst: burst as f64,
                tokens: burst as f64,
                last_update: Instant::now(),
            })),
        }
    }

    /// Checks if a request can proceed without blocking.
    pub async fn allow(&self) -> bool {
        self.allow_n(1).await
    }

    /// Checks if n requests can proceed without blocking.
    pub async fn allow_n(&self, n: usize) -> bool {
        let mut inner = self.inner.lock().await;
        inner.refill_tokens();

        if inner.tokens >= n as f64 {
            inner.tokens -= n as f64;
            true
        } else {
            false
        }
    }

    /// Waits until a request can proceed.
    pub async fn wait(&self) {
        self.wait_n(1).await
    }

    /// Waits until n requests can proceed.
    pub async fn wait_n(&self, n: usize) {
        loop {
            let wait_time = {
                let mut inner = self.inner.lock().await;
                inner.refill_tokens();

                if inner.tokens >= n as f64 {
                    inner.tokens -= n as f64;
                    return;
                }

                // Calculate wait time
                let needed = n as f64 - inner.tokens;
                Duration::from_secs_f64(needed / inner.rate)
            };

            sleep(wait_time).await;
        }
    }

    /// Returns the current number of available tokens.
    pub async fn tokens(&self) -> f64 {
        let mut inner = self.inner.lock().await;
        inner.refill_tokens();
        inner.tokens
    }

    /// Returns the rate limit (tokens per second).
    pub async fn limit(&self) -> usize {
        let inner = self.inner.lock().await;
        inner.rate as usize
    }

    /// Returns the burst size.
    pub async fn burst(&self) -> usize {
        let inner = self.inner.lock().await;
        inner.burst as usize
    }
}

impl RateLimiterInner {
    fn refill_tokens(&mut self) {
        let now = Instant::now();
        let elapsed = now.duration_since(self.last_update);
        self.last_update = now;

        let tokens_to_add = elapsed.as_secs_f64() * self.rate;
        self.tokens = (self.tokens + tokens_to_add).min(self.burst);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_rate_limiter_allow() {
        let limiter = RateLimiter::new(10, 10);

        // Should allow up to burst
        for _ in 0..10 {
            assert!(limiter.allow().await);
        }

        // Should deny after burst
        assert!(!limiter.allow().await);
    }

    #[tokio::test]
    async fn test_rate_limiter_refill() {
        let limiter = RateLimiter::new(10, 10);

        // Consume all tokens
        for _ in 0..10 {
            assert!(limiter.allow().await);
        }

        // Wait for refill (100ms should add 1 token at 10/sec)
        tokio::time::sleep(Duration::from_millis(100)).await;
        assert!(limiter.allow().await);
    }
}
