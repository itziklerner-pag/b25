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

    let is_focused = state.focused_panel == Panel::Fills;
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
            format!("{:<10}", "Symbol"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
        Span::raw(" "),
        Span::styled(
            format!("{:<4}", "Side"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
        Span::raw(" "),
        Span::styled(
            format!("{:>10}", "Price"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
        Span::raw(" "),
        Span::styled(
            format!("{:>8}", "Size"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
        Span::raw(" "),
        Span::styled(
            format!("{:>8}", "Fee"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
        Span::raw(" "),
        Span::styled(
            format!("{:>10}", "P&L"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
    ])];

    if state.fills.is_empty() {
        lines.push(Line::from(vec![Span::styled(
            "No recent fills",
            Style::default().fg(theme.text_dim),
        )]));
    } else {
        for (idx, fill) in state.fills.iter().enumerate() {
            let side_text = match fill.side {
                crate::types::OrderSide::Buy => "BUY",
                crate::types::OrderSide::Sell => "SELL",
            };

            let time_diff = Utc::now()
                .signed_duration_since(fill.timestamp)
                .num_seconds();
            let time_text = if time_diff < 60 {
                format!("{}s ago", time_diff)
            } else if time_diff < 3600 {
                format!("{}m ago", time_diff / 60)
            } else {
                fill.timestamp.format("%H:%M").to_string()
            };

            let side_style = theme.order_side_style(&fill.side);
            let pnl_style = theme.profit_style(fill.pnl);
            let mut line_style = Style::default().fg(theme.text);

            if is_focused && idx == state.selected_index {
                line_style = line_style
                    .bg(theme.border_focused)
                    .add_modifier(Modifier::BOLD);
            }

            lines.push(Line::from(vec![
                Span::styled(format!("{:<8}", time_text), line_style),
                Span::raw(" "),
                Span::styled(format!("{:<10}", fill.symbol), line_style),
                Span::raw(" "),
                Span::styled(format!("{:<4}", side_text), side_style),
                Span::raw(" "),
                Span::styled(format!("{:>10.2}", fill.price), line_style),
                Span::raw(" "),
                Span::styled(format!("{:>8.4}", fill.size), line_style),
                Span::raw(" "),
                Span::styled(format!("{:>8.4}", fill.fee), line_style),
                Span::raw(" "),
                Span::styled(format!("{:+>10.2}", fill.pnl), pnl_style),
            ]));
        }
    }

    let paragraph = Paragraph::new(lines).block(
        Block::default()
            .title(" RECENT FILLS ")
            .borders(Borders::ALL)
            .border_type(BorderType::Rounded)
            .border_style(border_style),
    );

    f.render_widget(paragraph, area);
}
