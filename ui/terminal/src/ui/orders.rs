use crate::state::AppState;
use crate::types::{OrderStatus, Panel};
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

    let is_focused = state.focused_panel == Panel::Orders;
    let border_style = if is_focused {
        Style::default().fg(theme.border_focused)
    } else {
        Style::default().fg(theme.border)
    };

    let mut lines = vec![Line::from(vec![
        Span::styled(
            format!("{:<12}", "ID"),
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
            format!("{:<10}", "Status"),
            Style::default().add_modifier(Modifier::BOLD),
        ),
    ])];

    if state.orders.is_empty() {
        lines.push(Line::from(vec![Span::styled(
            "No active orders",
            Style::default().fg(theme.text_dim),
        )]));
    } else {
        for (idx, order) in state.orders.iter().enumerate() {
            let side_text = match order.side {
                crate::types::OrderSide::Buy => "BUY",
                crate::types::OrderSide::Sell => "SELL",
            };

            let status_text = match order.status {
                OrderStatus::New => "NEW",
                OrderStatus::PartiallyFilled => "PARTIAL",
                OrderStatus::Filled => "FILLED",
                OrderStatus::Canceled => "CANCELED",
                OrderStatus::Rejected => "REJECTED",
            };

            let side_style = theme.order_side_style(&order.side);
            let mut line_style = Style::default().fg(theme.text);

            if is_focused && idx == state.selected_index {
                line_style = line_style
                    .bg(theme.border_focused)
                    .add_modifier(Modifier::BOLD);
            }

            lines.push(Line::from(vec![
                Span::styled(format!("{:<12}", order.id), line_style),
                Span::raw(" "),
                Span::styled(format!("{:<10}", order.symbol), line_style),
                Span::raw(" "),
                Span::styled(format!("{:<4}", side_text), side_style),
                Span::raw(" "),
                Span::styled(format!("{:>10.2}", order.price), line_style),
                Span::raw(" "),
                Span::styled(format!("{:>8.4}", order.size), line_style),
                Span::raw(" "),
                Span::styled(format!("{:<10}", status_text), line_style),
            ]));
        }
    }

    let paragraph = Paragraph::new(lines).block(
        Block::default()
            .title(" ACTIVE ORDERS ")
            .borders(Borders::ALL)
            .border_type(BorderType::Rounded)
            .border_style(border_style),
    );

    f.render_widget(paragraph, area);
}
