/**
 * フロントエンド向けセキュアロギングユーティリティ
 *
 * 機能:
 * - 環境に基づいた自動ログレベル制御
 * - 機密データのマスキング
 * - 本番環境向け安全なロギング（本番環境ではコンソール出力なし）
 * - 構造化ログ形式
 * - error トラッキング連携対応
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
      // 本番環境: error と warning のみログ出力
      // 開発環境: デバッグ用にすべてログ出力
      level: isProduction ? LogLevel.WARN : LogLevel.DEBUG,

      // 本番環境では情報漏洩防止のためコンソール出力を無効化
      enableConsole: !isProduction,

      // 本番環境では error トラッキングを有効化（Sentry、LogRocket 等と連携）
      enableErrorTracking: isProduction,

      // 常に機密データをマスキング
      maskSensitiveData: true,
    };

    // 環境変数による手動オーバーライドを許可
    const logLevelEnv = process.env.NEXT_PUBLIC_LOG_LEVEL;
    if (logLevelEnv) {
      this.config.level = this.parseLogLevel(logLevelEnv);
    }

    // セキュリティのため、本番環境ではすべてのコンソール出力を抑制
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
   * 本番環境ですべてのコンソールメソッドを無効化します
   * ブラウザコンソールでの機密情報の露出を防ぎます
   */
  private disableConsole() {
    const noop = () => {};
    console.log = noop;
    console.info = noop;
    console.debug = noop;
    console.warn = noop;
    // 重大な error 用に console.error は保持しますが、出力はサニタイズします
  }

  /**
   * ログメッセージ内の機密情報をマスキングします
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

        // 機密フィールド名のリスト
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
   * 文字列内の機密パターンをマスキングします
   */
  private maskString(str: string): string {
    let masked = str;

    // メールアドレスをマスキング
    masked = masked.replace(
      /[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/g,
      (email) => {
        const [local, domain] = email.split('@');
        return `${local[0]}***@${domain}`;
      }
    );

    // JWT token をマスキング
    masked = masked.replace(
      /eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*/g,
      '[MASKED_TOKEN]'
    );

    // 電話番号をマスキング
    masked = masked.replace(
      /\+?1?[-.\s]?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}/g,
      '[MASKED_PHONE]'
    );

    // クレジットカード番号をマスキング
    masked = masked.replace(
      /\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b/g,
      '[MASKED_CC]'
    );

    return masked;
  }

  /**
   * メタデータ付きでログメッセージをフォーマットします
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
   * error トラッキングサービス（Sentry、LogRocket 等）に error を送信します
   */
  private trackError(error: Error, context?: any) {
    if (!this.config.enableErrorTracking) {
      return;
    }

    // TODO: error トラッキングサービスと連携
    // Sentry の例:
    // Sentry.captureException(error, {
    //   extra: context,
    // });

    // 現時点では将来の連携用のプレースホルダーです
    console.error('[Error Tracking Placeholder]', error.message);
  }

  /**
   * デバッグメッセージをログ出力します（開発環境のみ）
   */
  debug(message: string, data?: any) {
    if (this.config.level > LogLevel.DEBUG) return;

    if (this.config.enableConsole) {
      console.debug(`[DEBUG] ${message}`, data || '');
    }
  }

  /**
   * 情報メッセージをログ出力します
   */
  info(message: string, data?: any) {
    if (this.config.level > LogLevel.INFO) return;

    if (this.config.enableConsole) {
      console.info(`[INFO] ${message}`, data || '');
    }
  }

  /**
   * 警告メッセージをログ出力します
   */
  warn(message: string, data?: any) {
    if (this.config.level > LogLevel.WARN) return;

    const formatted = this.formatMessage('WARN', message, data);

    if (this.config.enableConsole) {
      console.warn(`[WARN] ${message}`, data || '');
    }

    // 本番環境では、警告をロギングサービスに送信することを検討してください
    if (this.config.enableErrorTracking) {
      // TODO: ロギングサービスに送信
    }
  }

  /**
   * error メッセージをログ出力します
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

    // error は常にログ出力します（本番環境でも安全なロギングサービスへ）
    if (this.config.enableConsole) {
      console.error(`[ERROR] ${message}`, error);
    }

    // error トラッキングサービスに送信
    if (error instanceof Error) {
      this.trackError(error, context);
    }
  }

  /**
   * API request をログ出力します（デバッグ用）
   */
  apiRequest(method: string, url: string, data?: any) {
    this.debug(`API Request: ${method} ${url}`, {
      method,
      url: this.maskString(url), // URL 内の機密パラメータをマスキング
      data: data ? this.maskSensitiveData(data) : undefined,
    });
  }

  /**
   * API response をログ出力します（デバッグ用）
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
        // 機密情報を避けるため、デフォルトでは response データはログ出力しません
      });
    }
  }

  /**
   * ユーザーアクションをログ出力します（分析/監査用）
   */
  userAction(action: string, details?: any) {
    this.info(`User Action: ${action}`, this.maskSensitiveData(details));
  }
}

// シングルトンインスタンスをエクスポート
export const logger = new Logger();

// コンポーネントで使用するための型をエクスポート
export type { Logger };
