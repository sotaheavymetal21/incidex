import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './BasePage';

export type Severity = 'critical' | 'high' | 'medium' | 'low';
export type Status = 'open' | 'investigating' | 'resolved' | 'closed';
export type SortField = 'title' | 'severity' | 'status' | 'detected_at';

/**
 * Incident List Page Object
 */
export class IncidentListPage extends BasePage {
  // Header elements
  readonly pageTitle = this.page.getByRole('heading', { name: /インシデント一覧/i });
  readonly createButton = this.page.getByRole('link', { name: /新規作成/i });
  readonly exportButton = this.page.getByRole('button', { name: /CSV/i });

  // Filter sidebar - look for the heading within the sidebar
  readonly filterSidebar = this.page.locator('h2:has-text("フィルター")').locator('..');
  readonly searchInput = this.page.locator('input[placeholder*="タイトル"]');
  readonly clearAllFiltersButton = this.page.getByRole('button', { name: 'すべてクリア' });
  readonly unresolvedPresetButton = this.page.getByRole('button', { name: '未解決のインシデント' });
  readonly criticalPresetButton = this.page.getByRole('button', { name: 'Critical のみ' });

  // Table
  readonly incidentTable = this.page.locator('table');
  readonly tableRows = this.page.locator('table tbody tr');

  // Pagination
  readonly previousButton = this.page.getByRole('button', { name: /Previous/i });
  readonly nextButton = this.page.getByRole('button', { name: /Next/i });
  readonly paginationInfo = this.page.locator('text=/Showing page/');

  // Filter chips
  readonly filterChips = this.page.locator('[class*="rounded-full"]');

  constructor(page: Page) {
    super(page);
  }

  async goto(): Promise<void> {
    await this.page.goto('/incidents');
    await this.waitForPageLoad();
  }

  /**
   * Wait for incidents to load
   */
  async waitForIncidentsLoaded(): Promise<void> {
    await this.page.waitForLoadState('networkidle');
    // Wait for either table rows or empty message
    await Promise.race([
      this.tableRows.first().waitFor({ state: 'visible', timeout: 10000 }),
      this.page.locator('text=インシデントが見つかりませんでした').waitFor({ state: 'visible', timeout: 10000 }),
    ]).catch(() => {
      // Page may already be loaded
    });
  }

  /**
   * Get incident count from pagination
   */
  async getTotalIncidentCount(): Promise<number> {
    const infoText = await this.paginationInfo.textContent();
    const match = infoText?.match(/\((\d+) total\)/);
    return match ? parseInt(match[1], 10) : 0;
  }

  /**
   * Get all visible incident rows
   */
  async getIncidentRows(): Promise<
    Array<{
      title: string;
      severity: string;
      status: string;
      detectedAt: string;
      assignee: string;
    }>
  > {
    await this.waitForIncidentsLoaded();
    const rows = await this.tableRows.all();
    const incidents: Array<{
      title: string;
      severity: string;
      status: string;
      detectedAt: string;
      assignee: string;
    }> = [];

    for (const row of rows) {
      const cells = await row.locator('td').all();
      incidents.push({
        title: (await cells[0]?.textContent())?.trim() || '',
        severity: (await cells[1]?.textContent())?.trim() || '',
        status: (await cells[2]?.textContent())?.trim() || '',
        detectedAt: (await cells[3]?.textContent())?.trim() || '',
        assignee: (await cells[4]?.textContent())?.trim() || '',
      });
    }

    return incidents;
  }

  /**
   * Search for incidents
   */
  async search(query: string): Promise<void> {
    await this.searchInput.fill(query);
    await this.page.waitForLoadState('networkidle');
    await this.waitForIncidentsLoaded();
  }

  /**
   * Clear search
   */
  async clearSearch(): Promise<void> {
    await this.searchInput.clear();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Filter by severity
   */
  async filterBySeverity(severity: Severity | ''): Promise<void> {
    const label = severity || 'すべて';
    await this.page.getByLabel(severity ? severity.charAt(0).toUpperCase() + severity.slice(1) : 'すべて', { exact: false }).first().check();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Filter by status
   */
  async filterByStatus(status: Status | ''): Promise<void> {
    // Find and click the radio button for the status
    const statusLabels: Record<string, string> = {
      open: 'Open',
      investigating: 'Investigating',
      resolved: 'Resolved',
      closed: 'Closed',
      '': 'すべて',
    };
    const radioLabel = statusLabels[status] || 'すべて';
    await this.page.locator(`input[name="status"]`).locator(`..`).filter({ hasText: radioLabel }).locator('input').check();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Toggle tag filter
   */
  async toggleTagFilter(tagName: string): Promise<void> {
    await this.page.getByLabel(tagName, { exact: true }).check();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Apply unresolved preset
   */
  async applyUnresolvedPreset(): Promise<void> {
    await this.unresolvedPresetButton.click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Apply critical preset
   */
  async applyCriticalPreset(): Promise<void> {
    await this.criticalPresetButton.click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Clear all filters
   */
  async clearAllFilters(): Promise<void> {
    await this.clearAllFiltersButton.click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Sort by column
   */
  async sortBy(field: SortField): Promise<void> {
    const columnHeaders: Record<SortField, string> = {
      title: 'Title',
      severity: 'Severity',
      status: 'Status',
      detected_at: 'Detected At',
    };
    await this.page.getByRole('button', { name: columnHeaders[field] }).click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Click an incident row to navigate to detail
   */
  async clickIncident(index: number): Promise<void> {
    await this.tableRows.nth(index).click();
    await expect(this.page).toHaveURL(/\/incidents\/\d+/);
  }

  /**
   * Click incident by title
   */
  async clickIncidentByTitle(title: string): Promise<void> {
    await this.tableRows.filter({ hasText: title }).first().click();
    await expect(this.page).toHaveURL(/\/incidents\/\d+/);
  }

  /**
   * Click create button
   */
  async clickCreateButton(): Promise<void> {
    await this.createButton.click();
    await expect(this.page).toHaveURL(/\/incidents\/create/);
  }

  /**
   * Export to CSV
   */
  async exportCSV(): Promise<void> {
    // Set up download listener
    const downloadPromise = this.page.waitForEvent('download');
    await this.exportButton.click();
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toMatch(/incidents.*\.csv/);
  }

  /**
   * Go to next page
   */
  async goToNextPage(): Promise<void> {
    await this.nextButton.last().click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Go to previous page
   */
  async goToPreviousPage(): Promise<void> {
    await this.previousButton.last().click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Check if create button is visible
   */
  async isCreateButtonVisible(): Promise<boolean> {
    return await this.createButton.isVisible().catch(() => false);
  }

  /**
   * Check if empty state is displayed
   */
  async isEmptyStateDisplayed(): Promise<boolean> {
    return await this.page.locator('text=インシデントが見つかりませんでした').isVisible().catch(() => false);
  }

  /**
   * Get active filter chips
   */
  async getActiveFilterChips(): Promise<string[]> {
    const chipContainer = this.page.locator('text=適用中のフィルター').locator('..');
    if (!await chipContainer.isVisible().catch(() => false)) {
      return [];
    }

    const chips = chipContainer.locator('[class*="rounded-full"]');
    const count = await chips.count();
    const chipTexts: string[] = [];

    for (let i = 0; i < count; i++) {
      const text = await chips.nth(i).textContent();
      if (text) chipTexts.push(text.trim());
    }

    return chipTexts;
  }
}
