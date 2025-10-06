use axum::{
    extract::State,
    http::{StatusCode, HeaderMap, header::{ACCESS_CONTROL_ALLOW_ORIGIN, ACCESS_CONTROL_ALLOW_METHODS, ACCESS_CONTROL_ALLOW_HEADERS, CONTENT_TYPE}},
    response::{IntoResponse, Response},
    routing::get,
    Json, Router,
};
use serde_json::json;
use std::sync::Arc;
use tracing::info;

use crate::metrics;

pub struct HealthServer {
    port: u16,
}

impl HealthServer {
    pub fn new(port: u16) -> Self {
        Self { port }
    }

    pub async fn start(self) -> Result<(), Box<dyn std::error::Error>> {
        let app = Router::new()
            .route("/health", get(health_handler))
            .route("/metrics", get(metrics_handler))
            .route("/ready", get(readiness_handler));

        let addr = format!("0.0.0.0:{}", self.port);
        info!("Health server listening on {}", addr);

        let listener = tokio::net::TcpListener::bind(&addr).await?;
        axum::serve(listener, app).await?;

        Ok(())
    }
}

// Helper function to add CORS headers
fn add_cors_headers(mut headers: HeaderMap) -> HeaderMap {
    headers.insert(ACCESS_CONTROL_ALLOW_ORIGIN, "*".parse().unwrap());
    headers.insert(ACCESS_CONTROL_ALLOW_METHODS, "GET, OPTIONS".parse().unwrap());
    headers.insert(ACCESS_CONTROL_ALLOW_HEADERS, "Content-Type".parse().unwrap());
    headers
}

async fn health_handler() -> impl IntoResponse {
    let mut headers = HeaderMap::new();
    headers = add_cors_headers(headers);

    (
        headers,
        Json(json!({
            "status": "healthy",
            "service": "market-data",
            "version": env!("CARGO_PKG_VERSION"),
        }))
    )
}

async fn readiness_handler() -> impl IntoResponse {
    let mut headers = HeaderMap::new();
    headers = add_cors_headers(headers);

    // TODO: Check Redis connection, WebSocket status, etc.
    (
        headers,
        Json(json!({
            "status": "ready",
        }))
    )
}

async fn metrics_handler() -> Response {
    let mut headers = HeaderMap::new();
    headers = add_cors_headers(headers);

    match metrics::encode_metrics() {
        Ok(metrics) => {
            headers.insert(CONTENT_TYPE, "text/plain; charset=utf-8".parse().unwrap());
            (
                StatusCode::OK,
                headers,
                metrics,
            )
                .into_response()
        },
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            headers,
            format!("Failed to encode metrics: {}", e),
        )
            .into_response(),
    }
}
