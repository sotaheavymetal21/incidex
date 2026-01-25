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
  it('絵文字の顔を含む文字列の場合は true を返します', () => {
    expect(containsEmoji('Hello 😀')).toBe(true);
  });

  it('太陽の絵文字を含む文字列の場合は true を返します', () => {
    expect(containsEmoji('☀️ Sunny')).toBe(true);
  });

  it('通常のテキストの場合は false を返します', () => {
    expect(containsEmoji('Hello World')).toBe(false);
  });

  it('日本語テキストの場合は false を返します', () => {
    expect(containsEmoji('こんにちは')).toBe(false);
  });

  it('空文字列の場合は false を返します', () => {
    expect(containsEmoji('')).toBe(false);
  });
});

describe('isValidEmail', () => {
  it('有効なメールアドレスの場合は true を返します', () => {
    expect(isValidEmail('user@example.com')).toBe(true);
  });

  it('サブドメイン付きメールアドレスの場合は true を返します', () => {
    expect(isValidEmail('user@mail.example.com')).toBe(true);
  });

  it('プラス記号付きメールアドレスの場合は true を返します', () => {
    expect(isValidEmail('user+tag@example.com')).toBe(true);
  });

  it('空文字列の場合は false を返します', () => {
    expect(isValidEmail('')).toBe(false);
  });

  it('@ なしの文字列の場合は false を返します', () => {
    expect(isValidEmail('userexample.com')).toBe(false);
  });

  it('ドメインなしの文字列の場合は false を返します', () => {
    expect(isValidEmail('user@')).toBe(false);
  });
});

describe('isValidEmployeeNumber', () => {
  it('英数字とハイフンの有効な形式の場合は true を返します', () => {
    expect(isValidEmployeeNumber('EMP-001')).toBe(true);
  });

  it('数字のみの場合は true を返します', () => {
    expect(isValidEmployeeNumber('12345')).toBe(true);
  });

  it('空文字列の場合は true を返します', () => {
    expect(isValidEmployeeNumber('')).toBe(true);
  });

  it('スペースを含む文字列の場合は false を返します', () => {
    expect(isValidEmployeeNumber('EMP 001')).toBe(false);
  });

  it('特殊文字を含む文字列の場合は false を返します', () => {
    expect(isValidEmployeeNumber('EMP#001')).toBe(false);
  });
});

describe('validatePasswordStrength', () => {
  it('強力なパスワードの場合は valid を返します', () => {
    const result = validatePasswordStrength('StrongPass123!');
    expect(result.isValid).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  it('短いパスワードの場合は invalid を返します', () => {
    const result = validatePasswordStrength('Ab1!');
    expect(result.isValid).toBe(false);
    expect(result.errors.some(e => e.includes(ValidationLimits.PASSWORD_MIN_LENGTH.toString()))).toBe(true);
  });

  it('大文字がないパスワードの場合は invalid を返します', () => {
    const result = validatePasswordStrength('nouppercaseletter123!');
    expect(result.isValid).toBe(false);
    expect(result.errors.some(e => e.includes('大文字'))).toBe(true);
  });

  it('小文字がないパスワードの場合は invalid を返します', () => {
    const result = validatePasswordStrength('NOLOWERCASE123!');
    expect(result.isValid).toBe(false);
    expect(result.errors.some(e => e.includes('小文字'))).toBe(true);
  });

  it('数字がないパスワードの場合は invalid を返します', () => {
    const result = validatePasswordStrength('NoNumberHere!');
    expect(result.isValid).toBe(false);
    expect(result.errors.some(e => e.includes('数字'))).toBe(true);
  });
});

describe('containsDangerousChars', () => {
  it('script タグの場合は true を返します', () => {
    expect(containsDangerousChars('<script>alert("xss")</script>')).toBe(true);
  });

  it('javascript: プロトコルの場合は true を返します', () => {
    expect(containsDangerousChars('javascript:alert(1)')).toBe(true);
  });

  it('イベントハンドラーの場合は true を返します', () => {
    expect(containsDangerousChars('onload=alert(1)')).toBe(true);
  });

  it('通常のテキストの場合は false を返します', () => {
    expect(containsDangerousChars('Normal text content')).toBe(false);
  });
});

describe('isValidHexColor', () => {
  it('有効な6桁の16進数の場合は true を返します', () => {
    expect(isValidHexColor('#10b981')).toBe(true);
  });

  it('大文字の16進数の場合は true を返します', () => {
    expect(isValidHexColor('#FF5733')).toBe(true);
  });

  it('ハッシュなしの場合は false を返します', () => {
    expect(isValidHexColor('10b981')).toBe(false);
  });

  it('3桁の16進数の場合は false を返します', () => {
    expect(isValidHexColor('#fff')).toBe(false);
  });
});

describe('isValidUrl', () => {
  it('有効な http URL の場合は true を返します', () => {
    expect(isValidUrl('http://example.com')).toBe(true);
  });

  it('有効な https URL の場合は true を返します', () => {
    expect(isValidUrl('https://example.com/path')).toBe(true);
  });

  it('空文字列の場合は true を返します', () => {
    expect(isValidUrl('')).toBe(true);
  });

  it('無効な URL の場合は false を返します', () => {
    expect(isValidUrl('not-a-url')).toBe(false);
  });
});

describe('isValidSlackWebhook', () => {
  it('有効な Slack Webhook の場合は true を返します', () => {
    // シークレット検出を避けるためモック URL 構造を使用
    const mockWebhook = 'https://hooks.slack.com/services/TEST123/TEST456/mockTokenForTestingOnly';
    expect(isValidSlackWebhook(mockWebhook)).toBe(true);
  });

  it('空文字列の場合は true を返します', () => {
    expect(isValidSlackWebhook('')).toBe(true);
  });

  it('無効な Webhook の場合は false を返します', () => {
    expect(isValidSlackWebhook('https://example.com/webhook')).toBe(false);
  });
});

describe('validateName', () => {
  it('通常の名前の場合は valid を返します', () => {
    const result = validateName('田中太郎');
    expect(result.isValid).toBe(true);
  });

  it('必須で空の名前の場合は invalid を返します', () => {
    const result = validateName('', true);
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('必須');
  });

  it('必須でない場合に空の名前は valid を返します', () => {
    const result = validateName('', false);
    expect(result.isValid).toBe(true);
  });

  it('絵文字を含む名前の場合は invalid を返します', () => {
    const result = validateName('Test😀User');
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('絵文字');
  });

  it('最大長を超える名前の場合は invalid を返します', () => {
    const longName = 'a'.repeat(51);
    const result = validateName(longName);
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('50');
  });
});

describe('validateEmail', () => {
  it('正しいメールアドレスの場合は valid を返します', () => {
    const result = validateEmail('user@example.com');
    expect(result.isValid).toBe(true);
  });

  it('必須で空のメールアドレスの場合は invalid を返します', () => {
    const result = validateEmail('', true);
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('必須');
  });

  it('不正な形式のメールアドレスの場合は invalid を返します', () => {
    const result = validateEmail('invalid-email');
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('有効な');
  });
});

describe('validatePassword', () => {
  it('強力なパスワードの場合は valid を返します', () => {
    const result = validatePassword('StrongPass123!');
    expect(result.isValid).toBe(true);
  });

  it('空のパスワードの場合は invalid を返します', () => {
    const result = validatePassword('');
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('必須');
  });

  it('強度不要の場合に弱いパスワードは valid を返します', () => {
    const result = validatePassword('simple', false);
    expect(result.isValid).toBe(true);
  });
});

describe('validateEmployeeNumber', () => {
  it('正しい形式の場合は valid を返します', () => {
    const result = validateEmployeeNumber('EMP-001');
    expect(result.isValid).toBe(true);
  });

  it('空文字列の場合は valid を返します', () => {
    const result = validateEmployeeNumber('');
    expect(result.isValid).toBe(true);
  });

  it('特殊文字を含む場合は invalid を返します', () => {
    const result = validateEmployeeNumber('EMP#001');
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('英数字');
  });
});

describe('validateDepartment', () => {
  it('通常の部署名の場合は valid を返します', () => {
    const result = validateDepartment('Engineering');
    expect(result.isValid).toBe(true);
  });

  it('空文字列の場合は valid を返します', () => {
    const result = validateDepartment('');
    expect(result.isValid).toBe(true);
  });

  it('絵文字を含む部署名の場合は invalid を返します', () => {
    const result = validateDepartment('開発部😀');
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('絵文字');
  });
});

describe('validateIncidentTitle', () => {
  it('通常のタイトルの場合は valid を返します', () => {
    const result = validateIncidentTitle('Database outage');
    expect(result.isValid).toBe(true);
  });

  it('空のタイトルの場合は invalid を返します', () => {
    const result = validateIncidentTitle('');
    expect(result.isValid).toBe(false);
    expect(result.error).toContain('必須');
  });

  it('XSS を含むタイトルの場合は invalid を返します', () => {
    const result = validateIncidentTitle('<script>alert("xss")</script>');
    expect(result.isValid).toBe(false);
  });
});

describe('validateSeverity', () => {
  it('critical の場合は valid を返します', () => {
    expect(validateSeverity('critical').isValid).toBe(true);
  });

  it('high の場合は valid を返します', () => {
    expect(validateSeverity('high').isValid).toBe(true);
  });

  it('medium の場合は valid を返します', () => {
    expect(validateSeverity('medium').isValid).toBe(true);
  });

  it('low の場合は valid を返します', () => {
    expect(validateSeverity('low').isValid).toBe(true);
  });

  it('不明な重要度の場合は invalid を返します', () => {
    expect(validateSeverity('unknown').isValid).toBe(false);
  });
});

describe('validateStatus', () => {
  it('open の場合は valid を返します', () => {
    expect(validateStatus('open').isValid).toBe(true);
  });

  it('investigating の場合は valid を返します', () => {
    expect(validateStatus('investigating').isValid).toBe(true);
  });

  it('resolved の場合は valid を返します', () => {
    expect(validateStatus('resolved').isValid).toBe(true);
  });

  it('closed の場合は valid を返します', () => {
    expect(validateStatus('closed').isValid).toBe(true);
  });

  it('不明なステータスの場合は invalid を返します', () => {
    expect(validateStatus('unknown').isValid).toBe(false);
  });
});
