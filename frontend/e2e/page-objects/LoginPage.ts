import { Page, expect } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * Login Page Object
 */
export class LoginPage extends BasePage {
  readonly emailInput = this.page.getByLabel(/メール|Email/i);
  readonly passwordInput = this.page.getByLabel(/パスワード|Password/i);
  readonly loginButton = this.page.getByRole('button', { name: /ログイン|Login/i });
  readonly errorMessage = this.page.locator('[style*="error"], .error-message, [role="alert"]');
  readonly forgotPasswordLink = this.page.getByRole('link', { name: /パスワードを忘れた|forgot password/i });
  readonly signupLink = this.page.getByRole('link', { name: /アカウントをお持ちでない|sign up|register/i });

  constructor(page: Page) {
    super(page);
  }

  async goto(): Promise<void> {
    await this.page.goto('/login');
    await this.waitForPageLoad();
  }

  /**
   * Fill email field
   */
  async fillEmail(email: string): Promise<void> {
    await this.emailInput.fill(email);
  }

  /**
   * Fill password field
   */
  async fillPassword(password: string): Promise<void> {
    await this.passwordInput.fill(password);
  }

  /**
   * Submit login form
   */
  async submit(): Promise<void> {
    await this.loginButton.click();
  }

  /**
   * Perform complete login
   */
  async login(email: string, password: string): Promise<void> {
    await this.fillEmail(email);
    await this.fillPassword(password);
    await this.submit();
  }

  /**
   * Login and wait for redirect to dashboard
   */
  async loginAndWaitForDashboard(email: string, password: string): Promise<void> {
    await this.login(email, password);
    await expect(this.page).toHaveURL(/\/(dashboard|$)/, { timeout: 15000 });
  }

  /**
   * Check if login form is displayed
   */
  async isLoginFormVisible(): Promise<boolean> {
    return await this.emailInput.isVisible() &&
           await this.passwordInput.isVisible() &&
           await this.loginButton.isVisible();
  }

  /**
   * Get validation error message for email
   */
  async getEmailError(): Promise<string | null> {
    const errorElement = this.page.locator('input#email + p, [data-testid="email-error"]');
    if (await errorElement.isVisible().catch(() => false)) {
      return await errorElement.textContent();
    }
    return null;
  }

  /**
   * Get validation error message for password
   */
  async getPasswordError(): Promise<string | null> {
    const errorElement = this.page.locator('input#password + p, [data-testid="password-error"]');
    if (await errorElement.isVisible().catch(() => false)) {
      return await errorElement.textContent();
    }
    return null;
  }

  /**
   * Get general error message (e.g., invalid credentials)
   */
  async getGeneralError(): Promise<string | null> {
    if (await this.errorMessage.isVisible().catch(() => false)) {
      return await this.errorMessage.textContent();
    }
    return null;
  }

  /**
   * Navigate to forgot password page
   */
  async goToForgotPassword(): Promise<void> {
    await this.forgotPasswordLink.click();
    await expect(this.page).toHaveURL(/\/forgot-password/);
  }

  /**
   * Navigate to signup page
   */
  async goToSignup(): Promise<void> {
    await this.signupLink.click();
    await expect(this.page).toHaveURL(/\/signup/);
  }
}
