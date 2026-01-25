import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './BasePage';

export type UserRole = 'admin' | 'editor' | 'viewer';

export interface UserFormData {
  name: string;
  email: string;
  password?: string;
  role: UserRole;
  employeeNumber?: string;
  department?: string;
}

/**
 * Users Management Page Object (Admin only)
 */
export class UsersPage extends BasePage {
  // Header elements
  readonly pageTitle = this.page.getByRole('heading', { name: 'ユーザー管理' });
  readonly createUserButton = this.page.getByRole('button', { name: 'ユーザーを作成' });

  // User table
  readonly userTable = this.page.locator('table');
  readonly userRows = this.page.locator('table tbody tr');

  // Create/Edit Modal
  readonly userModal = this.page.locator('.fixed.inset-0, [role="dialog"]').filter({ has: this.page.locator('h2') });
  readonly nameInput = this.page.locator('input#name, input[name="name"]');
  readonly emailInput = this.page.locator('input#email, input[name="email"]');
  readonly passwordInput = this.page.locator('input#password, input[name="password"]');
  readonly roleSelect = this.page.locator('select#role, select[name="role"]');
  readonly employeeNumberInput = this.page.locator('input#employeeNumber, input[name="employeeNumber"]');
  readonly departmentInput = this.page.locator('input#department, input[name="department"]');
  readonly saveButton = this.page.getByRole('button', { name: '作成' });
  readonly modalCancelButton = this.page.getByRole('button', { name: 'キャンセル' });

  // Change Password Modal
  readonly changePasswordModal = this.page.locator('[role="dialog"]:has-text("パスワード")');
  readonly newPasswordInput = this.changePasswordModal.locator('input[name="newPassword"], #newPassword');
  readonly confirmPasswordInput = this.changePasswordModal.locator('input[name="confirmPassword"], #confirmPassword');
  readonly changePasswordButton = this.changePasswordModal.getByRole('button', { name: /変更|Change|更新/i });

  constructor(page: Page) {
    super(page);
  }

  async goto(): Promise<void> {
    await this.page.goto('/users');
    await this.waitForPageLoad();
  }

  /**
   * Wait for user list to load
   */
  async waitForUsersLoaded(): Promise<void> {
    await this.page.waitForLoadState('networkidle');
    await this.userRows.first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {
      // May have no users
    });
  }

  /**
   * Get all users from the table
   */
  async getUsers(): Promise<
    Array<{
      name: string;
      email: string;
      role: string;
      status: string;
    }>
  > {
    await this.waitForUsersLoaded();
    const rows = await this.userRows.all();
    const users: Array<{
      name: string;
      email: string;
      role: string;
      status: string;
    }> = [];

    for (const row of rows) {
      const cells = await row.locator('td').all();
      users.push({
        name: (await cells[0]?.textContent())?.trim() || '',
        email: (await cells[1]?.textContent())?.trim() || '',
        role: (await cells[2]?.textContent())?.trim() || '',
        status: (await cells[3]?.textContent())?.trim() || '',
      });
    }

    return users;
  }

  /**
   * Open create user modal
   */
  async openCreateModal(): Promise<void> {
    await this.createUserButton.click();
    await expect(this.userModal).toBeVisible();
  }

  /**
   * Fill user form
   */
  async fillUserForm(data: UserFormData): Promise<void> {
    await this.nameInput.fill(data.name);
    await this.emailInput.fill(data.email);
    if (data.password) {
      await this.passwordInput.fill(data.password);
    }
    await this.roleSelect.selectOption(data.role);
    if (data.employeeNumber) {
      await this.employeeNumberInput.fill(data.employeeNumber);
    }
    if (data.department) {
      await this.departmentInput.fill(data.department);
    }
  }

  /**
   * Create a new user
   */
  async createUser(data: UserFormData): Promise<void> {
    await this.openCreateModal();
    await this.fillUserForm(data);
    await this.saveButton.click();
    await expect(this.userModal).toBeHidden();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Open edit modal for a user
   */
  async openEditModal(email: string): Promise<void> {
    const row = this.userRows.filter({ hasText: email });
    await row.getByRole('button', { name: /編集|Edit/i }).click();
    await expect(this.userModal).toBeVisible();
  }

  /**
   * Edit a user
   */
  async editUser(email: string, data: Partial<UserFormData>): Promise<void> {
    await this.openEditModal(email);
    if (data.name) await this.nameInput.fill(data.name);
    if (data.role) await this.roleSelect.selectOption(data.role);
    if (data.employeeNumber) await this.employeeNumberInput.fill(data.employeeNumber);
    if (data.department) await this.departmentInput.fill(data.department);
    await this.saveButton.click();
    await expect(this.userModal).toBeHidden();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Open change password modal
   */
  async openChangePasswordModal(email: string): Promise<void> {
    const row = this.userRows.filter({ hasText: email });
    await row.getByRole('button', { name: /パスワード|Password/i }).click();
    await expect(this.changePasswordModal).toBeVisible();
  }

  /**
   * Change user password
   */
  async changePassword(email: string, newPassword: string): Promise<void> {
    await this.openChangePasswordModal(email);
    await this.newPasswordInput.fill(newPassword);
    await this.confirmPasswordInput.fill(newPassword);
    await this.changePasswordButton.click();
    await expect(this.changePasswordModal).toBeHidden();
  }

  /**
   * Toggle user active status
   */
  async toggleUserStatus(email: string): Promise<void> {
    const row = this.userRows.filter({ hasText: email });
    const toggleButton = row.getByRole('button', { name: /有効|無効|Enable|Disable|Active|Inactive/i });
    await toggleButton.click();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Delete a user
   */
  async deleteUser(email: string): Promise<void> {
    const row = this.userRows.filter({ hasText: email });
    await row.getByRole('button', { name: /削除|Delete/i }).click();
    await this.confirmDialog();
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Check if user exists in the table
   */
  async userExists(email: string): Promise<boolean> {
    const row = this.userRows.filter({ hasText: email });
    return await row.count() > 0;
  }

  /**
   * Get user by email
   */
  async getUserByEmail(
    email: string
  ): Promise<{ name: string; email: string; role: string; status: string } | null> {
    const users = await this.getUsers();
    return users.find((u) => u.email === email) || null;
  }

  /**
   * Check if create button is visible
   */
  async isCreateButtonVisible(): Promise<boolean> {
    return await this.createUserButton.isVisible().catch(() => false);
  }

  /**
   * Close modal
   */
  async closeModal(): Promise<void> {
    await this.modalCancelButton.click();
    await expect(this.userModal).toBeHidden();
  }
}
