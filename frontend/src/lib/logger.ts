/**
 * Secure Logging Utility for Frontend
 *
 * Features:
 * - Automatic log level control based on environment
 * - Sensitive data masking
 * - Production-safe logging (no console output in production)
 * - Structured logging format
 * - Error tracking integration ready
 */

export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
  NONE = 4,
}

interface LoggerConfig {
  level: LogLevel;
  enableConsole: boolean;
  enableErrorTracking: boolean;
  maskSensitiveData: boolean;
}

class Logger {
  private config: LoggerConfig;

  constructor() {
    const isProduction = process.env.NODE_ENV === 'production';

    this.config = {
      // In production: only log errors and warnings
      // In development: log everything for debugging
      level: isProduction ? LogLevel.WARN : LogLevel.DEBUG,

      // Disable console output in production to prevent information leakage
      enableConsole: !isProduction,

      // Enable error tracking in production (integrate with Sentry, LogRocket, etc.)
      enableErrorTracking: isProduction,

      // Always mask sensitive data
      maskSensitiveData: true,
    };

    // Allow manual override via environment variable
    const logLevelEnv = process.env.NEXT_PUBLIC_LOG_LEVEL;
    if (logLevelEnv) {
      this.config.level = this.parseLogLevel(logLevelEnv);
    }

    // Suppress all console output in production for security
    if (isProduction && this.config.enableConsole === false) {
      this.disableConsole();
    }
  }

  private parseLogLevel(level: string): LogLevel {
    switch (level.toLowerCase()) {
      case 'debug':
        return LogLevel.DEBUG;
      case 'info':
        return LogLevel.INFO;
      case 'warn':
      case 'warning':
        return LogLevel.WARN;
      case 'error':
        return LogLevel.ERROR;
      case 'none':
        return LogLevel.NONE;
      default:
        return LogLevel.INFO;
    }
  }

  /**
   * Disable all console methods in production
   * This prevents sensitive information from being exposed in browser console
   */
  private disableConsole() {
    const noop = () => {};
    console.log = noop;
    console.info = noop;
    console.debug = noop;
    console.warn = noop;
    // Keep console.error for critical errors, but sanitize the output
  }

  /**
   * Mask sensitive information in log messages
   */
  private maskSensitiveData(data: any): any {
    if (typeof data === 'string') {
      return this.maskString(data);
    }

    if (Array.isArray(data)) {
      return data.map(item => this.maskSensitiveData(item));
    }

    if (typeof data === 'object' && data !== null) {
      const masked: any = {};
      for (const [key, value] of Object.entries(data)) {
        const lowerKey = key.toLowerCase();

        // List of sensitive field names
        const sensitiveKeys = [
          'password', 'token', 'secret', 'api_key', 'apikey',
          'authorization', 'auth', 'credential', 'ssn',
          'credit_card', 'creditcard', 'cvv', 'pin',
          'access_token', 'refresh_token', 'jwt',
        ];

        if (sensitiveKeys.some(sk => lowerKey.includes(sk))) {
          masked[key] = '***REDACTED***';
        } else if (typeof value === 'object') {
          masked[key] = this.maskSensitiveData(value);
        } else {
          masked[key] = value;
        }
      }
      return masked;
    }

    return data;
  }

  /**
   * Mask sensitive patterns in strings
   */
  private maskString(str: string): string {
    let masked = str;

    // Mask email addresses
    masked = masked.replace(
      /[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/g,
      (email) => {
        const [local, domain] = email.split('@');
        return `${local[0]}***@${domain}`;
      }
    );

    // Mask JWT tokens
    masked = masked.replace(
      /eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*/g,
      '[MASKED_TOKEN]'
    );

    // Mask phone numbers
    masked = masked.replace(
      /\+?1?[-.\s]?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}/g,
      '[MASKED_PHONE]'
    );

    // Mask credit card numbers
    masked = masked.replace(
      /\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b/g,
      '[MASKED_CC]'
    );

    return masked;
  }

  /**
   * Format log message with metadata
   */
  private formatMessage(level: string, message: string, data?: any): string {
    const timestamp = new Date().toISOString();
    const maskedData = this.config.maskSensitiveData && data
      ? this.maskSensitiveData(data)
      : data;

    return JSON.stringify({
      timestamp,
      level,
      message: this.config.maskSensitiveData ? this.maskString(message) : message,
      ...(maskedData && { data: maskedData }),
      service: 'incidex-frontend',
      environment: process.env.NODE_ENV,
    });
  }

  /**
   * Send error to error tracking service (Sentry, LogRocket, etc.)
   */
  private trackError(error: Error, context?: any) {
    if (!this.config.enableErrorTracking) {
      return;
    }

    // TODO: Integrate with error tracking service
    // Example for Sentry:
    // Sentry.captureException(error, {
    //   extra: context,
    // });

    // For now, this is a placeholder for future integration
    console.error('[Error Tracking Placeholder]', error.message);
  }

  /**
   * Log debug message (development only)
   */
  debug(message: string, data?: any) {
    if (this.config.level > LogLevel.DEBUG) return;

    if (this.config.enableConsole) {
      console.debug(`[DEBUG] ${message}`, data || '');
    }
  }

  /**
   * Log informational message
   */
  info(message: string, data?: any) {
    if (this.config.level > LogLevel.INFO) return;

    if (this.config.enableConsole) {
      console.info(`[INFO] ${message}`, data || '');
    }
  }

  /**
   * Log warning message
   */
  warn(message: string, data?: any) {
    if (this.config.level > LogLevel.WARN) return;

    const formatted = this.formatMessage('WARN', message, data);

    if (this.config.enableConsole) {
      console.warn(`[WARN] ${message}`, data || '');
    }

    // In production, you might want to send warnings to a logging service
    if (this.config.enableErrorTracking) {
      // TODO: Send to logging service
    }
  }

  /**
   * Log error message
   */
  error(message: string, error?: Error | any, context?: any) {
    if (this.config.level > LogLevel.ERROR) return;

    const errorData = error instanceof Error ? {
      name: error.name,
      message: error.message,
      stack: error.stack,
      ...context,
    } : { error, ...context };

    const formatted = this.formatMessage('ERROR', message, errorData);

    // Always log errors, even in production (to a secure logging service)
    if (this.config.enableConsole) {
      console.error(`[ERROR] ${message}`, error);
    }

    // Send to error tracking service
    if (error instanceof Error) {
      this.trackError(error, context);
    }
  }

  /**
   * Log API request (for debugging)
   */
  apiRequest(method: string, url: string, data?: any) {
    this.debug(`API Request: ${method} ${url}`, {
      method,
      url: this.maskString(url), // Mask sensitive params in URL
      data: data ? this.maskSensitiveData(data) : undefined,
    });
  }

  /**
   * Log API response (for debugging)
   */
  apiResponse(method: string, url: string, status: number, data?: any) {
    const level = status >= 400 ? 'error' : 'debug';
    const message = `API Response: ${method} ${url} - ${status}`;

    if (level === 'error') {
      this.error(message, undefined, { method, url, status });
    } else {
      this.debug(message, {
        method,
        url: this.maskString(url),
        status,
        // Don't log response data by default to avoid sensitive info
      });
    }
  }

  /**
   * Log user action (for analytics/audit)
   */
  userAction(action: string, details?: any) {
    this.info(`User Action: ${action}`, this.maskSensitiveData(details));
  }
}

// Export singleton instance
export const logger = new Logger();

// Export type for use in components
export type { Logger };
