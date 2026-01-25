import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './BasePage';

export interface PostMortemFormData {
  summary?: string;
  rootCause?: string;
  impactAnalysis?: string;
  timeline?: string;
  lessonsLearned?: string;
  preventiveMeasures?: string;
}

/**
 * Post-Mortem Page Object
 */
export class PostMortemPage extends BasePage {
  // Header elements
  readonly pageTitle = this.page.locator('h1:has-text("Post-Mortem"), h1:has-text("ポストモーテム")');
  readonly incidentTitle = this.page.locator('[data-testid="incident-title"], .incident-title');

  // Status badge
  readonly statusBadge = this.page.locator('[data-testid="postmortem-status"], .postmortem-status');

  // Form sections
  readonly summarySection = this.page.locator('text=/概要|Summary/i').locator('..');
  readonly rootCauseSection = this.page.locator('text=/根本原因|Root Cause/i').locator('..');
  readonly impactSection = this.page.locator('text=/影響分析|Impact/i').locator('..');
  readonly timelineSection = this.page.locator('text=/タイムライン|Timeline/i').locator('..');
  readonly lessonsSection = this.page.locator('text=/教訓|Lessons/i').locator('..');
  readonly preventiveSection = this.page.locator('text=/予防策|Preventive/i').locator('..');

  // Form inputs
  readonly summaryInput = this.page.locator('textarea[name="summary"], #summary');
  readonly rootCauseInput = this.page.locator('textarea[name="rootCause"], #rootCause, textarea[name="root_cause"]');
  readonly impactInput = this.page.locator('textarea[name="impact"], #impact, textarea[name="impact_analysis"]');
  readonly timelineInput = this.page.locator('textarea[name="timeline"], #timeline');
  readonly lessonsInput = this.page.locator('textarea[name="lessons"], #lessons, textarea[name="lessons_learned"]');
  readonly preventiveInput = this.page.locator('textarea[name="preventive"], #preventive, textarea[name="preventive_measures"]');

  // Action items section
  readonly actionItemsSection = this.page.locator('text=/アクションアイテム|Action Items/i').locator('..');
  readonly addActionItemButton = this.page.getByRole('button', { name: /アクションアイテム追加|Add Action|追加/i });
  readonly actionItemsList = this.page.locator('[data-testid="action-items"], .action-items-list');

  // Action buttons
  readonly saveDraftButton = this.page.getByRole('button', { name: /下書き保存|Save Draft|保存/i });
  readonly publishButton = this.page.getByRole('button', { name: /公開|Publish/i });
  readonly backButton = this.page.getByRole('link', { name: /戻る|Back/i });

  constructor(page: Page) {
    super(page);
  }

  async goto(incidentId?: number): Promise<void> {
    if (incidentId) {
      await this.page.goto(`/incidents/${incidentId}/postmortem`);
    }
    await this.waitForPageLoad();
  }

  /**
   * Fill summary section
   */
  async fillSummary(summary: string): Promise<void> {
    await this.summaryInput.fill(summary);
  }

  /**
   * Fill root cause section
   */
  async fillRootCause(rootCause: string): Promise<void> {
    await this.rootCauseInput.fill(rootCause);
  }

  /**
   * Fill impact analysis section
   */
  async fillImpactAnalysis(impact: string): Promise<void> {
    await this.impactInput.fill(impact);
  }

  /**
   * Fill timeline section
   */
  async fillTimeline(timeline: string): Promise<void> {
    await this.timelineInput.fill(timeline);
  }

  /**
   * Fill lessons learned section
   */
  async fillLessonsLearned(lessons: string): Promise<void> {
    await this.lessonsInput.fill(lessons);
  }

  /**
   * Fill preventive measures section
   */
  async fillPreventiveMeasures(measures: string): Promise<void> {
    await this.preventiveInput.fill(measures);
  }

  /**
   * Fill entire form
   */
  async fillForm(data: PostMortemFormData): Promise<void> {
    if (data.summary) await this.fillSummary(data.summary);
    if (data.rootCause) await this.fillRootCause(data.rootCause);
    if (data.impactAnalysis) await this.fillImpactAnalysis(data.impactAnalysis);
    if (data.timeline) await this.fillTimeline(data.timeline);
    if (data.lessonsLearned) await this.fillLessonsLearned(data.lessonsLearned);
    if (data.preventiveMeasures) await this.fillPreventiveMeasures(data.preventiveMeasures);
  }

  /**
   * Save as draft
   */
  async saveDraft(): Promise<void> {
    await this.saveDraftButton.click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Publish post-mortem
   */
  async publish(): Promise<void> {
    await this.publishButton.click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Add an action item
   */
  async addActionItem(description: string, assignee?: string, dueDate?: string): Promise<void> {
    await this.addActionItemButton.click();

    // Find the new action item row (last one)
    const actionItems = this.actionItemsList.locator('[data-testid="action-item"], .action-item');
    const lastItem = actionItems.last();

    // Fill description
    await lastItem.locator('input[name*="description"], textarea').fill(description);

    // Fill assignee if provided
    if (assignee) {
      const assigneeSelect = lastItem.locator('select[name*="assignee"]');
      if (await assigneeSelect.isVisible()) {
        await assigneeSelect.selectOption({ label: assignee });
      }
    }

    // Fill due date if provided
    if (dueDate) {
      const dueDateInput = lastItem.locator('input[type="date"], input[name*="due"]');
      if (await dueDateInput.isVisible()) {
        await dueDateInput.fill(dueDate);
      }
    }
  }

  /**
   * Get all action items
   */
  async getActionItems(): Promise<
    Array<{
      description: string;
      assignee: string;
      status: string;
      dueDate: string;
    }>
  > {
    const items = this.actionItemsList.locator('[data-testid="action-item"], .action-item');
    const count = await items.count();
    const actionItems: Array<{
      description: string;
      assignee: string;
      status: string;
      dueDate: string;
    }> = [];

    for (let i = 0; i < count; i++) {
      const item = items.nth(i);
      actionItems.push({
        description: (await item.locator('.description, [data-testid="description"]').textContent())?.trim() || '',
        assignee: (await item.locator('.assignee, [data-testid="assignee"]').textContent())?.trim() || '',
        status: (await item.locator('.status, [data-testid="status"]').textContent())?.trim() || '',
        dueDate: (await item.locator('.due-date, [data-testid="due-date"]').textContent())?.trim() || '',
      });
    }

    return actionItems;
  }

  /**
   * Delete an action item by index
   */
  async deleteActionItem(index: number): Promise<void> {
    const items = this.actionItemsList.locator('[data-testid="action-item"], .action-item');
    await items.nth(index).getByRole('button', { name: /削除|Delete|×/i }).click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Toggle action item completion status
   */
  async toggleActionItemStatus(index: number): Promise<void> {
    const items = this.actionItemsList.locator('[data-testid="action-item"], .action-item');
    const checkbox = items.nth(index).locator('input[type="checkbox"], [role="checkbox"]');
    await checkbox.click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Get current status (draft/published)
   */
  async getStatus(): Promise<string> {
    if (await this.statusBadge.isVisible().catch(() => false)) {
      return (await this.statusBadge.textContent())?.trim().toLowerCase() || '';
    }
    return '';
  }

  /**
   * Check if form is editable
   */
  async isEditable(): Promise<boolean> {
    return await this.summaryInput.isEnabled().catch(() => false);
  }

  /**
   * Check if publish button is visible
   */
  async isPublishButtonVisible(): Promise<boolean> {
    return await this.publishButton.isVisible().catch(() => false);
  }

  /**
   * Go back to incident detail
   */
  async goBack(): Promise<void> {
    await this.backButton.click();
    await expect(this.page).toHaveURL(/\/incidents\/\d+$/);
  }
}
