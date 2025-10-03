use thiserror::Error;

/// Common error types for the B25 trading system.
#[derive(Error, Debug)]
pub enum B25Error {
    #[error("Invalid decimal value: {0}")]
    InvalidDecimal(String),

    #[error("Invalid timestamp: {0}")]
    InvalidTimestamp(String),

    #[error("Order validation failed: {0}")]
    OrderValidationFailed(String),

    #[error("Insufficient balance: required {required}, available {available}")]
    InsufficientBalance { required: String, available: String },

    #[error("Position limit exceeded: {0}")]
    PositionLimitExceeded(String),

    #[error("Rate limit exceeded")]
    RateLimitExceeded,

    #[error("Circuit breaker open")]
    CircuitBreakerOpen,

    #[error("Exchange error: {0}")]
    ExchangeError(String),

    #[error("WebSocket error: {0}")]
    WebSocketError(String),

    #[error("Network error: {0}")]
    NetworkError(String),

    #[error("Serialization error: {0}")]
    SerializationError(String),

    #[error("Configuration error: {0}")]
    ConfigError(String),

    #[error("Database error: {0}")]
    DatabaseError(String),

    #[error("Not found: {0}")]
    NotFound(String),

    #[error("Internal error: {0}")]
    InternalError(String),
}

/// Result type alias using B25Error.
pub type B25Result<T> = Result<T, B25Error>;

impl From<rust_decimal::Error> for B25Error {
    fn from(err: rust_decimal::Error) -> Self {
        B25Error::InvalidDecimal(err.to_string())
    }
}

impl From<serde_json::Error> for B25Error {
    fn from(err: serde_json::Error) -> Self {
        B25Error::SerializationError(err.to_string())
    }
}

impl From<bincode::Error> for B25Error {
    fn from(err: bincode::Error) -> Self {
        B25Error::SerializationError(err.to_string())
    }
}
