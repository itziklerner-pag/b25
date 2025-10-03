use anyhow::Result;
use crossterm::event::{Event, EventStream, KeyCode, KeyModifiers};
use futures::StreamExt;
use tokio::sync::mpsc;

#[derive(Debug, Clone)]
pub enum Action {
    Quit,
    NextPanel,
    PrevPanel,
    ShowHelp,
    ReloadConfig,
    EnterCommandMode,
    ExitCommandMode,
    CommandInput(char),
    CommandBackspace,
    ExecuteCommand(String),
    CancelSelectedOrder,
    CancelAllOrders,
    CloseSelectedPosition,
    CloseAllPositions,
    ScrollUp,
    ScrollDown,
}

pub struct KeyboardHandler {
    action_tx: mpsc::Sender<Action>,
}

impl KeyboardHandler {
    pub fn new(action_tx: mpsc::Sender<Action>) -> Self {
        Self { action_tx }
    }

    pub async fn run(&self) -> Result<()> {
        let mut reader = EventStream::new();

        while let Some(event) = reader.next().await {
            match event {
                Ok(Event::Key(key)) => {
                    if let Some(action) = self.handle_key(key) {
                        self.action_tx.send(action).await?;
                    }
                }
                Ok(_) => {}
                Err(e) => {
                    tracing::error!("Event stream error: {}", e);
                }
            }
        }

        Ok(())
    }

    fn handle_key(&self, key: crossterm::event::KeyEvent) -> Option<Action> {
        // Handle Ctrl+C globally
        if key.code == KeyCode::Char('c') && key.modifiers.contains(KeyModifiers::CONTROL) {
            return Some(Action::Quit);
        }

        match key.code {
            KeyCode::Char('q') => Some(Action::Quit),
            KeyCode::Char('?') => Some(Action::ShowHelp),
            KeyCode::Char('r') => Some(Action::ReloadConfig),
            KeyCode::Tab => Some(Action::NextPanel),
            KeyCode::BackTab => Some(Action::PrevPanel),
            KeyCode::Char(':') => Some(Action::EnterCommandMode),
            KeyCode::Char('c') => Some(Action::CancelSelectedOrder),
            KeyCode::Char('C') => Some(Action::CancelAllOrders),
            KeyCode::Char('x') => Some(Action::CloseSelectedPosition),
            KeyCode::Char('X') => Some(Action::CloseAllPositions),
            KeyCode::Up | KeyCode::Char('k') => Some(Action::ScrollUp),
            KeyCode::Down | KeyCode::Char('j') => Some(Action::ScrollDown),
            _ => None,
        }
    }
}

pub struct CommandModeHandler {
    action_tx: mpsc::Sender<Action>,
}

impl CommandModeHandler {
    pub fn new(action_tx: mpsc::Sender<Action>) -> Self {
        Self { action_tx }
    }

    pub async fn run(&self) -> Result<()> {
        let mut reader = EventStream::new();

        while let Some(event) = reader.next().await {
            match event {
                Ok(Event::Key(key)) => {
                    if let Some(action) = self.handle_key(key) {
                        self.action_tx.send(action).await?;
                    }
                }
                Ok(_) => {}
                Err(e) => {
                    tracing::error!("Event stream error: {}", e);
                }
            }
        }

        Ok(())
    }

    fn handle_key(&self, key: crossterm::event::KeyEvent) -> Option<Action> {
        match key.code {
            KeyCode::Enter => Some(Action::ExitCommandMode),
            KeyCode::Esc => Some(Action::ExitCommandMode),
            KeyCode::Char(c) => Some(Action::CommandInput(c)),
            KeyCode::Backspace => Some(Action::CommandBackspace),
            _ => None,
        }
    }
}
