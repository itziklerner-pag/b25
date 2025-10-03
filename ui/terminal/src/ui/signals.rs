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

    let is_focused = state.focused_panel == Panel::Signals;
    let border_style = if is_focused {
        Style::default().fg(theme.border_focused)
    } else {
        Style::default().fg(theme.border)
    };

    let mut lines = vec![Line::from(vec![
        Span::styled(
            format!("{:<8}", "Time"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
        Span::raw(" "),
        Span::styled(
            format!("{:<12}", "Strategy"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
        Span::raw(" "),
        Span::styled(
            format!("{:<10}", "Symbol"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
        Span::raw(" "),
        Span::styled(
            format!("{:<7}", "Signal"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
        Span::raw(" "),
        Span::styled(
            format!("{:<12}", "Strength"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
        Span::raw(" "),
        Span::styled(
            format!("{:>10}", "Target"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
    ])];

    if state.signals.is_empty() {
        lines.push(Line::from(vec![Span::styled(
            "No AI signals (placeholder)",
            Style::default().fg(theme.text_dim),
        )]));
    } else {
        for (idx, signal) in state.signals.iter().enumerate() {
            let signal_text = match signal.signal_type {
                crate::types::SignalType::Long => "LONG",
                crate::types::SignalType::Short => "SHORT",
                crate::types::SignalType::Neutral => "NEUTRAL",
            };

            let time_diff = Utc::now()
                .signed_duration_since(signal.timestamp)
                .num_seconds();
            let time_text = if time_diff < 60 {
                format!("{}s ago", time_diff)
            } else if time_diff < 3600 {
                format!("{}m ago", time_diff / 60)
            } else {
                signal.timestamp.format("%H:%M").to_string()
            };

            // Create strength bar (0.0 to 1.0)
            let bar_length = (signal.strength * 10.0) as usize;
            let strength_bar = "█".repeat(bar_length);

            let signal_style = theme.signal_type_style(&signal.signal_type);
            let mut line_style = Style::default().fg(theme.text);

            if is_focused && idx == state.selected_index {
                line_style = line_style
                    .bg(theme.border_focused)
                    .add_modifier(Modifier::BOLD);
            }

            lines.push(Line::from(vec![
                Span::styled(format!("{:<8}", time_text), line_style),
                Span::raw(" "),
                Span::styled(format!("{:<12}", signal.strategy), line_style),
                Span::raw(" "),
                Span::styled(format!("{:<10}", signal.symbol), line_style),
                Span::raw(" "),
                Span::styled(format!("{:<7}", signal_text), signal_style),
                Span::raw(" "),
                Span::styled(format!("{:<12}", strength_bar), signal_style),
                Span::raw(" "),
                Span::styled(
                    format!("{:>10.2}", signal.target_price),
                    line_style,
                ),
            ]));
        }
    }

    let paragraph = Paragraph::new(lines).block(
        Block::default()
            .title(" AI SIGNALS ")
            .borders(Borders::ALL)
            .border_type(BorderType::Rounded)
            .border_style(border_style),
    );

    f.render_widget(paragraph, area);
}
