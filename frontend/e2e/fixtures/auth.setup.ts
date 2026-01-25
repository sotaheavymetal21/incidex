import { test as setup, expect } from '@playwright/test';
import { testUsers, TestUserRole } from './test-users';
import path from 'path';

const STORAGE_STATE_DIR = path.join(__dirname, '..', 'storage-states');

// Run tests sequentially to avoid rate limiting
setup.describe.configure({ mode: 'serial' });

/**
 * Authenticates a user and saves the storage state for reuse
 */
async function authenticateUser(
  page: import('@playwright/test').Page,
  role: TestUserRole
): Promise<void> {
  const user = testUsers[role];

  await page.goto('/login');

  // Wait for login form to be ready
  await expect(page.getByLabel(/メール|Email/i)).toBeVisible({ timeout: 10000 });

  // Fill credentials
  await page.getByLabel(/メール|Email/i).fill(user.email);
  await page.getByLabel(/パスワード|Password/i).fill(user.password);

  // Submit login
  await page.getByRole('button', { name: /ログイン|Login/i }).click();

  // Wait for successful login - redirect to dashboard or home
  // Check for navigation away from login page and wait for dashboard elements
  await expect(page).not.toHaveURL(/\/login/, { timeout: 20000 });

  // Extra wait for page to stabilize
  await page.waitForLoadState('networkidle');

  // Verify user is logged in by checking for user-related UI elements
  await expect(page.locator('nav, [data-testid="sidebar"], header').first()).toBeVisible({ timeout: 10000 });
}

// Setup for admin user
setup('authenticate as admin', async ({ page }) => {
  await authenticateUser(page, 'admin');
  await page.context().storageState({
    path: path.join(STORAGE_STATE_DIR, 'admin.json')
  });
});

// Setup for editor user
setup('authenticate as editor', async ({ page }) => {
  // Longer delay to avoid rate limiting (5 requests per minute limit)
  await page.waitForTimeout(15000);
  await authenticateUser(page, 'editor');
  await page.context().storageState({
    path: path.join(STORAGE_STATE_DIR, 'editor.json')
  });
});

// Setup for viewer user
setup('authenticate as viewer', async ({ page }) => {
  // Longer delay to avoid rate limiting (5 requests per minute limit)
  await page.waitForTimeout(15000);
  await authenticateUser(page, 'viewer');
  await page.context().storageState({
    path: path.join(STORAGE_STATE_DIR, 'viewer.json')
  });
});
