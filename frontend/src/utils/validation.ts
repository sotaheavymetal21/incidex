/**
 * Frontend Validation Utilities
 * 共通のバリデーション関数とルール
 */

// ============================================
// Validation Constants
// ============================================

export const ValidationLimits = {
  // User fields
  NAME_MAX_LENGTH: 50,
  EMAIL_MAX_LENGTH: 254,
  PASSWORD_MIN_LENGTH: 8,
  PASSWORD_MIN_LENGTH_ADMIN: 6,
  EMPLOYEE_NUMBER_MAX_LENGTH: 20,
  DEPARTMENT_MAX_LENGTH: 50,

  // Incident fields
  TITLE_MAX_LENGTH: 500,
  DESCRIPTION_MAX_LENGTH: 10000,
  IMPACT_SCOPE_MAX_LENGTH: 500,

  // Comment/Activity fields
  COMMENT_MIN_LENGTH: 1,
  COMMENT_MAX_LENGTH: 5000,

  // Tag fields
  TAG_NAME_MAX_LENGTH: 50,
  TAG_COLOR_LENGTH: 7, // #RRGGBB

  // Post-mortem fields
  POSTMORTEM_TEXT_MAX_LENGTH: 10000,
  FIVE_WHYS_MAX_LENGTH: 1000,

  // Action Item fields
  ACTION_ITEM_TITLE_MIN_LENGTH: 1,
  ACTION_ITEM_TITLE_MAX_LENGTH: 500,
  ACTION_ITEM_DESCRIPTION_MAX_LENGTH: 5000,
  ACTION_ITEM_LINKS_MAX_LENGTH: 2000,

  // Notification settings
  SLACK_WEBHOOK_MAX_LENGTH: 500,
} as const;

// ============================================
// Validation Helper Functions
// ============================================

/**
 * Check if string contains emoji characters
 */
export const containsEmoji = (str: string): boolean => {
  const emojiRegex = /[\u{1F000}-\u{1F9FF}]|[\u{2600}-\u{27BF}]|[\u{1F600}-\u{1F64F}]|[\u{1F680}-\u{1F6FF}]|[\u{2702}-\u{27B0}]|[\u{FE00}-\u{FE0F}]|[\u{1F300}-\u{1F5FF}]|[\u{1F900}-\u{1F9FF}]|[\u{1FA00}-\u{1FA6F}]|[\u{1FA70}-\u{1FAFF}]|[\u{2300}-\u{23FF}]/u;
  return emojiRegex.test(str);
};

/**
 * Validate employee number format (alphanumeric and hyphens only)
 */
export const isValidEmployeeNumber = (str: string): boolean => {
  if (!str) return true;
  return /^[a-zA-Z0-9\-]*$/.test(str);
};

/**
 * Validate email format
 */
export const isValidEmail = (str: string): boolean => {
  if (!str) return false;
  return /^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$/.test(str);
};

/**
 * Validate password strength
 */
export const validatePasswordStrength = (password: string): {
  isValid: boolean;
  errors: string[];
} => {
  const errors: string[] = [];

  if (password.length < ValidationLimits.PASSWORD_MIN_LENGTH) {
    errors.push(`パスワードは${ValidationLimits.PASSWORD_MIN_LENGTH}文字以上である必要があります`);
  }
  if (!/[A-Z]/.test(password)) {
    errors.push('パスワードには大文字を含める必要があります');
  }
  if (!/[a-z]/.test(password)) {
    errors.push('パスワードには小文字を含める必要があります');
  }
  if (!/[0-9]/.test(password)) {
    errors.push('パスワードには数字を含める必要があります');
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
};

/**
 * Validate hex color format
 */
export const isValidHexColor = (str: string): boolean => {
  return /^#[0-9A-Fa-f]{6}$/.test(str);
};

/**
 * Validate URL format
 */
export const isValidUrl = (str: string): boolean => {
  if (!str) return true;
  try {
    new URL(str);
    return true;
  } catch {
    return false;
  }
};

/**
 * Validate Slack webhook URL format
 */
export const isValidSlackWebhook = (str: string): boolean => {
  if (!str) return true;
  return /^https:\/\/hooks\.slack\.com\/services\/[A-Z0-9]+\/[A-Z0-9]+\/[a-zA-Z0-9]+$/.test(str);
};

/**
 * Check for potentially dangerous characters (XSS prevention)
 */
export const containsDangerousChars = (str: string): boolean => {
  // Check for script tags, event handlers, and other XSS vectors
  const dangerousPatterns = [
    /<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi,
    /on\w+\s*=/gi,
    /javascript:/gi,
    /data:/gi,
    /vbscript:/gi,
  ];
  return dangerousPatterns.some((pattern) => pattern.test(str));
};

/**
 * Sanitize input by removing dangerous characters
 */
export const sanitizeInput = (str: string): string => {
  return str
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#x27;')
    .replace(/\//g, '&#x2F;');
};

// ============================================
// Field-specific Validation Functions
// ============================================

export interface ValidationResult {
  isValid: boolean;
  error?: string;
}

/**
 * Validate name field
 */
export const validateName = (value: string, required = true): ValidationResult => {
  if (required && !value.trim()) {
    return { isValid: false, error: '名前は必須です' };
  }
  if (value.length > ValidationLimits.NAME_MAX_LENGTH) {
    return { isValid: false, error: `名前は${ValidationLimits.NAME_MAX_LENGTH}文字以内で入力してください` };
  }
  if (containsEmoji(value)) {
    return { isValid: false, error: '名前に絵文字は使用できません' };
  }
  if (containsDangerousChars(value)) {
    return { isValid: false, error: '名前に使用できない文字が含まれています' };
  }
  return { isValid: true };
};

/**
 * Validate email field
 */
export const validateEmail = (value: string, required = true): ValidationResult => {
  if (required && !value.trim()) {
    return { isValid: false, error: 'メールアドレスは必須です' };
  }
  if (value.length > ValidationLimits.EMAIL_MAX_LENGTH) {
    return { isValid: false, error: `メールアドレスは${ValidationLimits.EMAIL_MAX_LENGTH}文字以内で入力してください` };
  }
  if (value && !isValidEmail(value)) {
    return { isValid: false, error: '有効なメールアドレスを入力してください' };
  }
  return { isValid: true };
};

/**
 * Validate password field
 */
export const validatePassword = (value: string, requireStrong = true): ValidationResult => {
  if (!value) {
    return { isValid: false, error: 'パスワードは必須です' };
  }

  if (requireStrong) {
    const { isValid, errors } = validatePasswordStrength(value);
    if (!isValid) {
      return { isValid: false, error: errors[0] };
    }
  } else {
    if (value.length < ValidationLimits.PASSWORD_MIN_LENGTH_ADMIN) {
      return { isValid: false, error: `パスワードは${ValidationLimits.PASSWORD_MIN_LENGTH_ADMIN}文字以上である必要があります` };
    }
  }

  return { isValid: true };
};

/**
 * Validate employee number field
 */
export const validateEmployeeNumber = (value: string): ValidationResult => {
  if (value && value.length > ValidationLimits.EMPLOYEE_NUMBER_MAX_LENGTH) {
    return { isValid: false, error: `社員番号は${ValidationLimits.EMPLOYEE_NUMBER_MAX_LENGTH}文字以内で入力してください` };
  }
  if (value && !isValidEmployeeNumber(value)) {
    return { isValid: false, error: '社員番号は英数字とハイフンのみ使用できます' };
  }
  return { isValid: true };
};

/**
 * Validate department field
 */
export const validateDepartment = (value: string): ValidationResult => {
  if (value && value.length > ValidationLimits.DEPARTMENT_MAX_LENGTH) {
    return { isValid: false, error: `所属部署は${ValidationLimits.DEPARTMENT_MAX_LENGTH}文字以内で入力してください` };
  }
  if (value && containsEmoji(value)) {
    return { isValid: false, error: '所属部署に絵文字は使用できません' };
  }
  return { isValid: true };
};

/**
 * Validate incident title field
 */
export const validateIncidentTitle = (value: string): ValidationResult => {
  if (!value.trim()) {
    return { isValid: false, error: 'タイトルは必須です' };
  }
  if (value.length > ValidationLimits.TITLE_MAX_LENGTH) {
    return { isValid: false, error: `タイトルは${ValidationLimits.TITLE_MAX_LENGTH}文字以内で入力してください` };
  }
  if (containsDangerousChars(value)) {
    return { isValid: false, error: 'タイトルに使用できない文字が含まれています' };
  }
  return { isValid: true };
};

/**
 * Validate incident description field
 */
export const validateIncidentDescription = (value: string): ValidationResult => {
  if (!value.trim()) {
    return { isValid: false, error: '説明は必須です' };
  }
  if (value.length > ValidationLimits.DESCRIPTION_MAX_LENGTH) {
    return { isValid: false, error: `説明は${ValidationLimits.DESCRIPTION_MAX_LENGTH}文字以内で入力してください` };
  }
  return { isValid: true };
};

/**
 * Validate impact scope field
 */
export const validateImpactScope = (value: string): ValidationResult => {
  if (value && value.length > ValidationLimits.IMPACT_SCOPE_MAX_LENGTH) {
    return { isValid: false, error: `影響範囲は${ValidationLimits.IMPACT_SCOPE_MAX_LENGTH}文字以内で入力してください` };
  }
  return { isValid: true };
};

/**
 * Validate comment field
 */
export const validateComment = (value: string): ValidationResult => {
  if (!value.trim()) {
    return { isValid: false, error: 'コメントは必須です' };
  }
  if (value.length < ValidationLimits.COMMENT_MIN_LENGTH) {
    return { isValid: false, error: 'コメントを入力してください' };
  }
  if (value.length > ValidationLimits.COMMENT_MAX_LENGTH) {
    return { isValid: false, error: `コメントは${ValidationLimits.COMMENT_MAX_LENGTH}文字以内で入力してください` };
  }
  return { isValid: true };
};

/**
 * Validate tag name field
 */
export const validateTagName = (value: string): ValidationResult => {
  if (!value.trim()) {
    return { isValid: false, error: 'タグ名は必須です' };
  }
  if (value.length > ValidationLimits.TAG_NAME_MAX_LENGTH) {
    return { isValid: false, error: `タグ名は${ValidationLimits.TAG_NAME_MAX_LENGTH}文字以内で入力してください` };
  }
  if (containsEmoji(value)) {
    return { isValid: false, error: 'タグ名に絵文字は使用できません' };
  }
  if (containsDangerousChars(value)) {
    return { isValid: false, error: 'タグ名に使用できない文字が含まれています' };
  }
  return { isValid: true };
};

/**
 * Validate tag color field
 */
export const validateTagColor = (value: string): ValidationResult => {
  if (!value) {
    return { isValid: false, error: '色は必須です' };
  }
  if (!isValidHexColor(value)) {
    return { isValid: false, error: '有効な色コード（例: #10b981）を入力してください' };
  }
  return { isValid: true };
};

/**
 * Validate post-mortem text fields
 */
export const validatePostMortemText = (value: string, fieldName: string): ValidationResult => {
  if (value && value.length > ValidationLimits.POSTMORTEM_TEXT_MAX_LENGTH) {
    return { isValid: false, error: `${fieldName}は${ValidationLimits.POSTMORTEM_TEXT_MAX_LENGTH}文字以内で入力してください` };
  }
  return { isValid: true };
};

/**
 * Validate action item title field
 */
export const validateActionItemTitle = (value: string): ValidationResult => {
  if (!value.trim()) {
    return { isValid: false, error: 'タイトルは必須です' };
  }
  if (value.length < ValidationLimits.ACTION_ITEM_TITLE_MIN_LENGTH) {
    return { isValid: false, error: 'タイトルを入力してください' };
  }
  if (value.length > ValidationLimits.ACTION_ITEM_TITLE_MAX_LENGTH) {
    return { isValid: false, error: `タイトルは${ValidationLimits.ACTION_ITEM_TITLE_MAX_LENGTH}文字以内で入力してください` };
  }
  return { isValid: true };
};

/**
 * Validate action item description field
 */
export const validateActionItemDescription = (value: string): ValidationResult => {
  if (value && value.length > ValidationLimits.ACTION_ITEM_DESCRIPTION_MAX_LENGTH) {
    return { isValid: false, error: `説明は${ValidationLimits.ACTION_ITEM_DESCRIPTION_MAX_LENGTH}文字以内で入力してください` };
  }
  return { isValid: true };
};

/**
 * Validate action item related links field
 */
export const validateRelatedLinks = (value: string): ValidationResult => {
  if (value && value.length > ValidationLimits.ACTION_ITEM_LINKS_MAX_LENGTH) {
    return { isValid: false, error: `関連リンクは${ValidationLimits.ACTION_ITEM_LINKS_MAX_LENGTH}文字以内で入力してください` };
  }
  // Validate each URL in comma-separated list
  if (value) {
    const links = value.split(',').map((link) => link.trim()).filter(Boolean);
    for (const link of links) {
      const urlToCheck = link.startsWith('http') ? link : `https://${link}`;
      if (!isValidUrl(urlToCheck)) {
        return { isValid: false, error: `無効なURL形式です: ${link}` };
      }
    }
  }
  return { isValid: true };
};

/**
 * Validate Slack webhook URL field
 */
export const validateSlackWebhookUrl = (value: string): ValidationResult => {
  if (value && value.length > ValidationLimits.SLACK_WEBHOOK_MAX_LENGTH) {
    return { isValid: false, error: `Webhook URLは${ValidationLimits.SLACK_WEBHOOK_MAX_LENGTH}文字以内で入力してください` };
  }
  if (value && !isValidSlackWebhook(value)) {
    return { isValid: false, error: '有効なSlack Webhook URLを入力してください（例: https://hooks.slack.com/services/...）' };
  }
  return { isValid: true };
};

/**
 * Validate datetime field (must not be in future for certain fields)
 */
export const validateDatetime = (value: string, allowFuture = true, fieldName = '日時'): ValidationResult => {
  if (!value) {
    return { isValid: false, error: `${fieldName}は必須です` };
  }

  const date = new Date(value);
  if (isNaN(date.getTime())) {
    return { isValid: false, error: `有効な${fieldName}を入力してください` };
  }

  if (!allowFuture && date > new Date()) {
    return { isValid: false, error: `${fieldName}は未来の日時に設定できません` };
  }

  return { isValid: true };
};

// ============================================
// Severity and Status Validation
// ============================================

export const VALID_SEVERITIES = ['low', 'medium', 'high', 'critical'] as const;
export const VALID_STATUSES = ['open', 'investigating', 'resolved', 'closed'] as const;
export const VALID_PRIORITIES = ['low', 'medium', 'high'] as const;
export const VALID_ACTION_STATUSES = ['pending', 'in_progress', 'completed'] as const;
export const VALID_TIMELINE_EVENT_TYPES = [
  'detected',
  'investigation_started',
  'root_cause_identified',
  'mitigation',
  'timeline_resolved',
  'other',
] as const;

export type Severity = (typeof VALID_SEVERITIES)[number];
export type Status = (typeof VALID_STATUSES)[number];
export type Priority = (typeof VALID_PRIORITIES)[number];
export type ActionStatus = (typeof VALID_ACTION_STATUSES)[number];
export type TimelineEventType = (typeof VALID_TIMELINE_EVENT_TYPES)[number];

export const validateSeverity = (value: string): ValidationResult => {
  if (!VALID_SEVERITIES.includes(value as Severity)) {
    return { isValid: false, error: '有効な重要度を選択してください' };
  }
  return { isValid: true };
};

export const validateStatus = (value: string): ValidationResult => {
  if (!VALID_STATUSES.includes(value as Status)) {
    return { isValid: false, error: '有効なステータスを選択してください' };
  }
  return { isValid: true };
};

export const validatePriority = (value: string): ValidationResult => {
  if (!VALID_PRIORITIES.includes(value as Priority)) {
    return { isValid: false, error: '有効な優先度を選択してください' };
  }
  return { isValid: true };
};

export const validateActionStatus = (value: string): ValidationResult => {
  if (!VALID_ACTION_STATUSES.includes(value as ActionStatus)) {
    return { isValid: false, error: '有効なステータスを選択してください' };
  }
  return { isValid: true };
};

export const validateTimelineEventType = (value: string): ValidationResult => {
  if (!VALID_TIMELINE_EVENT_TYPES.includes(value as TimelineEventType)) {
    return { isValid: false, error: '有効なイベントタイプを選択してください' };
  }
  return { isValid: true };
};
