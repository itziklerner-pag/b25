use std::sync::atomic::{AtomicU64, Ordering};
use std::time::{SystemTime, UNIX_EPOCH};
use uuid::Uuid;

static SEQUENCE: AtomicU64 = AtomicU64::new(0);

/// Generates a unique order ID with prefix.
pub fn generate_order_id(prefix: &str) -> String {
    let seq = SEQUENCE.fetch_add(1, Ordering::Relaxed);
    let timestamp = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_nanos();
    format!("{}_{}", prefix, timestamp, seq)
}

/// Generates a unique request ID.
pub fn generate_request_id() -> String {
    Uuid::new_v4().to_string()
}

/// Generates a UUID.
pub fn generate_uuid() -> String {
    Uuid::new_v4().to_string()
}

/// Generates a client order ID for exchange submission.
pub fn generate_client_order_id(strategy_id: &str) -> String {
    let seq = SEQUENCE.fetch_add(1, Ordering::Relaxed);
    let timestamp = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_millis();
    format!("{}_{}_{}", strategy_id, timestamp, seq)
}

/// Generates a trace ID for distributed tracing.
pub fn generate_trace_id() -> String {
    format!("{:032x}", Uuid::new_v4().as_u128())
}

/// Generates a span ID for distributed tracing.
pub fn generate_span_id() -> String {
    format!("{:016x}", SEQUENCE.fetch_add(1, Ordering::Relaxed))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_generate_uuid() {
        let id1 = generate_uuid();
        let id2 = generate_uuid();
        assert_ne!(id1, id2);
        assert_eq!(id1.len(), 36); // UUID format
    }

    #[test]
    fn test_generate_client_order_id() {
        let id = generate_client_order_id("strategy1");
        assert!(id.starts_with("strategy1_"));
    }
}
