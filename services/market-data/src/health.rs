use axum::{
    extract::State,
    http::StatusCode,
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

async fn health_handler() -> impl IntoResponse {
    Json(json!({
        "status": "healthy",
        "service": "market-data",
        "version": env!("CARGO_PKG_VERSION"),
    }))
}

async fn readiness_handler() -> impl IntoResponse {
    // TODO: Check Redis connection, WebSocket status, etc.
    Json(json!({
        "status": "ready",
    }))
}

async fn metrics_handler() -> Response {
    match metrics::encode_metrics() {
        Ok(metrics) => (
            StatusCode::OK,
            [("Content-Type", "text/plain; charset=utf-8")],
            metrics,
        )
            .into_response(),
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            format!("Failed to encode metrics: {}", e),
        )
            .into_response(),
    }
}
