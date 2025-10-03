use crate::state::AppState;
use crate::ui::theme::Theme;
use ratatui::{
    layout::{Alignment, Rect},
    style::{Color, Modifier, Style},
    text::{Line, Span},
    widgets::{Block, Borders, Clear, Paragraph},
    Frame,
};

pub fn render(f: &mut Frame, area: Rect, _state: &AppState) {
    let theme = Theme::default();

    // Create a centered popup
    let popup_area = centered_rect(60, 70, area);

    // Clear the background
    f.render_widget(Clear, popup_area);

    let help_text = vec![
        Line::from(vec![Span::styled(
            "B25 Terminal UI - Keyboard Shortcuts",
            Style::default()
                .fg(theme.highlight)
                .add_modifier(Modifier::BOLD),
        )]),
        Line::from(""),
        Line::from(vec![Span::styled(
            "Global Commands:",
            Style::default().add_modifier(Modifier::BOLD),
        )]),
        Line::from("  q, Ctrl+C    Quit application"),
        Line::from("  ?            Toggle this help screen"),
        Line::from("  r            Reload configuration"),
        Line::from("  Tab          Next panel"),
        Line::from("  Shift+Tab    Previous panel"),
        Line::from("  :            Enter command mode"),
        Line::from(""),
        Line::from(vec![Span::styled(
            "Navigation:",
            Style::default().add_modifier(Modifier::BOLD),
        )]),
        Line::from("  j / Down     Scroll down"),
        Line::from("  k / Up       Scroll up"),
        Line::from(""),
        Line::from(vec![Span::styled(
            "Panel-Specific (Orders):",
            Style::default().add_modifier(Modifier::BOLD),
        )]),
        Line::from("  c            Cancel selected order"),
        Line::from("  C            Cancel all orders"),
        Line::from(""),
        Line::from(vec![Span::styled(
            "Panel-Specific (Positions):",
            Style::default().add_modifier(Modifier::BOLD),
        )]),
        Line::from("  x            Close selected position"),
        Line::from("  X            Close all positions"),
        Line::from(""),
        Line::from(vec![Span::styled(
            "Command Mode:",
            Style::default().add_modifier(Modifier::BOLD),
        )]),
        Line::from("  :buy <symbol> <size> <price>    Place limit buy order"),
        Line::from("  :sell <symbol> <size> <price>   Place limit sell order"),
        Line::from("  :market <side> <symbol> <size>  Place market order"),
        Line::from("  :cancel <order_id>              Cancel specific order"),
        Line::from("  :close <symbol>                 Close position"),
        Line::from(""),
        Line::from(vec![Span::styled(
            "Press ? to close this help screen",
            Style::default().fg(theme.text_dim),
        )]),
    ];

    let paragraph = Paragraph::new(help_text)
        .block(
            Block::default()
                .title(" Help ")
                .borders(Borders::ALL)
                .border_style(Style::default().fg(theme.highlight)),
        )
        .style(Style::default().fg(theme.text).bg(Color::Black))
        .alignment(Alignment::Left);

    f.render_widget(paragraph, popup_area);
}

fn centered_rect(percent_x: u16, percent_y: u16, r: Rect) -> Rect {
    let popup_layout = ratatui::layout::Layout::default()
        .direction(ratatui::layout::Direction::Vertical)
        .constraints([
            ratatui::layout::Constraint::Percentage((100 - percent_y) / 2),
            ratatui::layout::Constraint::Percentage(percent_y),
            ratatui::layout::Constraint::Percentage((100 - percent_y) / 2),
        ])
        .split(r);

    ratatui::layout::Layout::default()
        .direction(ratatui::layout::Direction::Horizontal)
        .constraints([
            ratatui::layout::Constraint::Percentage((100 - percent_x) / 2),
            ratatui::layout::Constraint::Percentage(percent_x),
            ratatui::layout::Constraint::Percentage((100 - percent_x) / 2),
        ])
        .split(popup_layout[1])[1]
}
