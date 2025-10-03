use crate::state::AppState;
use crate::types::Panel;
use crate::ui::theme::Theme;
use ratatui::{
    layout::{Constraint, Direction, Layout, Rect},
    style::{Modifier, Style},
    text::{Line, Span},
    widgets::{Block, Borders, BorderType, Paragraph},
    Frame,
};

pub fn render(f: &mut Frame, area: Rect, state: &AppState) {
    let theme = Theme::default();

    let is_focused = state.focused_panel == Panel::OrderBook;
    let border_style = if is_focused {
        Style::default().fg(theme.border_focused)
    } else {
        Style::default().fg(theme.border)
    };

    if let Some(orderbook) = &state.orderbook {
        let title = format!(" ORDER BOOK ({}) ", orderbook.symbol);

        let block = Block::default()
            .title(title)
            .borders(Borders::ALL)
            .border_type(BorderType::Rounded)
            .border_style(border_style);

        let inner = block.inner(area);
        f.render_widget(block, area);

        // Split into three columns: Bids | Price | Asks
        let columns = Layout::default()
            .direction(Direction::Horizontal)
            .constraints([
                Constraint::Percentage(40),
                Constraint::Percentage(20),
                Constraint::Percentage(40),
            ])
            .split(inner);

        render_bids(f, columns[0], &orderbook.bids, &theme);
        render_prices(f, columns[1], orderbook, &theme);
        render_asks(f, columns[2], &orderbook.asks, &theme);
    } else {
        let paragraph = Paragraph::new(vec![Line::from(vec![Span::styled(
            "No orderbook data",
            Style::default().fg(theme.text_dim),
        )])])
        .block(
            Block::default()
                .title(" ORDER BOOK ")
                .borders(Borders::ALL)
                .border_type(BorderType::Rounded)
                .border_style(border_style),
        );

        f.render_widget(paragraph, area);
    }
}

fn render_bids(f: &mut Frame, area: Rect, bids: &[(f64, f64)], theme: &Theme) {
    let max_size = bids.iter().map(|(_, s)| *s).fold(0.0f64, f64::max);

    let mut lines = vec![Line::from(vec![Span::styled(
        format!("{:>8} {:>10}", "Size", "Price"),
        Style::default().add_modifier(Modifier::BOLD),
    )])];

    for (price, size) in bids.iter().take(10) {
        let bar_length = if max_size > 0.0 {
            ((*size / max_size) * 15.0) as usize
        } else {
            0
        };
        let bar = "█".repeat(bar_length);

        lines.push(Line::from(vec![
            Span::styled(
                format!("{:>8.4}", size),
                Style::default().fg(theme.buy),
            ),
            Span::raw(" "),
            Span::styled(bar, Style::default().fg(theme.buy)),
            Span::raw(" "),
            Span::styled(
                format!("{:>10.2}", price),
                Style::default().fg(theme.text),
            ),
        ]));
    }

    let paragraph = Paragraph::new(lines);
    f.render_widget(paragraph, area);
}

fn render_prices(
    f: &mut Frame,
    area: Rect,
    orderbook: &crate::types::OrderBook,
    theme: &Theme,
) {
    let spread = orderbook.spread();
    let spread_pct = orderbook.spread_percent();
    let mid = orderbook.mid_price();

    let lines = vec![
        Line::from(""),
        Line::from(""),
        Line::from(""),
        Line::from(""),
        Line::from(vec![Span::styled(
            "─────────",
            Style::default().fg(theme.border),
        )]),
        Line::from(""),
        Line::from(vec![
            Span::styled("Spread: ", Style::default().fg(theme.text_dim)),
            Span::styled(
                format!("{:.2}", spread),
                Style::default().fg(theme.warning),
            ),
        ]),
        Line::from(vec![
            Span::styled("      (", Style::default().fg(theme.text_dim)),
            Span::styled(
                format!("{:.4}%", spread_pct),
                Style::default().fg(theme.warning),
            ),
            Span::styled(")", Style::default().fg(theme.text_dim)),
        ]),
        Line::from(vec![
            Span::styled("Mid: ", Style::default().fg(theme.text_dim)),
            Span::styled(format!("{:.2}", mid), Style::default().fg(theme.text)),
        ]),
    ];

    let paragraph = Paragraph::new(lines);
    f.render_widget(paragraph, area);
}

fn render_asks(f: &mut Frame, area: Rect, asks: &[(f64, f64)], theme: &Theme) {
    let max_size = asks.iter().map(|(_, s)| *s).fold(0.0f64, f64::max);

    let mut lines = vec![Line::from(vec![Span::styled(
        format!("{:<10} {:<8}", "Price", "Size"),
        Style::default().add_modifier(Modifier::BOLD),
    )])];

    for (price, size) in asks.iter().take(10) {
        let bar_length = if max_size > 0.0 {
            ((*size / max_size) * 15.0) as usize
        } else {
            0
        };
        let bar = "█".repeat(bar_length);

        lines.push(Line::from(vec![
            Span::styled(
                format!("{:<10.2}", price),
                Style::default().fg(theme.text),
            ),
            Span::raw(" "),
            Span::styled(bar, Style::default().fg(theme.sell)),
            Span::raw(" "),
            Span::styled(
                format!("{:<8.4}", size),
                Style::default().fg(theme.sell),
            ),
        ]));
    }

    let paragraph = Paragraph::new(lines);
    f.render_widget(paragraph, area);
}
