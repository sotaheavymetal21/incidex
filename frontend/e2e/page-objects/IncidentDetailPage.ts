import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * Incident Detail Page Object
 */
export class IncidentDetailPage extends BasePage {
  // Header actions
  readonly editButton = this.page.getByRole('button', { name: /編集|Edit/i }).or(
    this.page.getByRole('link', { name: /編集|Edit/i })
  );
  readonly deleteButton = this.page.getByRole('button', { name: /削除|Delete/i });
  readonly backButton = this.page.getByRole('link', { name: /戻る|Back/i });
  readonly pdfExportButton = this.page.getByRole('button', { name: /PDF|エクスポート/i });
  readonly postMortemButton = this.page.getByRole('button', { name: /Post-Mortem|ポストモーテム/i }).or(
    this.page.getByRole('link', { name: /Post-Mortem|ポストモーテム/i })
  );

  // Incident info
  readonly incidentTitle = this.page.locator('h1').first();
  // Severity badge shows text like "CRITICAL", "HIGH", "MEDIUM", "LOW"
  readonly severityBadge = this.page.locator('span.rounded-full').filter({ hasText: /^(CRITICAL|HIGH|MEDIUM|LOW)$/ }).first();
  // Status badge shows text like "Open", "Investigating", "Resolved", "Closed"
  readonly statusBadge = this.page.locator('span.rounded-full').filter({ hasText: /^(Open|Investigating|Resolved|Closed)$/ }).first();
  readonly description = this.page.locator('text=説明').locator('..').locator('p').first();

  // Comment section
  readonly commentInput = this.page.locator('textarea[placeholder*="コメント"], [data-testid="comment-input"]');
  readonly addCommentButton = this.page.getByRole('button', { name: /コメント追加|Add Comment|投稿/i });
  readonly commentsList = this.page.locator('[data-testid="comments-list"], .comments-list');

  // Timeline/Activity section
  readonly activitySection = this.page.locator('text=アクティビティ').locator('..').or(
    this.page.locator('text=タイムライン').locator('..')
  );
  readonly addActivityButton = this.page.getByRole('button', { name: /アクティビティ追加|Add Activity|イベント追加/i });

  // Attachments section
  readonly attachmentsSection = this.page.locator('text=添付ファイル').locator('..');
  readonly uploadFileButton = this.page.getByRole('button', { name: /アップロード|Upload/i });
  readonly fileInput = this.page.locator('input[type="file"]');

  // Assignee section
  readonly assigneeSection = this.page.locator('text=担当者').locator('..');
  readonly changeAssigneeButton = this.page.getByRole('button', { name: /担当者変更|Change Assignee|割り当て/i });

  constructor(page: Page) {
    super(page);
  }

  async goto(incidentId?: number): Promise<void> {
    if (incidentId) {
      await this.page.goto(`/incidents/${incidentId}`);
    }
    await this.waitForPageLoad();
  }

  /**
   * Get incident title text
   */
  async getTitle(): Promise<string> {
    return (await this.incidentTitle.textContent())?.trim() || '';
  }

  /**
   * Get severity value
   */
  async getSeverity(): Promise<string> {
    return (await this.severityBadge.textContent())?.trim().toLowerCase() || '';
  }

  /**
   * Get status value
   */
  async getStatus(): Promise<string> {
    return (await this.statusBadge.textContent())?.trim().toLowerCase() || '';
  }

  /**
   * Get description text
   */
  async getDescription(): Promise<string> {
    return (await this.description.textContent())?.trim() || '';
  }

  /**
   * Navigate to edit page
   */
  async clickEdit(): Promise<void> {
    await this.editButton.click();
    await expect(this.page).toHaveURL(/\/incidents\/\d+\/edit/);
  }

  /**
   * Delete incident (with confirmation)
   */
  async deleteIncident(): Promise<void> {
    await this.deleteButton.click();
    await this.confirmDialog();
    await expect(this.page).toHaveURL(/\/incidents$/);
  }

  /**
   * Add a comment
   */
  async addComment(comment: string): Promise<void> {
    await this.commentInput.fill(comment);
    await this.addCommentButton.click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Get all comments
   */
  async getComments(): Promise<Array<{ author: string; content: string; timestamp: string }>> {
    const commentItems = this.commentsList.locator('[data-testid="comment-item"], .comment-item, > div');
    const count = await commentItems.count();
    const comments: Array<{ author: string; content: string; timestamp: string }> = [];

    for (let i = 0; i < count; i++) {
      const item = commentItems.nth(i);
      comments.push({
        author: (await item.locator('.author, [data-testid="author"]').textContent())?.trim() || '',
        content: (await item.locator('.content, [data-testid="content"], p').textContent())?.trim() || '',
        timestamp: (await item.locator('.timestamp, [data-testid="timestamp"], time').textContent())?.trim() || '',
      });
    }

    return comments;
  }

  /**
   * Upload an attachment
   */
  async uploadAttachment(filePath: string): Promise<void> {
    await this.fileInput.setInputFiles(filePath);
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Get uploaded attachments
   */
  async getAttachments(): Promise<string[]> {
    const attachmentLinks = this.attachmentsSection.locator('a, [data-testid="attachment-item"]');
    const count = await attachmentLinks.count();
    const attachments: string[] = [];

    for (let i = 0; i < count; i++) {
      const name = await attachmentLinks.nth(i).textContent();
      if (name) attachments.push(name.trim());
    }

    return attachments;
  }

  /**
   * Download attachment by name
   */
  async downloadAttachment(name: string): Promise<void> {
    const downloadPromise = this.page.waitForEvent('download');
    await this.attachmentsSection.getByRole('link', { name }).click();
    await downloadPromise;
  }

  /**
   * Navigate to post-mortem
   */
  async clickPostMortem(): Promise<void> {
    await this.postMortemButton.click();
    await expect(this.page).toHaveURL(/\/incidents\/\d+\/postmortem/);
  }

  /**
   * Export to PDF
   */
  async exportPDF(): Promise<void> {
    const downloadPromise = this.page.waitForEvent('download');
    await this.pdfExportButton.click();
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toMatch(/\.pdf$/);
  }

  /**
   * Check if edit button is visible
   */
  async isEditButtonVisible(): Promise<boolean> {
    return await this.editButton.isVisible().catch(() => false);
  }

  /**
   * Check if delete button is visible
   */
  async isDeleteButtonVisible(): Promise<boolean> {
    return await this.deleteButton.isVisible().catch(() => false);
  }

  /**
   * Get activity/timeline events
   */
  async getActivityEvents(): Promise<Array<{ type: string; description: string; timestamp: string }>> {
    const activityItems = this.activitySection.locator('[data-testid="activity-item"], .activity-item, > div');
    const count = await activityItems.count();
    const events: Array<{ type: string; description: string; timestamp: string }> = [];

    for (let i = 0; i < count; i++) {
      const item = activityItems.nth(i);
      events.push({
        type: (await item.locator('.type, [data-testid="type"]').textContent())?.trim() || '',
        description: (await item.locator('.description, [data-testid="description"], p').textContent())?.trim() || '',
        timestamp: (await item.locator('.timestamp, [data-testid="timestamp"], time').textContent())?.trim() || '',
      });
    }

    return events;
  }
}
