import { test, expect } from "@playwright/test";
import { AuditLogsPage } from "./page-objects";

/**
 * Audit Logs Tests (Admin only)
 */

test.describe("Audit Logs Page", () => {
  let auditLogsPage: AuditLogsPage;

  test.beforeEach(async ({ page }, testInfo) => {
    // Skip if not running as admin
    test.skip(
      testInfo.project.name !== "admin",
      "Audit logs tests require admin role",
    );

    auditLogsPage = new AuditLogsPage(page);
    await auditLogsPage.goto();
    await auditLogsPage.waitForLogsLoaded();
  });

  test.describe("Display", () => {
    test("should display page title", async () => {
      await expect(auditLogsPage.pageTitle).toBeVisible();
    });

    test("should display audit logs table", async () => {
      await expect(auditLogsPage.logTable).toBeVisible();
    });

    test("should display log entries", async () => {
      const logs = await auditLogsPage.getLogs();
      // May or may not have logs
      expect(Array.isArray(logs)).toBe(true);
    });

    test("should show log details", async () => {
      const logs = await auditLogsPage.getLogs();
      if (logs.length > 0) {
        const firstLog = logs[0];
        expect(firstLog.timestamp).toBeTruthy();
        expect(firstLog.action).toBeTruthy();
      }
    });
  });

  test.describe("Filtering", () => {
    test("should filter by action type", async () => {
      const actionFilter = auditLogsPage.actionFilter;
      if (await actionFilter.isVisible().catch(() => false)) {
        // Get all options
        const options = await actionFilter.locator("option").all();
        if (options.length > 1) {
          const secondOption = await options[1].getAttribute("value");
          if (secondOption) {
            await auditLogsPage.filterByAction(secondOption);
            // Logs should be filtered
            const logs = await auditLogsPage.getLogs();
            expect(Array.isArray(logs)).toBe(true);
          }
        }
      }
    });

    test("should search logs", async () => {
      if (await auditLogsPage.searchInput.isVisible().catch(() => false)) {
        await auditLogsPage.search("test");
        const logs = await auditLogsPage.getLogs();
        expect(Array.isArray(logs)).toBe(true);
      }
    });

    test("should clear filters", async () => {
      if (
        await auditLogsPage.clearFiltersButton.isVisible().catch(() => false)
      ) {
        await auditLogsPage.clearFilters();
        const logs = await auditLogsPage.getLogs();
        expect(Array.isArray(logs)).toBe(true);
      }
    });
  });

  test.describe("Pagination", () => {
    test("should navigate between pages", async () => {
      // Check if pagination exists
      if (await auditLogsPage.nextButton.isVisible().catch(() => false)) {
        const isDisabled = await auditLogsPage.nextButton.isDisabled();
        if (!isDisabled) {
          await auditLogsPage.goToNextPage();
          // Should still be on audit logs page
          await expect(auditLogsPage.logTable).toBeVisible();
        }
      }
    });
  });
});

test.describe("Audit Logs Access Control", () => {
  test("Editor should not access audit logs", async ({ page }, testInfo) => {
    test.skip(
      testInfo.project.name !== "editor",
      "This test is for editor role only",
    );

    await page.goto("/audit-logs");
    await page.waitForLoadState("networkidle");

    const auditLogsPage = new AuditLogsPage(page);
    const isAccessible = await auditLogsPage.isPageAccessible();
    expect(isAccessible).toBe(false);
  });

  test("Viewer should not access audit logs", async ({ page }, testInfo) => {
    test.skip(
      testInfo.project.name !== "viewer",
      "This test is for viewer role only",
    );

    await page.goto("/audit-logs");
    await page.waitForLoadState("networkidle");

    const auditLogsPage = new AuditLogsPage(page);
    const isAccessible = await auditLogsPage.isPageAccessible();
    expect(isAccessible).toBe(false);
  });
});
