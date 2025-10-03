use crate::config::Config;
use crate::types::*;
use chrono::Utc;
use std::collections::VecDeque;

#[derive(Debug, Clone)]
pub struct AppState {
    pub config: Config,
    pub positions: Vec<Position>,
    pub orders: Vec<Order>,
    pub orderbook: Option<OrderBook>,
    pub fills: VecDeque<Fill>,
    pub signals: VecDeque<Signal>,
    pub alerts: VecDeque<Alert>,
    pub connection_status: ConnectionStatus,
    pub connection_latency_ms: u64,
    pub last_update: chrono::DateTime<Utc>,
    pub focused_panel: Panel,
    pub input_mode: InputMode,
    pub command_buffer: String,
    pub show_help: bool,
    pub selected_index: usize,
    pub scroll_offset: usize,
    pub dirty: DirtyFlags,
}

#[derive(Debug, Clone, Default)]
pub struct DirtyFlags {
    pub positions: bool,
    pub orders: bool,
    pub orderbook: bool,
    pub fills: bool,
    pub signals: bool,
    pub alerts: bool,
    pub status: bool,
    pub all: bool,
}

#[derive(Debug, Clone)]
pub enum StateUpdate {
    Positions(Vec<Position>),
    Orders(Vec<Order>),
    OrderBook(OrderBook),
    Fills(Vec<Fill>),
    Signals(Vec<Signal>),
    Alerts(Vec<Alert>),
    ConnectionStatus(ConnectionStatus, u64),
}

impl AppState {
    pub fn new(config: Config) -> Self {
        Self {
            config,
            positions: Vec::new(),
            orders: Vec::new(),
            orderbook: None,
            fills: VecDeque::new(),
            signals: VecDeque::new(),
            alerts: VecDeque::new(),
            connection_status: ConnectionStatus::Disconnected,
            connection_latency_ms: 0,
            last_update: Utc::now(),
            focused_panel: Panel::Positions,
            input_mode: InputMode::Normal,
            command_buffer: String::new(),
            show_help: false,
            selected_index: 0,
            scroll_offset: 0,
            dirty: DirtyFlags::default(),
        }
    }

    pub fn apply_update(&mut self, update: StateUpdate) {
        self.last_update = Utc::now();

        match update {
            StateUpdate::Positions(positions) => {
                if self.positions != positions {
                    self.positions = positions;
                    self.dirty.positions = true;
                    self.dirty.all = true;
                }
            }
            StateUpdate::Orders(orders) => {
                if self.orders != orders {
                    self.orders = orders;
                    self.dirty.orders = true;
                    self.dirty.all = true;
                }
            }
            StateUpdate::OrderBook(orderbook) => {
                self.orderbook = Some(orderbook);
                self.dirty.orderbook = true;
                self.dirty.all = true;
            }
            StateUpdate::Fills(fills) => {
                for fill in fills {
                    self.fills.push_front(fill);
                }
                while self.fills.len() > self.config.panels.max_fills_display {
                    self.fills.pop_back();
                }
                self.dirty.fills = true;
                self.dirty.all = true;
            }
            StateUpdate::Signals(signals) => {
                for signal in signals {
                    self.signals.push_front(signal);
                }
                while self.signals.len() > self.config.panels.max_signals_display {
                    self.signals.pop_back();
                }
                self.dirty.signals = true;
                self.dirty.all = true;
            }
            StateUpdate::Alerts(alerts) => {
                for alert in alerts {
                    self.alerts.push_front(alert);
                }
                while self.alerts.len() > self.config.panels.max_alerts_display {
                    self.alerts.pop_back();
                }
                self.dirty.alerts = true;
                self.dirty.all = true;
            }
            StateUpdate::ConnectionStatus(status, latency) => {
                self.connection_status = status;
                self.connection_latency_ms = latency;
                self.dirty.status = true;
                self.dirty.all = true;
            }
        }
    }

    pub fn clear_dirty(&mut self) {
        self.dirty = DirtyFlags::default();
    }

    pub fn next_panel(&mut self) {
        self.focused_panel = self.focused_panel.next();
        self.selected_index = 0;
        self.scroll_offset = 0;
        self.dirty.all = true;
    }

    pub fn prev_panel(&mut self) {
        self.focused_panel = self.focused_panel.prev();
        self.selected_index = 0;
        self.scroll_offset = 0;
        self.dirty.all = true;
    }

    pub fn toggle_help(&mut self) {
        self.show_help = !self.show_help;
        self.dirty.all = true;
    }

    pub fn enter_command_mode(&mut self) {
        self.input_mode = InputMode::Command;
        self.command_buffer.clear();
        self.dirty.all = true;
    }

    pub fn exit_command_mode(&mut self) {
        self.input_mode = InputMode::Normal;
        self.command_buffer.clear();
        self.dirty.all = true;
    }

    pub fn command_input(&mut self, c: char) {
        self.command_buffer.push(c);
        self.dirty.all = true;
    }

    pub fn command_backspace(&mut self) {
        self.command_buffer.pop();
        self.dirty.all = true;
    }

    pub fn get_selected_order_id(&self) -> Option<String> {
        if self.focused_panel == Panel::Orders && self.selected_index < self.orders.len() {
            Some(self.orders[self.selected_index].id.clone())
        } else {
            None
        }
    }

    pub fn get_selected_position_symbol(&self) -> Option<String> {
        if self.focused_panel == Panel::Positions && self.selected_index < self.positions.len() {
            Some(self.positions[self.selected_index].symbol.clone())
        } else {
            None
        }
    }

    pub fn scroll_up(&mut self) {
        if self.selected_index > 0 {
            self.selected_index -= 1;
            self.dirty.all = true;
        }
    }

    pub fn scroll_down(&mut self) {
        let max_index = match self.focused_panel {
            Panel::Positions => self.positions.len().saturating_sub(1),
            Panel::Orders => self.orders.len().saturating_sub(1),
            Panel::Fills => self.fills.len().saturating_sub(1),
            Panel::Signals => self.signals.len().saturating_sub(1),
            Panel::Alerts => self.alerts.len().saturating_sub(1),
            _ => 0,
        };

        if self.selected_index < max_index {
            self.selected_index += 1;
            self.dirty.all = true;
        }
    }

    pub fn is_stale(&self) -> bool {
        let elapsed = Utc::now()
            .signed_duration_since(self.last_update)
            .num_seconds();
        elapsed > self.config.ui.stale_data_threshold_s as i64
    }
}
