import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './BasePage';

export interface TagFormData {
  name: string;
  color: string;
}

/**
 * Tags Management Page Object
 */
export class TagsPage extends BasePage {
  // Header elements
  readonly pageTitle = this.page.getByRole('heading', { name: 'タグ管理' });
  readonly createTagButton = this.page.getByRole('button', { name: 'タグを作成' });

  // Tag table or list
  readonly tagTable = this.page.locator('table');
  readonly tagRows = this.page.locator('table tbody tr');
  readonly tagCards = this.page.locator('[data-testid="tag-card"], .tag-card');

  // Create/Edit Modal or Form
  readonly tagModal = this.page.locator('[role="dialog"], .modal, .fixed.inset-0');
  readonly nameInput = this.page.locator('input#name, input[name="name"]');
  readonly colorInput = this.page.locator('input#color, input[name="color"], input[type="color"]');
  readonly saveButton = this.page.getByRole('button', { name: '保存' });
  readonly cancelButton = this.page.getByRole('button', { name: 'キャンセル' });

  constructor(page: Page) {
    super(page);
  }

  async goto(): Promise<void> {
    await this.page.goto('/tags');
    await this.waitForPageLoad();
  }

  /**
   * Wait for tags to load
   */
  async waitForTagsLoaded(): Promise<void> {
    await this.page.waitForLoadState('networkidle');
    await Promise.race([
      this.tagRows.first().waitFor({ state: 'visible', timeout: 10000 }),
      this.tagCards.first().waitFor({ state: 'visible', timeout: 10000 }),
    ]).catch(() => {
      // May have no tags
    });
  }

  /**
   * Get all tags
   */
  async getTags(): Promise<Array<{ name: string; color: string }>> {
    await this.waitForTagsLoaded();

    // Try table format first
    if (await this.tagTable.isVisible().catch(() => false)) {
      const rows = await this.tagRows.all();
      const tags: Array<{ name: string; color: string }> = [];

      for (const row of rows) {
        const cells = await row.locator('td').all();
        const colorCell = await row.locator('[style*="background"], .color-preview').first();
        tags.push({
          name: (await cells[0]?.textContent())?.trim() || '',
          color: (await colorCell.getAttribute('style'))?.match(/#[0-9a-fA-F]{6}|rgb\([^)]+\)/)?.[0] || '',
        });
      }
      return tags;
    }

    // Try card format
    const cards = await this.tagCards.all();
    const tags: Array<{ name: string; color: string }> = [];

    for (const card of cards) {
      const name = await card.locator('.tag-name, [data-testid="tag-name"]').textContent();
      const colorElement = await card.locator('[style*="background"], .color-preview').first();
      const style = await colorElement.getAttribute('style');
      tags.push({
        name: name?.trim() || '',
        color: style?.match(/#[0-9a-fA-F]{6}|rgb\([^)]+\)/)?.[0] || '',
      });
    }

    return tags;
  }

  /**
   * Open create tag modal/form
   */
  async openCreateForm(): Promise<void> {
    await this.createTagButton.click();
    // Wait for either modal or inline form
    await Promise.race([
      expect(this.tagModal).toBeVisible(),
      expect(this.nameInput).toBeVisible(),
    ]);
  }

  /**
   * Fill tag form
   */
  async fillTagForm(data: TagFormData): Promise<void> {
    await this.nameInput.fill(data.name);
    // Handle color input - may be text or color picker
    const colorInputType = await this.colorInput.getAttribute('type');
    if (colorInputType === 'color') {
      await this.colorInput.fill(data.color);
    } else {
      await this.colorInput.fill(data.color);
    }
  }

  /**
   * Create a new tag
   */
  async createTag(data: TagFormData): Promise<void> {
    await this.openCreateForm();
    await this.fillTagForm(data);
    await this.saveButton.click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Open edit form for a tag
   */
  async openEditForm(tagName: string): Promise<void> {
    // Try table row
    if (await this.tagTable.isVisible().catch(() => false)) {
      const row = this.tagRows.filter({ hasText: tagName });
      await row.getByRole('button', { name: /編集|Edit/i }).click();
    } else {
      // Try card
      const card = this.tagCards.filter({ hasText: tagName });
      await card.getByRole('button', { name: /編集|Edit/i }).click();
    }
    await expect(this.nameInput).toBeVisible();
  }

  /**
   * Edit a tag
   */
  async editTag(tagName: string, newData: Partial<TagFormData>): Promise<void> {
    await this.openEditForm(tagName);
    if (newData.name) await this.nameInput.fill(newData.name);
    if (newData.color) await this.colorInput.fill(newData.color);
    await this.saveButton.click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Delete a tag
   */
  async deleteTag(tagName: string): Promise<void> {
    // Try table row
    if (await this.tagTable.isVisible().catch(() => false)) {
      const row = this.tagRows.filter({ hasText: tagName });
      await row.getByRole('button', { name: /削除|Delete/i }).click();
    } else {
      // Try card
      const card = this.tagCards.filter({ hasText: tagName });
      await card.getByRole('button', { name: /削除|Delete/i }).click();
    }
    await this.confirmDialog();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Check if tag exists
   */
  async tagExists(tagName: string): Promise<boolean> {
    const tags = await this.getTags();
    return tags.some((t) => t.name === tagName);
  }

  /**
   * Get tag by name
   */
  async getTagByName(tagName: string): Promise<{ name: string; color: string } | null> {
    const tags = await this.getTags();
    return tags.find((t) => t.name === tagName) || null;
  }

  /**
   * Check if create button is visible
   */
  async isCreateButtonVisible(): Promise<boolean> {
    return await this.createTagButton.isVisible().catch(() => false);
  }

  /**
   * Cancel form
   */
  async cancelForm(): Promise<void> {
    await this.cancelButton.click();
  }
}
