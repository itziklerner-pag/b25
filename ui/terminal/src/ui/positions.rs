use crate::state::AppState;
use crate::types::{Panel, PositionSide};
use crate::ui::theme::Theme;
use ratatui::{
    layout::Rect,
    style::{Modifier, Style},
    text::{Line, Span},
    widgets::{Block, Borders, BorderType, Paragraph},
    Frame,
};

pub fn render(f: &mut Frame, area: Rect, state: &AppState) {
    let theme = Theme::default();

    let is_focused = state.focused_panel == Panel::Positions;
    let border_style = if is_focused {
        Style::default().fg(theme.border_focused)
    } else {
        Style::default().fg(theme.border)
    };

    let mut lines = vec![Line::from(vec![
        Span::styled(
            format!("{:<10}", "Symbol"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
        Span::raw(" "),
        Span::styled(
            format!("{:>8}", "Size"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
        Span::raw(" "),
        Span::styled(
            format!("{:>10}", "Entry"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
        Span::raw(" "),
        Span::styled(
            format!("{:>10}", "Current"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
        Span::raw(" "),
        Span::styled(
            format!("{:>12}", "P&L"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
        Span::raw(" "),
        Span::styled(
            format!("{:>8}", "P&L %"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
    ])];

    if state.positions.is_empty() {
        lines.push(Line::from(vec![Span::styled(
            "No positions",
            Style::default().fg(theme.text_dim),
        )]));
    } else {
        for (idx, position) in state.positions.iter().enumerate() {
            let side_symbol = match position.side {
                PositionSide::Long => "+",
                PositionSide::Short => "-",
            };

            let pnl_style = theme.profit_style(position.unrealized_pnl);
            let mut line_style = Style::default().fg(theme.text);

            if is_focused && idx == state.selected_index {
                line_style = line_style
                    .bg(theme.border_focused)
                    .add_modifier(Modifier::BOLD);
            }

            lines.push(Line::from(vec![
                Span::styled(format!("{:<10}", position.symbol), line_style),
                Span::raw(" "),
                Span::styled(
                    format!("{}{:>7.4}", side_symbol, position.size.abs()),
                    line_style,
                ),
                Span::raw(" "),
                Span::styled(format!("{:>10.2}", position.entry_price), line_style),
                Span::raw(" "),
                Span::styled(format!("{:>10.2}", position.current_price), line_style),
                Span::raw(" "),
                Span::styled(
                    format!("{:+>12.2}", position.unrealized_pnl),
                    pnl_style,
                ),
                Span::raw(" "),
                Span::styled(
                    format!("{:+>7.2}%", position.pnl_percent),
                    pnl_style,
                ),
            ]));
        }
    }

    let paragraph = Paragraph::new(lines).block(
        Block::default()
            .title(" POSITIONS ")
            .borders(Borders::ALL)
            .border_type(BorderType::Rounded)
            .border_style(border_style),
    );

    f.render_widget(paragraph, area);
}
