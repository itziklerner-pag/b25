use crate::state::AppState;
use parking_lot::RwLock;
use ratatui::{
    layout::{Constraint, Direction, Layout, Rect},
    Frame,
};
use std::sync::Arc;

mod theme;
mod status;
mod help;
mod positions;
mod orders;
mod fills;
mod orderbook;
mod signals;
mod alerts;

pub use theme::Theme;

pub fn render(f: &mut Frame, state: &Arc<RwLock<AppState>>) {
    let state = state.read();

    // Create main layout
    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Length(1), // Status bar
            Constraint::Min(0),    // Main content
            Constraint::Length(3), // Alerts
            Constraint::Length(1), // Help bar
        ])
        .split(f.size());

    // Render status bar
    status::render(f, chunks[0], &state);

    // Render main content area
    render_main_content(f, chunks[1], &state);

    // Render alerts panel
    alerts::render(f, chunks[2], &state);

    // Render help bar
    render_help_bar(f, chunks[3], &state);

    // Render help overlay if shown
    if state.show_help {
        help::render(f, f.size(), &state);
    }
}

fn render_main_content(f: &mut Frame, area: Rect, state: &AppState) {
    // Split into left and right columns
    let columns = Layout::default()
        .direction(Direction::Horizontal)
        .constraints([Constraint::Percentage(40), Constraint::Percentage(60)])
        .split(area);

    // Left column: Positions, Orders, Fills
    let left_panels = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Percentage(30),
            Constraint::Percentage(35),
            Constraint::Percentage(35),
        ])
        .split(columns[0]);

    positions::render(f, left_panels[0], state);
    orders::render(f, left_panels[1], state);
    fills::render(f, left_panels[2], state);

    // Right column: OrderBook, Signals
    let right_panels = Layout::default()
        .direction(Direction::Vertical)
        .constraints([Constraint::Percentage(60), Constraint::Percentage(40)])
        .split(columns[1]);

    orderbook::render(f, right_panels[0], state);
    signals::render(f, right_panels[1], state);
}

fn render_help_bar(f: &mut Frame, area: Rect, state: &AppState) {
    use ratatui::{
        style::{Color, Style},
        text::{Line, Span},
        widgets::Paragraph,
    };

    let theme = Theme::default();

    let help_text = if state.input_mode == crate::types::InputMode::Command {
        vec![
            Span::styled("Enter", Style::default().fg(theme.highlight)),
            Span::raw(" Execute | "),
            Span::styled("Esc", Style::default().fg(theme.highlight)),
            Span::raw(" Cancel"),
        ]
    } else {
        vec![
            Span::styled("?", Style::default().fg(theme.highlight)),
            Span::raw(" Help | "),
            Span::styled(":", Style::default().fg(theme.highlight)),
            Span::raw(" Command | "),
            Span::styled("q", Style::default().fg(theme.highlight)),
            Span::raw(" Quit | "),
            Span::styled("Tab", Style::default().fg(theme.highlight)),
            Span::raw(" Switch Panel | "),
            Span::styled("c", Style::default().fg(theme.highlight)),
            Span::raw(" Cancel Order | "),
            Span::styled("x", Style::default().fg(theme.highlight)),
            Span::raw(" Close Position"),
        ]
    };

    let paragraph = Paragraph::new(Line::from(help_text))
        .style(Style::default().fg(theme.text).bg(Color::Black));

    f.render_widget(paragraph, area);
}
