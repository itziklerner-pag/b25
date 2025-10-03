use ratatui::style::{Color, Modifier, Style};

#[derive(Debug, Clone)]
pub struct Theme {
    // Status colors
    pub connected: Color,
    pub disconnected: Color,
    pub warning: Color,
    pub error: Color,

    // P&L colors
    pub profit: Color,
    pub loss: Color,
    pub neutral: Color,

    // Order colors
    pub buy: Color,
    pub sell: Color,

    // Signal colors
    pub long: Color,
    pub short: Color,

    // UI elements
    pub border: Color,
    pub border_focused: Color,
    pub text: Color,
    pub text_dim: Color,
    pub highlight: Color,
    pub background: Color,
}

impl Default for Theme {
    fn default() -> Self {
        Self {
            connected: Color::Green,
            disconnected: Color::Red,
            warning: Color::Yellow,
            error: Color::Red,

            profit: Color::Green,
            loss: Color::Red,
            neutral: Color::Gray,

            buy: Color::Cyan,
            sell: Color::Magenta,

            long: Color::Green,
            short: Color::Red,

            border: Color::DarkGray,
            border_focused: Color::Cyan,
            text: Color::White,
            text_dim: Color::DarkGray,
            highlight: Color::Yellow,
            background: Color::Black,
        }
    }
}

impl Theme {
    pub fn profit_style(&self, value: f64) -> Style {
        if value > 0.0 {
            Style::default().fg(self.profit)
        } else if value < 0.0 {
            Style::default().fg(self.loss)
        } else {
            Style::default().fg(self.neutral)
        }
    }

    pub fn order_side_style(&self, side: &crate::types::OrderSide) -> Style {
        match side {
            crate::types::OrderSide::Buy => {
                Style::default().fg(self.buy).add_modifier(Modifier::BOLD)
            }
            crate::types::OrderSide::Sell => {
                Style::default().fg(self.sell).add_modifier(Modifier::BOLD)
            }
        }
    }

    pub fn signal_type_style(&self, signal_type: &crate::types::SignalType) -> Style {
        match signal_type {
            crate::types::SignalType::Long => {
                Style::default().fg(self.long).add_modifier(Modifier::BOLD)
            }
            crate::types::SignalType::Short => {
                Style::default().fg(self.short).add_modifier(Modifier::BOLD)
            }
            crate::types::SignalType::Neutral => Style::default().fg(self.neutral),
        }
    }

    pub fn alert_level_style(&self, level: &crate::types::AlertLevel) -> Style {
        match level {
            crate::types::AlertLevel::Error => Style::default().fg(self.error),
            crate::types::AlertLevel::Warning => Style::default().fg(self.warning),
            crate::types::AlertLevel::Info => Style::default().fg(self.text),
        }
    }

    pub fn connection_status_style(&self, status: &crate::types::ConnectionStatus) -> Style {
        match status {
            crate::types::ConnectionStatus::Connected => Style::default().fg(self.connected),
            crate::types::ConnectionStatus::Connecting => Style::default().fg(self.warning),
            crate::types::ConnectionStatus::Disconnected => Style::default().fg(self.disconnected),
            crate::types::ConnectionStatus::Error => Style::default().fg(self.error),
        }
    }
}
