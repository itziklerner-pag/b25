// Simple structured logger for the authentication service

const LogLevel = {
  ERROR: 'ERROR',
  WARN: 'WARN',
  INFO: 'INFO',
  DEBUG: 'DEBUG',
};

class Logger {
  constructor(context = 'AuthService') {
    this.context = context;
  }

  log(level, message, meta) {
    const timestamp = new Date().toISOString();
    const logEntry = {
      timestamp,
      level,
      context: this.context,
      message,
      ...(meta && { meta }),
    };

    const logString = JSON.stringify(logEntry);

    switch (level) {
      case LogLevel.ERROR:
        console.error(logString);
        break;
      case LogLevel.WARN:
        console.warn(logString);
        break;
      case LogLevel.INFO:
        console.info(logString);
        break;
      case LogLevel.DEBUG:
        console.debug(logString);
        break;
    }
  }

  error(message, meta) {
    this.log(LogLevel.ERROR, message, meta);
  }

  warn(message, meta) {
    this.log(LogLevel.WARN, message, meta);
  }

  info(message, meta) {
    this.log(LogLevel.INFO, message, meta);
  }

  debug(message, meta) {
    this.log(LogLevel.DEBUG, message, meta);
  }

  child(context) {
    return new Logger(`${this.context}:${context}`);
  }
}

export { LogLevel };
export const logger = new Logger('AuthService');
export default logger;
