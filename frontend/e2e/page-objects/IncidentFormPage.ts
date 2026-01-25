import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './BasePage';

export type Severity = 'critical' | 'high' | 'medium' | 'low';
export type Status = 'open' | 'investigating' | 'resolved' | 'closed';

export interface IncidentFormData {
  title: string;
  description?: string;
  severity?: Severity;
  status?: Status;
  impactScope?: string;
  assigneeId?: number;
  tagIds?: number[];
}

/**
 * Incident Create/Edit Form Page Object
 */
export class IncidentFormPage extends BasePage {
  // Form fields
  readonly titleInput = this.page.locator('#title');
  readonly descriptionInput = this.page.locator('#description');
  readonly severitySelect = this.page.locator('#severity');
  readonly statusSelect = this.page.locator('#status');
  readonly impactScopeInput = this.page.locator('#impactScope');
  readonly assigneeSelect = this.page.locator('#assignee');

  // Tag selection
  readonly tagCheckboxes = this.page.locator('input[type="checkbox"][id^="tag-"]');

  // Form actions
  readonly submitButton = this.page.getByRole('button', { name: /インシデントを作成|更新|保存/i });
  readonly cancelButton = this.page.getByRole('link', { name: /戻る|キャンセル/i });

  // Validation errors
  readonly titleError = this.page.locator('[data-testid="title-error"], #title + .error, #title ~ p[class*="error"]');
  readonly formError = this.page.locator('[data-testid="form-error"], .form-error, [role="alert"]');

  constructor(page: Page) {
    super(page);
  }

  async goto(): Promise<void> {
    await this.page.goto('/incidents/create');
    await this.waitForPageLoad();
  }

  async gotoEdit(incidentId: number): Promise<void> {
    await this.page.goto(`/incidents/${incidentId}/edit`);
    await this.waitForPageLoad();
  }

  /**
   * Fill the title field
   */
  async fillTitle(title: string): Promise<void> {
    await this.titleInput.fill(title);
  }

  /**
   * Fill the description field
   */
  async fillDescription(description: string): Promise<void> {
    await this.descriptionInput.fill(description);
  }

  /**
   * Select severity
   */
  async selectSeverity(severity: Severity): Promise<void> {
    await this.severitySelect.selectOption(severity);
  }

  /**
   * Select status
   */
  async selectStatus(status: Status): Promise<void> {
    await this.statusSelect.selectOption(status);
  }

  /**
   * Fill impact scope
   */
  async fillImpactScope(impactScope: string): Promise<void> {
    await this.impactScopeInput.fill(impactScope);
  }

  /**
   * Select assignee by ID
   */
  async selectAssignee(assigneeId: number | string): Promise<void> {
    await this.assigneeSelect.selectOption(String(assigneeId));
  }

  /**
   * Select tags by name
   */
  async selectTags(tagNames: string[]): Promise<void> {
    for (const tagName of tagNames) {
      const checkbox = this.page.getByLabel(tagName, { exact: true });
      if (await checkbox.isVisible()) {
        await checkbox.check();
      }
    }
  }

  /**
   * Deselect a tag by name
   */
  async deselectTag(tagName: string): Promise<void> {
    const checkbox = this.page.getByLabel(tagName, { exact: true });
    if (await checkbox.isVisible()) {
      await checkbox.uncheck();
    }
  }

  /**
   * Fill the entire form
   */
  async fillForm(data: IncidentFormData): Promise<void> {
    await this.fillTitle(data.title);

    if (data.description) {
      await this.fillDescription(data.description);
    }

    if (data.severity) {
      await this.selectSeverity(data.severity);
    }

    if (data.status) {
      await this.selectStatus(data.status);
    }

    if (data.impactScope) {
      await this.fillImpactScope(data.impactScope);
    }

    if (data.assigneeId) {
      await this.selectAssignee(data.assigneeId);
    }
  }

  /**
   * Submit the form
   */
  async submit(): Promise<void> {
    await this.submitButton.click();
  }

  /**
   * Create incident and wait for redirect to detail page
   */
  async createIncident(data: IncidentFormData): Promise<void> {
    await this.fillForm(data);
    await this.submit();
    await expect(this.page).toHaveURL(/\/incidents\/\d+$/, { timeout: 10000 });
  }

  /**
   * Update incident and wait for redirect to detail page
   */
  async updateIncident(data: Partial<IncidentFormData>): Promise<void> {
    if (data.title) await this.fillTitle(data.title);
    if (data.description) await this.fillDescription(data.description);
    if (data.severity) await this.selectSeverity(data.severity);
    if (data.status) await this.selectStatus(data.status);
    if (data.impactScope) await this.fillImpactScope(data.impactScope);
    if (data.assigneeId) await this.selectAssignee(data.assigneeId);

    await this.submit();
    await expect(this.page).toHaveURL(/\/incidents\/\d+$/, { timeout: 10000 });
  }

  /**
   * Cancel and go back
   */
  async cancel(): Promise<void> {
    await this.cancelButton.click();
  }

  /**
   * Check if title error is displayed
   */
  async getTitleError(): Promise<string | null> {
    if (await this.titleError.isVisible().catch(() => false)) {
      return await this.titleError.textContent();
    }
    return null;
  }

  /**
   * Check if form error is displayed
   */
  async getFormError(): Promise<string | null> {
    if (await this.formError.isVisible().catch(() => false)) {
      return await this.formError.textContent();
    }
    return null;
  }

  /**
   * Get current form values
   */
  async getFormValues(): Promise<IncidentFormData> {
    return {
      title: await this.titleInput.inputValue(),
      description: await this.descriptionInput.inputValue(),
      severity: (await this.severitySelect.inputValue()) as Severity,
      status: (await this.statusSelect.inputValue()) as Status,
      impactScope: await this.impactScopeInput.inputValue(),
    };
  }

  /**
   * Check if form is in valid state (no visible errors)
   */
  async isFormValid(): Promise<boolean> {
    const titleError = await this.getTitleError();
    const formError = await this.getFormError();
    return !titleError && !formError;
  }

  /**
   * Get available severity options
   */
  async getSeverityOptions(): Promise<string[]> {
    const options = await this.severitySelect.locator('option').all();
    const values: string[] = [];
    for (const option of options) {
      const value = await option.getAttribute('value');
      if (value) values.push(value);
    }
    return values;
  }

  /**
   * Get available status options
   */
  async getStatusOptions(): Promise<string[]> {
    const options = await this.statusSelect.locator('option').all();
    const values: string[] = [];
    for (const option of options) {
      const value = await option.getAttribute('value');
      if (value) values.push(value);
    }
    return values;
  }
}
