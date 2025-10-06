/**
 * Professional Logging System for B25 Web Dashboard
 *
 * Features:
 * - Configurable log levels (ERROR, WARN, INFO, DEBUG, TRACE)
 * - Color-coded console output
 * - Timestamps on all logs
 * - Context/module support
 * - Environment-based defaults
 * - Runtime configuration via localStorage
 * - Performance-optimized (no overhead when disabled)
 */

export enum LogLevel {
  ERROR = 0,
  WARN = 1,
  INFO = 2,
  DEBUG = 3,
  TRACE = 4,
}

export type LogLevelString = 'ERROR' | 'WARN' | 'INFO' | 'DEBUG' | 'TRACE';

interface LogConfig {
  level: LogLevel;
  enableTimestamps: boolean;
  enableColors: boolean;
}

class Logger {
  private config: LogConfig;
  private readonly STORAGE_KEY = 'b25_log_level';

  // Color codes for different log levels
  private readonly colors = {
    ERROR: '#ef4444', // red-500
    WARN: '#f59e0b',  // amber-500
    INFO: '#3b82f6',  // blue-500
    DEBUG: '#8b5cf6', // violet-500
    TRACE: '#6b7280', // gray-500
  };

  private readonly bgColors = {
    ERROR: '#fee2e2', // red-100
    WARN: '#fef3c7',  // amber-100
    INFO: '#dbeafe',  // blue-100
    DEBUG: '#ede9fe', // violet-100
    TRACE: '#f3f4f6', // gray-100
  };

  constructor() {
    this.config = {
      level: this.getInitialLogLevel(),
      enableTimestamps: true,
      enableColors: true,
    };

    // Listen for log level changes
    window.addEventListener('storage', this.handleStorageChange);
  }

  /**
   * Get initial log level from environment or localStorage
   */
  private getInitialLogLevel(): LogLevel {
    // First, check localStorage (user override)
    const storedLevel = localStorage.getItem(this.STORAGE_KEY);
    if (storedLevel) {
      return this.parseLogLevel(storedLevel);
    }

    // Then check environment variable
    const envLevel = import.meta.env.VITE_LOG_LEVEL;
    if (envLevel) {
      return this.parseLogLevel(envLevel);
    }

    // Default: WARN for production, DEBUG for development
    return import.meta.env.PROD ? LogLevel.WARN : LogLevel.DEBUG;
  }

  /**
   * Parse log level string to enum
   */
  private parseLogLevel(level: string): LogLevel {
    const upperLevel = level.toUpperCase();
    switch (upperLevel) {
      case 'ERROR':
        return LogLevel.ERROR;
      case 'WARN':
        return LogLevel.WARN;
      case 'INFO':
        return LogLevel.INFO;
      case 'DEBUG':
        return LogLevel.DEBUG;
      case 'TRACE':
        return LogLevel.TRACE;
      default:
        return LogLevel.WARN;
    }
  }

  /**
   * Convert LogLevel enum to string
   */
  private levelToString(level: LogLevel): LogLevelString {
    switch (level) {
      case LogLevel.ERROR:
        return 'ERROR';
      case LogLevel.WARN:
        return 'WARN';
      case LogLevel.INFO:
        return 'INFO';
      case LogLevel.DEBUG:
        return 'DEBUG';
      case LogLevel.TRACE:
        return 'TRACE';
    }
  }

  /**
   * Handle localStorage changes (e.g., from Debug Panel)
   */
  private handleStorageChange = (e: StorageEvent) => {
    if (e.key === this.STORAGE_KEY && e.newValue) {
      this.config.level = this.parseLogLevel(e.newValue);
    }
  };

  /**
   * Set log level at runtime
   */
  public setLevel(level: LogLevel | LogLevelString): void {
    const newLevel = typeof level === 'string' ? this.parseLogLevel(level) : level;
    this.config.level = newLevel;
    localStorage.setItem(this.STORAGE_KEY, this.levelToString(newLevel));
  }

  /**
   * Get current log level
   */
  public getLevel(): LogLevel {
    return this.config.level;
  }

  /**
   * Get current log level as string
   */
  public getLevelString(): LogLevelString {
    return this.levelToString(this.config.level);
  }

  /**
   * Check if a log level is enabled
   */
  private shouldLog(level: LogLevel): boolean {
    return level <= this.config.level;
  }

  /**
   * Format timestamp
   */
  private getTimestamp(): string {
    const now = new Date();
    const hours = now.getHours().toString().padStart(2, '0');
    const minutes = now.getMinutes().toString().padStart(2, '0');
    const seconds = now.getSeconds().toString().padStart(2, '0');
    const ms = now.getMilliseconds().toString().padStart(3, '0');
    return `${hours}:${minutes}:${seconds}.${ms}`;
  }

  /**
   * Format log message with context
   */
  private formatMessage(
    level: LogLevelString,
    context: string,
    message: string
  ): string {
    const timestamp = this.config.enableTimestamps ? `[${this.getTimestamp()}]` : '';
    const ctx = context ? `[${context}]` : '';
    return `${timestamp} [${level}] ${ctx} ${message}`.trim();
  }

  /**
   * Log with color styling
   */
  private logWithStyle(
    level: LogLevelString,
    context: string,
    message: string,
    data?: unknown,
    consoleMethod: 'log' | 'warn' | 'error' = 'log'
  ): void {
    if (!this.config.enableColors || !import.meta.env.DEV) {
      // Simple logging without colors (production)
      const formatted = this.formatMessage(level, context, message);
      if (data !== undefined) {
        console[consoleMethod](formatted, data);
      } else {
        console[consoleMethod](formatted);
      }
      return;
    }

    // Colored logging (development)
    const timestamp = this.config.enableTimestamps ? `[${this.getTimestamp()}]` : '';
    const ctx = context ? `[${context}]` : '';

    const style = `
      background: ${this.bgColors[level]};
      color: ${this.colors[level]};
      font-weight: bold;
      padding: 2px 6px;
      border-radius: 3px;
    `;

    const parts = [
      timestamp && `%c${timestamp}%c`,
      `%c${level}%c`,
      ctx && `%c${ctx}%c`,
      message,
    ].filter(Boolean);

    const styles = [
      timestamp && 'color: #6b7280; font-weight: normal',
      timestamp && '',
      style,
      '',
      ctx && 'color: #8b5cf6; font-weight: bold',
      ctx && '',
    ].filter(Boolean);

    if (data !== undefined) {
      console[consoleMethod](parts.join(' '), ...styles, data);
    } else {
      console[consoleMethod](parts.join(' '), ...styles);
    }
  }

  /**
   * ERROR level logging
   */
  public error(contextOrMessage: string, messageOrData?: string | unknown, data?: unknown): void {
    if (!this.shouldLog(LogLevel.ERROR)) return;

    if (typeof messageOrData === 'string') {
      this.logWithStyle('ERROR', contextOrMessage, messageOrData, data, 'error');
    } else {
      this.logWithStyle('ERROR', '', contextOrMessage, messageOrData, 'error');
    }
  }

  /**
   * WARN level logging
   */
  public warn(contextOrMessage: string, messageOrData?: string | unknown, data?: unknown): void {
    if (!this.shouldLog(LogLevel.WARN)) return;

    if (typeof messageOrData === 'string') {
      this.logWithStyle('WARN', contextOrMessage, messageOrData, data, 'warn');
    } else {
      this.logWithStyle('WARN', '', contextOrMessage, messageOrData, 'warn');
    }
  }

  /**
   * INFO level logging
   */
  public info(contextOrMessage: string, messageOrData?: string | unknown, data?: unknown): void {
    if (!this.shouldLog(LogLevel.INFO)) return;

    if (typeof messageOrData === 'string') {
      this.logWithStyle('INFO', contextOrMessage, messageOrData, data);
    } else {
      this.logWithStyle('INFO', '', contextOrMessage, messageOrData);
    }
  }

  /**
   * DEBUG level logging
   */
  public debug(contextOrMessage: string, messageOrData?: string | unknown, data?: unknown): void {
    if (!this.shouldLog(LogLevel.DEBUG)) return;

    if (typeof messageOrData === 'string') {
      this.logWithStyle('DEBUG', contextOrMessage, messageOrData, data);
    } else {
      this.logWithStyle('DEBUG', '', contextOrMessage, messageOrData);
    }
  }

  /**
   * TRACE level logging
   */
  public trace(contextOrMessage: string, messageOrData?: string | unknown, data?: unknown): void {
    if (!this.shouldLog(LogLevel.TRACE)) return;

    if (typeof messageOrData === 'string') {
      this.logWithStyle('TRACE', contextOrMessage, messageOrData, data);
    } else {
      this.logWithStyle('TRACE', '', contextOrMessage, messageOrData);
    }
  }

  /**
   * Clear console
   */
  public clear(): void {
    console.clear();
  }
}

// Export singleton instance
export const logger = new Logger();

// Export type for external use
export type { Logger };
