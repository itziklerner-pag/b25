use crate::state::AppState;
use crate::types::Panel;
use crate::ui::theme::Theme;
use chrono::Utc;
use ratatui::{
    layout::Rect,
    style::{Modifier, Style},
    text::{Line, Span},
    widgets::{Block, Borders, BorderType, Paragraph},
    Frame,
};

pub fn render(f: &mut Frame, area: Rect, state: &AppState) {
    let theme = Theme::default();

    let is_focused = state.focused_panel == Panel::Alerts;
    let border_style = if is_focused {
        Style::default().fg(theme.border_focused)
    } else {
        Style::default().fg(theme.border)
    };

    let mut lines = Vec::new();

    if state.alerts.is_empty() {
        lines.push(Line::from(vec![Span::styled(
            "No alerts",
            Style::default().fg(theme.text_dim),
        )]));
    } else {
        for (idx, alert) in state.alerts.iter().take(3).enumerate() {
            let level_text = match alert.level {
                crate::types::AlertLevel::Error => "[ERROR]",
                crate::types::AlertLevel::Warning => "[WARN] ",
                crate::types::AlertLevel::Info => "[INFO] ",
            };

            let time_diff = Utc::now()
                .signed_duration_since(alert.timestamp)
                .num_seconds();
            let time_text = if time_diff < 60 {
                format!("{:>2}s", time_diff)
            } else if time_diff < 3600 {
                format!("{:>2}m", time_diff / 60)
            } else {
                alert.timestamp.format("%H:%M").to_string()
            };

            let level_style = theme.alert_level_style(&alert.level);
            let mut line_style = Style::default().fg(theme.text);

            if is_focused && idx == state.selected_index {
                line_style = line_style
                    .bg(theme.border_focused)
                    .add_modifier(Modifier::BOLD);
            }

            lines.push(Line::from(vec![
                Span::styled(level_text, level_style),
                Span::raw("  "),
                Span::styled(time_text, Style::default().fg(theme.text_dim)),
                Span::raw("  "),
                Span::styled(&alert.message, line_style),
            ]));
        }
    }

    let paragraph = Paragraph::new(lines).block(
        Block::default()
            .title(" ALERTS ")
            .borders(Borders::ALL)
            .border_type(BorderType::Rounded)
            .border_style(border_style),
    );

    f.render_widget(paragraph, area);
}
