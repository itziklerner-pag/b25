use anyhow::Result;
use clap::Parser;
use crossterm::{
    event::{DisableMouseCapture, EnableMouseCapture},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use ratatui::{backend::CrosstermBackend, Terminal};
use std::sync::Arc;
use tokio::sync::mpsc;
use parking_lot::RwLock;

mod config;
mod state;
mod ui;
mod websocket;
mod keyboard;
mod types;

use config::Config;
use state::{AppState, StateUpdate};
use websocket::WsClient;
use keyboard::{KeyboardHandler, Action};

#[derive(Parser, Debug)]
#[command(name = "b25-terminal-ui")]
#[command(about = "Real-time terminal UI for B25 HFT trading system", long_about = None)]
struct Args {
    /// Configuration file path
    #[arg(short, long, default_value = "config.yaml")]
    config: String,

    /// Dashboard WebSocket URL (overrides config)
    #[arg(short, long)]
    url: Option<String>,

    /// Log level (trace, debug, info, warn, error)
    #[arg(short, long)]
    log_level: Option<String>,
}

#[tokio::main]
async fn main() -> Result<()> {
    let args = Args::parse();

    // Load configuration
    let mut config = Config::load(&args.config).unwrap_or_else(|e| {
        eprintln!("Failed to load config from {}: {}", args.config, e);
        eprintln!("Using default configuration");
        Config::default()
    });

    // Override config with command-line arguments
    if let Some(url) = args.url {
        config.connection.dashboard_url = url;
    }

    if let Some(level) = args.log_level {
        config.logging.level = level;
    }

    // Initialize logging
    init_logging(&config)?;
    tracing::info!("Starting B25 Terminal UI");
    tracing::debug!("Configuration: {:?}", config);

    // Run the application
    if let Err(e) = run_app(config).await {
        tracing::error!("Application error: {}", e);
        eprintln!("Error: {}", e);
        return Err(e);
    }

    Ok(())
}

async fn run_app(config: Config) -> Result<()> {
    // Setup terminal
    enable_raw_mode()?;
    let mut stdout = std::io::stdout();
    execute!(stdout, EnterAlternateScreen, EnableMouseCapture)?;
    let backend = CrosstermBackend::new(stdout);
    let mut terminal = Terminal::new(backend)?;

    // Initialize shared state
    let state = Arc::new(RwLock::new(AppState::new(config.clone())));

    // Create channels
    let (state_tx, mut state_rx) = mpsc::channel::<StateUpdate>(1000);
    let (action_tx, mut action_rx) = mpsc::channel::<Action>(100);

    // Spawn WebSocket client
    let ws_client = WsClient::new(config.connection.clone(), state_tx);
    let ws_handle = tokio::spawn(async move {
        ws_client.connect_with_retry().await
    });

    // Spawn keyboard handler
    let keyboard_handler = KeyboardHandler::new(action_tx.clone());
    let keyboard_handle = tokio::spawn(async move {
        keyboard_handler.run().await
    });

    // Spawn state updater
    let state_clone = state.clone();
    tokio::spawn(async move {
        while let Some(update) = state_rx.recv().await {
            state_clone.write().apply_update(update);
        }
    });

    // Main render and action loop
    let mut interval = tokio::time::interval(
        std::time::Duration::from_millis(config.ui.refresh_rate_ms)
    );

    let result = loop {
        tokio::select! {
            _ = interval.tick() => {
                // Render UI
                if let Err(e) = terminal.draw(|f| {
                    ui::render(f, &state);
                }) {
                    tracing::error!("Render error: {}", e);
                    break Err(e.into());
                }

                // Clear dirty flags after render
                state.write().clear_dirty();
            }

            Some(action) = action_rx.recv() => {
                match action {
                    Action::Quit => {
                        tracing::info!("Quit action received");
                        break Ok(());
                    }
                    action => {
                        if let Err(e) = handle_action(action, &state).await {
                            tracing::error!("Action handler error: {}", e);
                        }
                    }
                }
            }
        }
    };

    // Cleanup
    tracing::info!("Shutting down...");
    ws_handle.abort();
    keyboard_handle.abort();

    disable_raw_mode()?;
    execute!(
        terminal.backend_mut(),
        LeaveAlternateScreen,
        DisableMouseCapture
    )?;
    terminal.show_cursor()?;

    result
}

async fn handle_action(action: Action, state: &Arc<RwLock<AppState>>) -> Result<()> {
    match action {
        Action::NextPanel => {
            state.write().next_panel();
        }
        Action::PrevPanel => {
            state.write().prev_panel();
        }
        Action::ShowHelp => {
            state.write().toggle_help();
        }
        Action::ReloadConfig => {
            tracing::info!("Config reload requested");
            // TODO: Implement config reload
        }
        Action::EnterCommandMode => {
            state.write().enter_command_mode();
        }
        Action::ExitCommandMode => {
            state.write().exit_command_mode();
        }
        Action::CommandInput(c) => {
            state.write().command_input(c);
        }
        Action::CommandBackspace => {
            state.write().command_backspace();
        }
        Action::ExecuteCommand(cmd) => {
            tracing::info!("Executing command: {}", cmd);
            execute_command(&cmd, state).await?;
        }
        Action::CancelSelectedOrder => {
            if let Some(order_id) = state.read().get_selected_order_id() {
                tracing::info!("Canceling selected order: {}", order_id);
                // TODO: Send cancel request to order execution service
            }
        }
        Action::CloseSelectedPosition => {
            if let Some(symbol) = state.read().get_selected_position_symbol() {
                tracing::info!("Closing selected position: {}", symbol);
                // TODO: Send close position request
            }
        }
        Action::ScrollUp => {
            state.write().scroll_up();
        }
        Action::ScrollDown => {
            state.write().scroll_down();
        }
        _ => {}
    }
    Ok(())
}

async fn execute_command(cmd: &str, state: &Arc<RwLock<AppState>>) -> Result<()> {
    let parts: Vec<&str> = cmd.split_whitespace().collect();
    if parts.is_empty() {
        return Ok(());
    }

    match parts[0] {
        "buy" | "sell" => {
            if parts.len() < 4 {
                tracing::warn!("Invalid order command format. Usage: buy/sell <symbol> <size> <price>");
                return Ok(());
            }
            let side = parts[0];
            let symbol = parts[1];
            let size = parts[2].parse::<f64>().ok();
            let price = parts[3].parse::<f64>().ok();

            if let (Some(size), Some(price)) = (size, price) {
                tracing::info!("Placing {} order: {} {} @ {}", side, symbol, size, price);
                // TODO: Send order request to order execution service
            }
        }
        "market" => {
            if parts.len() < 3 {
                tracing::warn!("Invalid market order format. Usage: market <buy/sell> <symbol> <size>");
                return Ok(());
            }
            let side = parts[1];
            let symbol = parts[2];
            let size = parts[3].parse::<f64>().ok();

            if let Some(size) = size {
                tracing::info!("Placing market {} order: {} {}", side, symbol, size);
                // TODO: Send market order request
            }
        }
        "cancel" => {
            if parts.len() < 2 {
                tracing::warn!("Invalid cancel command. Usage: cancel <order_id>");
                return Ok(());
            }
            let order_id = parts[1];
            tracing::info!("Canceling order: {}", order_id);
            // TODO: Send cancel request
        }
        "close" => {
            if parts.len() < 2 {
                tracing::warn!("Invalid close command. Usage: close <symbol>");
                return Ok(());
            }
            let symbol = parts[1];
            tracing::info!("Closing position: {}", symbol);
            // TODO: Send close position request
        }
        _ => {
            tracing::warn!("Unknown command: {}", parts[0]);
        }
    }

    Ok(())
}

fn init_logging(config: &Config) -> Result<()> {
    use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

    let env_filter = EnvFilter::try_from_default_env()
        .unwrap_or_else(|_| EnvFilter::new(&config.logging.level));

    let registry = tracing_subscriber::registry().with(env_filter);

    if config.logging.json {
        registry
            .with(tracing_subscriber::fmt::layer().json())
            .init();
    } else {
        registry
            .with(tracing_subscriber::fmt::layer())
            .init();
    }

    Ok(())
}
