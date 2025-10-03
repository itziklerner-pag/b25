use crate::state::AppState;
use crate::types::ConnectionStatus;
use crate::ui::theme::Theme;
use chrono::Utc;
use ratatui::{
    layout::Rect,
    style::{Color, Style},
    text::{Line, Span},
    widgets::Paragraph,
    Frame,
};

pub fn render(f: &mut Frame, area: Rect, state: &AppState) {
    let theme = Theme::default();

    let status_symbol = match state.connection_status {
        ConnectionStatus::Connected => "●",
        ConnectionStatus::Connecting => "◐",
        ConnectionStatus::Disconnected => "○",
        ConnectionStatus::Error => "✖",
    };

    let status_text = match state.connection_status {
        ConnectionStatus::Connected => "Connected",
        ConnectionStatus::Connecting => "Connecting",
        ConnectionStatus::Disconnected => "Disconnected",
        ConnectionStatus::Error => "Error",
    };

    let latency_text = if state.connection_status == ConnectionStatus::Connected {
        format!(" ({}ms)", state.connection_latency_ms)
    } else {
        String::new()
    };

    let stale_indicator = if state.is_stale() {
        Span::styled(" ⚠ STALE DATA", Style::default().fg(theme.warning))
    } else {
        Span::raw("")
    };

    let current_time = Utc::now().format("%H:%M:%S UTC").to_string();

    let line = Line::from(vec![
        Span::styled(
            " B25 Trading System ",
            Style::default().fg(theme.highlight),
        ),
        Span::raw("│ WS: "),
        Span::styled(
            status_symbol,
            theme.connection_status_style(&state.connection_status),
        ),
        Span::raw(" "),
        Span::styled(
            format!("{}{}", status_text, latency_text),
            theme.connection_status_style(&state.connection_status),
        ),
        stale_indicator,
        Span::raw(" │ "),
        Span::styled(current_time, Style::default().fg(theme.text_dim)),
    ]);

    let paragraph = Paragraph::new(line).style(Style::default().fg(theme.text).bg(Color::Black));

    f.render_widget(paragraph, area);
}
