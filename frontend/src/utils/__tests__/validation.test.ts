import { describe, it, expect } from 'vitest';
import {
  containsEmoji,
  isValidEmail,
  isValidEmployeeNumber,
  validatePasswordStrength,
  containsDangerousChars,
  isValidHexColor,
  isValidUrl,
  isValidSlackWebhook,
  validateName,
  validateEmail,
  validatePassword,
  validateEmployeeNumber,
  validateDepartment,
  validateIncidentTitle,
  validateSeverity,
  validateStatus,
  ValidationLimits,
} from '../validation';

describe('containsEmoji', () => {
  it('returns true for string with emoji face', () => {
    expect(containsEmoji('Hello 😀')).toBe(true);
  });

  it('returns true for string with emoji sun', () => {
    expect(containsEmoji('☀️ Sunny')).toBe(true);
  });

  it('returns false for normal text', () => {
    expect(containsEmoji('Hello World')).toBe(false);
  });

  it('returns false for Japanese text', () => {
    expect(containsEmoji('こんにちは')).toBe(false);
  });

  it('returns false for empty string', () => {
    expect(containsEmoji('')).toBe(false);
  });
});

describe('isValidEmail', () => {
  it('returns true for valid email', () => {
    expect(isValidEmail('user@example.com')).toBe(true);
  });

  it('returns true for email with subdomain', () => {
    expect(isValidEmail('user@mail.example.com')).toBe(true);
  });

  it('returns true for email with plus sign', () => {
    expect(isValidEmail('user+tag@example.com')).toBe(true);
  });

  it('returns false for empty string', () => {
    expect(isValidEmail('')).toBe(false);
  });

  it('returns false for string without @', () => {
    expect(isValidEmail('userexample.com')).toBe(false);
  });

  it('returns false for string without domain', () => {
    expect(isValidEmail('user@')).toBe(false);
  });
});

describe('isValidEmployeeNumber', () => {
  it('returns true for valid alphanumeric with hyphen', () => {
    expect(isValidEmployeeNumber('EMP-001')).toBe(true);
  });

  it('returns true for numeric only', () => {
    expect(isValidEmployeeNumber('12345')).toBe(true);
  });

  it('returns true for empty string', () => {
    expect(isValidEmployeeNumber('')).toBe(true);
  });

  it('returns false for string with spaces', () => {
    expect(isValidEmployeeNumber('EMP 001')).toBe(false);
  });

  it('returns false for string with special characters', () => {
    expect(isValidEmployeeNumber('EMP#001')).toBe(false);
  });
});

describe('validatePasswordStrength', () => {
  it('returns valid for strong password', () => {
    const result = validatePasswordStrength('StrongPass123!');
    expect(result.isValid).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  it('returns invalid for short password', () => {
    const result = validatePasswordStrength('Ab1!');
    expect(result.isValid).toBe(false);
    expect(result.errors.some(e => e.includes(ValidationLimits.PASSWORD_MIN_LENGTH.toString()))).toBe(true);
  });

  it('returns invalid for password without uppercase', () => {
    const result = validatePasswordStrength('nouppercaseletter123!');
    expect(result.isValid).toBe(false);
    expect(result.errors.some(e => e.includes('大文字'))).toBe(true);
  });

  it('returns invalid for password without lowercase', () => {
    const result = validatePasswordStrength('NOLOWERCASE123!');
    expect(result.isValid).toBe(false);
    expect(result.errors.some(e => e.includes('小文字'))).toBe(true);
  });

  it('returns invalid for password without number', () => {
    const result = validatePasswordStrength('NoNumberHere!');
    expect(result.isValid).toBe(false);
    expect(result.errors.some(e => e.includes('数字'))).toBe(true);
  });
});

describe('containsDangerousChars', () => {
  it('returns true for script tags', () => {
    expect(containsDangerousChars('<script>alert("xss")</script>')).toBe(true);
  });

  it('returns true for javascript: protocol', () => {
    expect(containsDangerousChars('javascript:alert(1)')).toBe(true);
  });

  it('returns true for event handlers', () => {
    expect(containsDangerousChars('onload=alert(1)')).toBe(true);
  });

  it('returns false for normal text', () => {
    expect(containsDangerousChars('Normal text content')).toBe(false);
  });
});

describe('isValidHexColor', () => {
  it('returns true for valid 6-digit hex', () => {
    expect(isValidHexColor('#10b981')).toBe(true);
  });

  it('returns true for uppercase hex', () => {
    expect(isValidHexColor('#FF5733')).toBe(true);
  });

  it('returns false without hash', () => {
    expect(isValidHexColor('10b981')).toBe(false);
  });

  it('returns false for 3-digit hex', () => {
    expect(isValidHexColor('#fff')).toBe(false);
  });
});

describe('isValidUrl', () => {
  it('returns true for valid http URL', () => {
    expect(isValidUrl('http://example.com')).toBe(true);
  });

  it('returns true for valid https URL', () => {
    expect(isValidUrl('https://example.com/path')).toBe(true);
  });

  it('returns true for empty string', () => {
    expect(isValidUrl('')).toBe(true);
  });

  it('returns false for invalid URL', () => {
    expect(isValidUrl('not-a-url')).toBe(false);
  });
});

describe('isValidSlackWebhook', () => {
  it('returns true for valid Slack webhook', () => {
    // Using mock URL structure to avoid secret detection
    const mockWebhook = 'https://hooks.slack.com/services/TEST123/TEST456/mockTokenForTestingOnly';
    expect(isValidSlackWebhook(mockWebhook)).toBe(true);
  });

  it('returns true for empty string', () => {
    expect(isValidSlackWebhook('')).toBe(true);
  });

  it('returns false for invalid webhook', () => {
    expect(isValidSlackWebhook('https://example.com/webhook')).toBe(false);
  });
});

describe('validateName', () => {
  it('returns valid for normal name', () => {
    const result = validateName('田中太郎');
    expect(result.isValid).toBe(true);
  });

  it('returns invalid for empty name when required', () => {
    const result = validateName('', true);
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('必須');
  });

  it('returns valid for empty name when not required', () => {
    const result = validateName('', false);
    expect(result.isValid).toBe(true);
  });

  it('returns invalid for name with emoji', () => {
    const result = validateName('Test😀User');
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('絵文字');
  });

  it('returns invalid for name exceeding max length', () => {
    const longName = 'a'.repeat(51);
    const result = validateName(longName);
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('50');
  });
});

describe('validateEmail', () => {
  it('returns valid for correct email', () => {
    const result = validateEmail('user@example.com');
    expect(result.isValid).toBe(true);
  });

  it('returns invalid for empty email when required', () => {
    const result = validateEmail('', true);
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('必須');
  });

  it('returns invalid for malformed email', () => {
    const result = validateEmail('invalid-email');
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('有効な');
  });
});

describe('validatePassword', () => {
  it('returns valid for strong password', () => {
    const result = validatePassword('StrongPass123!');
    expect(result.isValid).toBe(true);
  });

  it('returns invalid for empty password', () => {
    const result = validatePassword('');
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('必須');
  });

  it('returns valid for weak password when not requiring strong', () => {
    const result = validatePassword('simple', false);
    expect(result.isValid).toBe(true);
  });
});

describe('validateEmployeeNumber', () => {
  it('returns valid for correct format', () => {
    const result = validateEmployeeNumber('EMP-001');
    expect(result.isValid).toBe(true);
  });

  it('returns valid for empty string', () => {
    const result = validateEmployeeNumber('');
    expect(result.isValid).toBe(true);
  });

  it('returns invalid for special characters', () => {
    const result = validateEmployeeNumber('EMP#001');
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('英数字');
  });
});

describe('validateDepartment', () => {
  it('returns valid for normal department', () => {
    const result = validateDepartment('Engineering');
    expect(result.isValid).toBe(true);
  });

  it('returns valid for empty string', () => {
    const result = validateDepartment('');
    expect(result.isValid).toBe(true);
  });

  it('returns invalid for department with emoji', () => {
    const result = validateDepartment('開発部😀');
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('絵文字');
  });
});

describe('validateIncidentTitle', () => {
  it('returns valid for normal title', () => {
    const result = validateIncidentTitle('Database outage');
    expect(result.isValid).toBe(true);
  });

  it('returns invalid for empty title', () => {
    const result = validateIncidentTitle('');
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('必須');
  });

  it('returns invalid for title with XSS', () => {
    const result = validateIncidentTitle('<script>alert("xss")</script>');
    expect(result.isValid).toBe(false);
  });
});

describe('validateSeverity', () => {
  it('returns valid for critical', () => {
    expect(validateSeverity('critical').isValid).toBe(true);
  });

  it('returns valid for high', () => {
    expect(validateSeverity('high').isValid).toBe(true);
  });

  it('returns valid for medium', () => {
    expect(validateSeverity('medium').isValid).toBe(true);
  });

  it('returns valid for low', () => {
    expect(validateSeverity('low').isValid).toBe(true);
  });

  it('returns invalid for unknown severity', () => {
    expect(validateSeverity('unknown').isValid).toBe(false);
  });
});

describe('validateStatus', () => {
  it('returns valid for open', () => {
    expect(validateStatus('open').isValid).toBe(true);
  });

  it('returns valid for investigating', () => {
    expect(validateStatus('investigating').isValid).toBe(true);
  });

  it('returns valid for resolved', () => {
    expect(validateStatus('resolved').isValid).toBe(true);
  });

  it('returns valid for closed', () => {
    expect(validateStatus('closed').isValid).toBe(true);
  });

  it('returns invalid for unknown status', () => {
    expect(validateStatus('unknown').isValid).toBe(false);
  });
});
