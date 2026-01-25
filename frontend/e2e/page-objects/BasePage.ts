import { Page, Locator, expect } from '@playwright/test';

/**
 * Base Page Object with common functionality
 */
export abstract class BasePage {
  readonly page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  /**
   * Navigate to the page URL
   */
  abstract goto(): Promise<void>;

  /**
   * Wait for page to be fully loaded
   */
  async waitForPageLoad(): Promise<void> {
    // Wait for network to be idle
    await this.page.waitForLoadState('networkidle');
    // Wait for loading indicators to disappear
    const loadingIndicators = this.page.locator(
      '[data-testid="loading"], .loading, text="Loading...", text="読み込み中..."'
    );
    await expect(loadingIndicators).toHaveCount(0, { timeout: 10000 }).catch(() => {
      // Some pages may not have loading indicators
    });
  }

  /**
   * Get current logged-in user info from localStorage
   */
  async getCurrentUser(): Promise<{
    id: number;
    email: string;
    name: string;
    role: string;
  } | null> {
    return await this.page.evaluate(() => {
      const userStr = localStorage.getItem('user');
      return userStr ? JSON.parse(userStr) : null;
    });
  }

  /**
   * Check if user is authenticated
   */
  async isAuthenticated(): Promise<boolean> {
    const token = await this.page.evaluate(() => localStorage.getItem('token'));
    return !!token;
  }

  /**
   * Logout from the application
   */
  async logout(): Promise<void> {
    // Look for logout button in nav or dropdown
    const logoutButton = this.page.getByRole('button', { name: /ログアウト|Logout/i });
    const userMenu = this.page.locator('[data-testid="user-menu"], [aria-label="User menu"]');

    // Try to open user menu if exists
    if (await userMenu.isVisible().catch(() => false)) {
      await userMenu.click();
      await logoutButton.click();
    } else if (await logoutButton.isVisible().catch(() => false)) {
      await logoutButton.click();
    } else {
      // Fallback: clear storage directly
      await this.page.evaluate(() => {
        localStorage.clear();
      });
      await this.page.goto('/login');
    }

    // Wait for redirect to login
    await expect(this.page).toHaveURL(/\/login/);
  }

  /**
   * Get toast/notification message
   */
  async getToastMessage(): Promise<string | null> {
    const toast = this.page.locator(
      '[role="alert"], [data-testid="toast"], .toast, .notification'
    ).first();

    if (await toast.isVisible({ timeout: 5000 }).catch(() => false)) {
      return await toast.textContent();
    }
    return null;
  }

  /**
   * Wait for and get success toast
   */
  async expectSuccessToast(message?: string | RegExp): Promise<void> {
    const toast = this.page.locator(
      '[role="alert"].success, [data-testid="toast-success"], .toast-success'
    ).first();

    if (message) {
      await expect(toast).toContainText(message);
    } else {
      await expect(toast).toBeVisible();
    }
  }

  /**
   * Wait for and get error toast
   */
  async expectErrorToast(message?: string | RegExp): Promise<void> {
    const toast = this.page.locator(
      '[role="alert"].error, [data-testid="toast-error"], .toast-error, [style*="error"]'
    ).first();

    if (message) {
      await expect(toast).toContainText(message);
    } else {
      await expect(toast).toBeVisible();
    }
  }

  /**
   * Click confirmation dialog OK/Yes button
   */
  async confirmDialog(): Promise<void> {
    const confirmButton = this.page.getByRole('button', {
      name: /確認|OK|はい|Yes|削除|Delete|続行|Continue/i
    });
    await confirmButton.click();
  }

  /**
   * Click confirmation dialog Cancel/No button
   */
  async cancelDialog(): Promise<void> {
    const cancelButton = this.page.getByRole('button', {
      name: /キャンセル|Cancel|いいえ|No|閉じる|Close/i
    });
    await cancelButton.click();
  }

  /**
   * Check if element exists
   */
  async elementExists(locator: Locator): Promise<boolean> {
    return await locator.count() > 0;
  }

  /**
   * Wait for navigation to complete
   */
  async waitForNavigation(urlPattern: string | RegExp): Promise<void> {
    await this.page.waitForURL(urlPattern);
    await this.waitForPageLoad();
  }
}
