use anyhow::{Result, Context};
use crossbeam::queue::ArrayQueue;
use std::sync::Arc;
use tracing::warn;

/// Shared memory ring buffer for ultra-low latency local IPC
/// Uses lock-free queue for high throughput
pub struct SharedMemoryRing {
    name: String,
    queue: Arc<ArrayQueue<Vec<u8>>>,
    max_message_size: usize,
}

impl SharedMemoryRing {
    pub fn new(name: &str, capacity: usize) -> Result<Self> {
        // For now, use in-memory queue
        // TODO: Replace with actual shared memory implementation using shared_memory crate
        let queue = Arc::new(ArrayQueue::new(1024)); // 1024 messages

        Ok(Self {
            name: name.to_string(),
            queue,
            max_message_size: 64 * 1024, // 64KB max message
        })
    }

    pub fn write(&self, data: &[u8]) -> Result<()> {
        if data.len() > self.max_message_size {
            return Err(anyhow::anyhow!("Message too large"));
        }

        match self.queue.push(data.to_vec()) {
            Ok(_) => Ok(()),
            Err(_) => {
                warn!("Shared memory ring buffer full, dropping message");
                Err(anyhow::anyhow!("Ring buffer full"))
            }
        }
    }

    pub fn read(&self) -> Option<Vec<u8>> {
        self.queue.pop()
    }

    pub fn len(&self) -> usize {
        self.queue.len()
    }

    pub fn is_empty(&self) -> bool {
        self.queue.is_empty()
    }
}

// TODO: Implement true shared memory using the shared_memory crate
// This would allow other processes on the same machine to read market data
// with <1μs latency instead of going through Redis
