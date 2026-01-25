import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * Dashboard Page Object
 */
export class DashboardPage extends BasePage {
  // Statistics cards - locate by finding the p element and going to its parent card container
  // The cards have onClick handlers and contain the title in a p.text-sm element
  readonly totalIncidentsCard = this.page.locator('div.cursor-pointer').filter({ has: this.page.locator('p:has-text("総インシデント数")') }).first();
  readonly criticalCard = this.page.locator('div.cursor-pointer').filter({ has: this.page.getByRole('paragraph').filter({ hasText: /^Critical$/ }) }).first();
  readonly openCard = this.page.locator('div.cursor-pointer').filter({ has: this.page.locator('p:has-text("Open（未対応）")') }).first();
  readonly resolvedCard = this.page.locator('div.cursor-pointer').filter({ has: this.page.locator('p:has-text("Resolved（解決済み）")') }).first();

  // Graph type buttons
  readonly pieChartButton = this.page.getByRole('button', { name: '分布グラフ' });
  readonly timeseriesButton = this.page.getByRole('button', { name: '時系列グラフ' });
  readonly barChartButton = this.page.getByRole('button', { name: '棒グラフ' });

  // Period buttons
  readonly dailyButton = this.page.getByRole('button', { name: /日次|日別/i });
  readonly weeklyButton = this.page.getByRole('button', { name: /週次|週別/i });
  readonly monthlyButton = this.page.getByRole('button', { name: /月次|月別/i });

  // Chart sections
  readonly severityChart = this.page.locator('text=重要度別分布').locator('..');
  readonly statusChart = this.page.locator('text=ステータス別分布').locator('..');
  readonly trendChart = this.page.locator('text=インシデント発生トレンド').locator('..');

  // Recent incidents table
  readonly recentIncidentsTable = this.page.locator('text=最近のインシデント').locator('..').locator('..');

  constructor(page: Page) {
    super(page);
  }

  async goto(): Promise<void> {
    await this.page.goto('/dashboard');
    await this.waitForPageLoad();
  }

  /**
   * Get total incidents count
   */
  async getTotalIncidentsCount(): Promise<number> {
    const countText = await this.totalIncidentsCard.locator('p.text-4xl').textContent();
    return parseInt(countText || '0', 10);
  }

  /**
   * Get critical incidents count
   */
  async getCriticalCount(): Promise<number> {
    const countText = await this.criticalCard.locator('p.text-4xl').textContent();
    return parseInt(countText || '0', 10);
  }

  /**
   * Get open incidents count
   */
  async getOpenCount(): Promise<number> {
    const countText = await this.openCard.locator('p.text-4xl').textContent();
    return parseInt(countText || '0', 10);
  }

  /**
   * Get resolved incidents count
   */
  async getResolvedCount(): Promise<number> {
    const countText = await this.resolvedCard.locator('p.text-4xl').textContent();
    return parseInt(countText || '0', 10);
  }

  /**
   * Click total incidents card to navigate to incidents list
   */
  async clickTotalIncidentsCard(): Promise<void> {
    await this.totalIncidentsCard.click();
    await expect(this.page).toHaveURL(/\/incidents/);
  }

  /**
   * Click critical card to navigate to critical incidents
   */
  async clickCriticalCard(): Promise<void> {
    await this.criticalCard.click();
    await expect(this.page).toHaveURL(/\/incidents\?severity=critical/);
  }

  /**
   * Click open card to navigate to open incidents
   */
  async clickOpenCard(): Promise<void> {
    await this.openCard.click();
    await expect(this.page).toHaveURL(/\/incidents\?status=open/);
  }

  /**
   * Click resolved card to navigate to resolved incidents
   */
  async clickResolvedCard(): Promise<void> {
    await this.resolvedCard.click();
    await expect(this.page).toHaveURL(/\/incidents\?status=resolved/);
  }

  /**
   * Switch to pie chart view
   */
  async switchToPieChart(): Promise<void> {
    await this.pieChartButton.click();
    await expect(this.severityChart).toBeVisible();
  }

  /**
   * Switch to time series view
   */
  async switchToTimeSeries(): Promise<void> {
    await this.timeseriesButton.click();
    await expect(this.trendChart).toBeVisible();
  }

  /**
   * Switch to bar chart view
   */
  async switchToBarChart(): Promise<void> {
    await this.barChartButton.click();
    // Wait for bar chart to render
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Set period to daily
   */
  async setPeriodDaily(): Promise<void> {
    await this.dailyButton.first().click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Set period to weekly
   */
  async setPeriodWeekly(): Promise<void> {
    await this.weeklyButton.first().click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Set period to monthly
   */
  async setPeriodMonthly(): Promise<void> {
    await this.monthlyButton.first().click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Get recent incidents from table
   */
  async getRecentIncidents(): Promise<Array<{ title: string; severity: string; status: string }>> {
    const rows = this.recentIncidentsTable.locator('tbody tr');
    const count = await rows.count();
    const incidents: Array<{ title: string; severity: string; status: string }> = [];

    for (let i = 0; i < count; i++) {
      const row = rows.nth(i);
      const title = await row.locator('td').first().textContent();
      const severity = await row.locator('td').nth(1).textContent();
      const status = await row.locator('td').nth(2).textContent();

      incidents.push({
        title: title?.trim() || '',
        severity: severity?.trim() || '',
        status: status?.trim() || '',
      });
    }

    return incidents;
  }

  /**
   * Click a recent incident to navigate to detail
   */
  async clickRecentIncident(index: number): Promise<void> {
    const rows = this.recentIncidentsTable.locator('tbody tr');
    await rows.nth(index).click();
    await expect(this.page).toHaveURL(/\/incidents\/\d+/);
  }

  /**
   * Check if all statistics cards are displayed
   */
  async areStatisticsCardsVisible(): Promise<boolean> {
    // Check for key text elements that indicate cards are loaded
    // Use getByRole('paragraph') to specifically target p elements with the card titles
    const totalVisible = await this.page.getByRole('paragraph').filter({ hasText: '総インシデント数' }).isVisible();
    const criticalVisible = await this.page.getByRole('paragraph').filter({ hasText: /^Critical$/ }).isVisible();
    const openVisible = await this.page.getByRole('paragraph').filter({ hasText: 'Open（未対応）' }).isVisible();
    const resolvedVisible = await this.page.getByRole('paragraph').filter({ hasText: 'Resolved（解決済み）' }).isVisible();

    return totalVisible && criticalVisible && openVisible && resolvedVisible;
  }
}
