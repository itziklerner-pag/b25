use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::RwLock;

/// Circuit breaker state.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CircuitBreakerState {
    Closed,
    Open,
    HalfOpen,
}

/// Circuit breaker configuration.
#[derive(Debug, Clone)]
pub struct CircuitBreakerConfig {
    pub max_failures: usize,
    pub timeout: Duration,
    pub half_open_max_requests: usize,
}

impl Default for CircuitBreakerConfig {
    fn default() -> Self {
        CircuitBreakerConfig {
            max_failures: 5,
            timeout: Duration::from_secs(30),
            half_open_max_requests: 3,
        }
    }
}

/// Circuit breaker implementation.
pub struct CircuitBreaker {
    config: CircuitBreakerConfig,
    state: Arc<RwLock<CircuitBreakerInner>>,
}

struct CircuitBreakerInner {
    state: CircuitBreakerState,
    failures: usize,
    successes: usize,
    last_fail_time: Option<Instant>,
    half_open_requests: usize,
}

impl CircuitBreaker {
    /// Creates a new circuit breaker.
    pub fn new(config: CircuitBreakerConfig) -> Self {
        CircuitBreaker {
            config,
            state: Arc::new(RwLock::new(CircuitBreakerInner {
                state: CircuitBreakerState::Closed,
                failures: 0,
                successes: 0,
                last_fail_time: None,
                half_open_requests: 0,
            })),
        }
    }

    /// Executes a function with circuit breaker protection.
    pub async fn execute<F, T, E>(&self, f: F) -> Result<T, CircuitBreakerError<E>>
    where
        F: FnOnce() -> Result<T, E>,
    {
        self.before_request().await?;

        let result = f();

        self.after_request(result.is_ok()).await;

        result.map_err(CircuitBreakerError::Inner)
    }

    /// Executes an async function with circuit breaker protection.
    pub async fn execute_async<F, Fut, T, E>(&self, f: F) -> Result<T, CircuitBreakerError<E>>
    where
        F: FnOnce() -> Fut,
        Fut: std::future::Future<Output = Result<T, E>>,
    {
        self.before_request().await?;

        let result = f().await;

        self.after_request(result.is_ok()).await;

        result.map_err(CircuitBreakerError::Inner)
    }

    /// Returns the current state.
    pub async fn get_state(&self) -> CircuitBreakerState {
        self.state.read().await.state
    }

    /// Resets the circuit breaker to closed state.
    pub async fn reset(&self) {
        let mut inner = self.state.write().await;
        inner.state = CircuitBreakerState::Closed;
        inner.failures = 0;
        inner.successes = 0;
        inner.half_open_requests = 0;
    }

    async fn before_request(&self) -> Result<(), CircuitBreakerError<()>> {
        let mut inner = self.state.write().await;

        match inner.state {
            CircuitBreakerState::Closed => Ok(()),

            CircuitBreakerState::Open => {
                if let Some(last_fail) = inner.last_fail_time {
                    if last_fail.elapsed() > self.config.timeout {
                        inner.state = CircuitBreakerState::HalfOpen;
                        inner.half_open_requests = 1;
                        Ok(())
                    } else {
                        Err(CircuitBreakerError::Open)
                    }
                } else {
                    Err(CircuitBreakerError::Open)
                }
            }

            CircuitBreakerState::HalfOpen => {
                if inner.half_open_requests >= self.config.half_open_max_requests {
                    Err(CircuitBreakerError::TooManyRequests)
                } else {
                    inner.half_open_requests += 1;
                    Ok(())
                }
            }
        }
    }

    async fn after_request(&self, success: bool) {
        let mut inner = self.state.write().await;

        if success {
            self.on_success(&mut inner);
        } else {
            self.on_failure(&mut inner);
        }
    }

    fn on_success(&self, inner: &mut CircuitBreakerInner) {
        match inner.state {
            CircuitBreakerState::Closed => {
                inner.failures = 0;
            }

            CircuitBreakerState::HalfOpen => {
                inner.successes += 1;
                if inner.successes >= self.config.half_open_max_requests {
                    inner.state = CircuitBreakerState::Closed;
                    inner.failures = 0;
                    inner.successes = 0;
                }
            }

            _ => {}
        }
    }

    fn on_failure(&self, inner: &mut CircuitBreakerInner) {
        inner.failures += 1;
        inner.last_fail_time = Some(Instant::now());

        match inner.state {
            CircuitBreakerState::Closed => {
                if inner.failures >= self.config.max_failures {
                    inner.state = CircuitBreakerState::Open;
                }
            }

            CircuitBreakerState::HalfOpen => {
                inner.state = CircuitBreakerState::Open;
            }

            _ => {}
        }
    }
}

/// Circuit breaker error types.
#[derive(Debug, thiserror::Error)]
pub enum CircuitBreakerError<E> {
    #[error("Circuit breaker is open")]
    Open,

    #[error("Too many requests")]
    TooManyRequests,

    #[error("Inner error: {0}")]
    Inner(E),
}
