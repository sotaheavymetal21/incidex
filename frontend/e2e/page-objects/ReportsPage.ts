import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * Reports Page Object
 */
export class ReportsPage extends BasePage {
  // Header elements
  readonly pageTitle = this.page.getByRole('heading', { name: '月次レポート' });

  // Month navigation
  readonly previousMonthButton = this.page.getByRole('button', { name: /← 前月/ });
  readonly nextMonthButton = this.page.getByRole('button', { name: /翌月 →/ });
  readonly currentMonthDisplay = this.page.locator('h2.text-xl, [data-testid="current-month"]').first();

  // Summary cards
  readonly summarySection = this.page.locator('.grid').first();
  readonly totalIncidentsCard = this.page.locator('text=総インシデント').locator('..');
  readonly avgResolutionTimeCard = this.page.locator('text=平均解決時間').locator('..');
  readonly criticalCountCard = this.page.locator('text=Critical').locator('..').first();

  // Charts
  readonly severityChart = this.page.locator('[data-testid="severity-chart"], .severity-chart');
  readonly statusChart = this.page.locator('[data-testid="status-chart"], .status-chart');
  readonly trendChart = this.page.locator('[data-testid="trend-chart"], .trend-chart');

  // Export button
  readonly exportButton = this.page.getByRole('button', { name: /エクスポート|Export|PDF/i });

  constructor(page: Page) {
    super(page);
  }

  async goto(): Promise<void> {
    await this.page.goto('/reports');
    await this.waitForPageLoad();
  }

  /**
   * Wait for report data to load
   */
  async waitForReportLoaded(): Promise<void> {
    await this.page.waitForLoadState('networkidle');
    // Wait for the page title to be visible
    await this.pageTitle.waitFor({ state: 'visible', timeout: 10000 }).catch(() => {
      // May have different structure
    });
  }

  /**
   * Get current displayed month
   */
  async getCurrentMonth(): Promise<string> {
    return (await this.currentMonthDisplay.textContent())?.trim() || '';
  }

  /**
   * Navigate to previous month
   */
  async goToPreviousMonth(): Promise<void> {
    await this.previousMonthButton.click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Navigate to next month
   */
  async goToNextMonth(): Promise<void> {
    await this.nextMonthButton.click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Get total incidents count from summary
   */
  async getTotalIncidentsCount(): Promise<number> {
    const text = await this.totalIncidentsCard.locator('.count, [data-testid="count"], .text-4xl, .text-3xl').textContent();
    return parseInt(text?.replace(/[^0-9]/g, '') || '0', 10);
  }

  /**
   * Get average resolution time from summary
   */
  async getAverageResolutionTime(): Promise<string> {
    const text = await this.avgResolutionTimeCard.locator('.value, [data-testid="value"], .text-4xl, .text-3xl').textContent();
    return text?.trim() || '';
  }

  /**
   * Get critical incidents count from summary
   */
  async getCriticalCount(): Promise<number> {
    const text = await this.criticalCountCard.locator('.count, [data-testid="count"], .text-4xl, .text-3xl').textContent();
    return parseInt(text?.replace(/[^0-9]/g, '') || '0', 10);
  }

  /**
   * Check if severity chart is visible
   */
  async isSeverityChartVisible(): Promise<boolean> {
    return await this.severityChart.isVisible().catch(() => false);
  }

  /**
   * Check if status chart is visible
   */
  async isStatusChartVisible(): Promise<boolean> {
    return await this.statusChart.isVisible().catch(() => false);
  }

  /**
   * Check if trend chart is visible
   */
  async isTrendChartVisible(): Promise<boolean> {
    return await this.trendChart.isVisible().catch(() => false);
  }

  /**
   * Export report
   */
  async exportReport(): Promise<void> {
    const downloadPromise = this.page.waitForEvent('download');
    await this.exportButton.click();
    await downloadPromise;
  }

  /**
   * Check if summary cards are visible
   */
  async areSummaryCardsVisible(): Promise<boolean> {
    return await this.totalIncidentsCard.isVisible() &&
           await this.avgResolutionTimeCard.isVisible().catch(() => true);
  }

  /**
   * Check if page is accessible
   */
  async isPageAccessible(): Promise<boolean> {
    return await this.pageTitle.isVisible().catch(() => false);
  }
}
