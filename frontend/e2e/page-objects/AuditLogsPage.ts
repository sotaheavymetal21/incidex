import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * Audit Logs Page Object (Admin only)
 */
export class AuditLogsPage extends BasePage {
  // Header elements
  readonly pageTitle = this.page.locator('h1:has-text("監査ログ"), h1:has-text("Audit")');

  // Filter controls
  readonly userFilter = this.page.locator('select[name="user"], #userFilter');
  readonly actionFilter = this.page.locator('select[name="action"], #actionFilter');
  readonly dateFromInput = this.page.locator('input[name="dateFrom"], #dateFrom');
  readonly dateToInput = this.page.locator('input[name="dateTo"], #dateTo');
  readonly searchInput = this.page.locator('input[placeholder*="検索"], input[name="search"]');
  readonly clearFiltersButton = this.page.getByRole('button', { name: /クリア|Clear/i });

  // Log table
  readonly logTable = this.page.locator('table');
  readonly logRows = this.page.locator('table tbody tr');

  // Pagination
  readonly previousButton = this.page.getByRole('button', { name: /Previous|前/i });
  readonly nextButton = this.page.getByRole('button', { name: /Next|次/i });
  readonly paginationInfo = this.page.locator('text=/page|ページ/i');

  constructor(page: Page) {
    super(page);
  }

  async goto(): Promise<void> {
    await this.page.goto('/audit-logs');
    await this.waitForPageLoad();
  }

  /**
   * Wait for logs to load
   */
  async waitForLogsLoaded(): Promise<void> {
    await this.page.waitForLoadState('networkidle');
    await this.logRows.first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {
      // May have no logs
    });
  }

  /**
   * Get all audit logs from the table
   */
  async getLogs(): Promise<
    Array<{
      timestamp: string;
      user: string;
      action: string;
      resource: string;
      details: string;
    }>
  > {
    await this.waitForLogsLoaded();
    const rows = await this.logRows.all();
    const logs: Array<{
      timestamp: string;
      user: string;
      action: string;
      resource: string;
      details: string;
    }> = [];

    for (const row of rows) {
      const cells = await row.locator('td').all();
      logs.push({
        timestamp: (await cells[0]?.textContent())?.trim() || '',
        user: (await cells[1]?.textContent())?.trim() || '',
        action: (await cells[2]?.textContent())?.trim() || '',
        resource: (await cells[3]?.textContent())?.trim() || '',
        details: (await cells[4]?.textContent())?.trim() || '',
      });
    }

    return logs;
  }

  /**
   * Filter by user
   */
  async filterByUser(userId: string): Promise<void> {
    await this.userFilter.selectOption(userId);
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Filter by action
   */
  async filterByAction(action: string): Promise<void> {
    await this.actionFilter.selectOption(action);
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Filter by date range
   */
  async filterByDateRange(from: string, to: string): Promise<void> {
    await this.dateFromInput.fill(from);
    await this.dateToInput.fill(to);
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Search logs
   */
  async search(query: string): Promise<void> {
    await this.searchInput.fill(query);
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Clear all filters
   */
  async clearFilters(): Promise<void> {
    await this.clearFiltersButton.click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Go to next page
   */
  async goToNextPage(): Promise<void> {
    await this.nextButton.click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Go to previous page
   */
  async goToPreviousPage(): Promise<void> {
    await this.previousButton.click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Get total log count (if displayed)
   */
  async getTotalCount(): Promise<number | null> {
    const infoText = await this.paginationInfo.textContent();
    const match = infoText?.match(/(\d+)\s*(total|件)/i);
    return match ? parseInt(match[1], 10) : null;
  }

  /**
   * Check if page is accessible (for RBAC testing)
   */
  async isPageAccessible(): Promise<boolean> {
    return await this.pageTitle.isVisible().catch(() => false);
  }

  /**
   * Check if empty state is displayed
   */
  async isEmptyStateDisplayed(): Promise<boolean> {
    return await this.page.locator('text=/ログがありません|No logs/i').isVisible().catch(() => false);
  }
}
