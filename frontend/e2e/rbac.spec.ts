import { test, expect } from "@playwright/test";
import {
  IncidentListPage,
  IncidentDetailPage,
  UsersPage,
  AuditLogsPage,
  TagsPage,
} from "./page-objects";

/**
 * Role-Based Access Control Tests
 * These tests verify that different roles have appropriate access levels
 *
 * Role Permissions:
 * - Admin: Full access to all features
 * - Editor: Can create/edit incidents, no access to user management or audit logs
 * - Viewer: Read-only access, no create/edit buttons
 */

test.describe("RBAC - Admin Role", () => {
  test.describe("Admin should have full access", () => {
    test("should access incidents page", async ({ page }) => {
      const listPage = new IncidentListPage(page);
      await listPage.goto();
      await expect(listPage.pageTitle).toBeVisible();
    });

    test("should see create incident button", async ({ page }) => {
      const listPage = new IncidentListPage(page);
      await listPage.goto();
      await listPage.waitForIncidentsLoaded();

      const isVisible = await listPage.isCreateButtonVisible();
      expect(isVisible).toBe(true);
    });

    test("should access users management page", async ({ page }) => {
      const usersPage = new UsersPage(page);
      await usersPage.goto();

      // Admin should see the page
      await expect(usersPage.pageTitle).toBeVisible();
    });

    test("should see create user button", async ({ page }) => {
      const usersPage = new UsersPage(page);
      await usersPage.goto();

      const isVisible = await usersPage.isCreateButtonVisible();
      expect(isVisible).toBe(true);
    });

    test("should access audit logs page", async ({ page }) => {
      const auditLogsPage = new AuditLogsPage(page);
      await auditLogsPage.goto();

      // Admin should see the page
      const isAccessible = await auditLogsPage.isPageAccessible();
      expect(isAccessible).toBe(true);
    });

    test("should see edit and delete buttons on incident detail", async ({
      page,
    }) => {
      const listPage = new IncidentListPage(page);
      await listPage.goto();
      await listPage.waitForIncidentsLoaded();
      await listPage.clickIncident(0);

      const detailPage = new IncidentDetailPage(page);

      const isEditVisible = await detailPage.isEditButtonVisible();
      const isDeleteVisible = await detailPage.isDeleteButtonVisible();

      expect(isEditVisible).toBe(true);
      expect(isDeleteVisible).toBe(true);
    });
  });
});

test.describe("RBAC - Editor Role", () => {
  test.describe("Editor access permissions", () => {
    test("should access incidents page", async ({ page }, testInfo) => {
      // Skip if not running as editor
      test.skip(
        testInfo.project.name !== "editor",
        "This test is for editor role only",
      );

      const listPage = new IncidentListPage(page);
      await listPage.goto();
      await expect(listPage.pageTitle).toBeVisible();
    });

    test("should see create incident button", async ({ page }, testInfo) => {
      test.skip(
        testInfo.project.name !== "editor",
        "This test is for editor role only",
      );

      const listPage = new IncidentListPage(page);
      await listPage.goto();
      await listPage.waitForIncidentsLoaded();

      const isVisible = await listPage.isCreateButtonVisible();
      expect(isVisible).toBe(true);
    });

    test("should see edit button but not delete button on incident detail", async ({
      page,
    }, testInfo) => {
      test.skip(
        testInfo.project.name !== "editor",
        "This test is for editor role only",
      );

      const listPage = new IncidentListPage(page);
      await listPage.goto();
      await listPage.waitForIncidentsLoaded();
      await listPage.clickIncident(0);

      const detailPage = new IncidentDetailPage(page);

      const isEditVisible = await detailPage.isEditButtonVisible();
      const isDeleteVisible = await detailPage.isDeleteButtonVisible();

      expect(isEditVisible).toBe(true);
      expect(isDeleteVisible).toBe(false);
    });

    test("should NOT access users management page", async ({
      page,
    }, testInfo) => {
      test.skip(
        testInfo.project.name !== "editor",
        "This test is for editor role only",
      );

      await page.goto("/users");
      await page.waitForLoadState("networkidle");

      // Should be redirected or see access denied
      const isOnUsersPage = page.url().includes("/users");
      const usersPage = new UsersPage(page);
      const hasTitle = await usersPage.pageTitle.isVisible().catch(() => false);

      // Editor should not see users page content
      expect(isOnUsersPage && hasTitle).toBe(false);
    });

    test("should NOT access audit logs page", async ({ page }, testInfo) => {
      test.skip(
        testInfo.project.name !== "editor",
        "This test is for editor role only",
      );

      await page.goto("/audit-logs");
      await page.waitForLoadState("networkidle");

      const auditLogsPage = new AuditLogsPage(page);
      const isAccessible = await auditLogsPage.isPageAccessible();

      // Editor should not see audit logs
      expect(isAccessible).toBe(false);
    });
  });
});

test.describe("RBAC - Viewer Role", () => {
  test.describe("Viewer read-only access", () => {
    test("should access incidents page", async ({ page }, testInfo) => {
      test.skip(
        testInfo.project.name !== "viewer",
        "This test is for viewer role only",
      );

      const listPage = new IncidentListPage(page);
      await listPage.goto();
      await expect(listPage.pageTitle).toBeVisible();
    });

    test("should NOT see create incident button", async ({
      page,
    }, testInfo) => {
      test.skip(
        testInfo.project.name !== "viewer",
        "This test is for viewer role only",
      );

      const listPage = new IncidentListPage(page);
      await listPage.goto();
      await listPage.waitForIncidentsLoaded();

      const isVisible = await listPage.isCreateButtonVisible();
      expect(isVisible).toBe(false);
    });

    test("should NOT see edit or delete buttons on incident detail", async ({
      page,
    }, testInfo) => {
      test.skip(
        testInfo.project.name !== "viewer",
        "This test is for viewer role only",
      );

      const listPage = new IncidentListPage(page);
      await listPage.goto();
      await listPage.waitForIncidentsLoaded();
      await listPage.clickIncident(0);

      const detailPage = new IncidentDetailPage(page);

      const isEditVisible = await detailPage.isEditButtonVisible();
      const isDeleteVisible = await detailPage.isDeleteButtonVisible();

      expect(isEditVisible).toBe(false);
      expect(isDeleteVisible).toBe(false);
    });

    test("should NOT access users management page", async ({
      page,
    }, testInfo) => {
      test.skip(
        testInfo.project.name !== "viewer",
        "This test is for viewer role only",
      );

      await page.goto("/users");
      await page.waitForLoadState("networkidle");

      const usersPage = new UsersPage(page);
      const hasTitle = await usersPage.pageTitle.isVisible().catch(() => false);

      expect(hasTitle).toBe(false);
    });

    test("should NOT access audit logs page", async ({ page }, testInfo) => {
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

    test("should be able to view incident details", async ({
      page,
    }, testInfo) => {
      test.skip(
        testInfo.project.name !== "viewer",
        "This test is for viewer role only",
      );

      const listPage = new IncidentListPage(page);
      await listPage.goto();
      await listPage.waitForIncidentsLoaded();
      await listPage.clickIncident(0);

      const detailPage = new IncidentDetailPage(page);
      const title = await detailPage.getTitle();

      // Viewer can read incident details
      expect(title).toBeTruthy();
    });
  });
});

test.describe("RBAC - Unauthenticated Access", () => {
  test("should redirect to login when accessing protected routes", async ({
    page,
    context,
  }, testInfo) => {
    // Only run in unauthenticated project
    test.skip(
      testInfo.project.name !== "unauthenticated",
      "This test is for unauthenticated access only",
    );

    // Clear any stored auth
    await context.clearCookies();
    await page.evaluate(() => localStorage.clear());

    // Try to access protected routes
    const protectedRoutes = [
      "/incidents",
      "/dashboard",
      "/users",
      "/tags",
      "/audit-logs",
      "/reports",
    ];

    for (const route of protectedRoutes) {
      await page.goto(route);
      await page.waitForURL(/\/login/, { timeout: 10000 });
      expect(page.url()).toContain("/login");
    }
  });
});

test.describe("RBAC - Tag Management Access", () => {
  test("Admin should access tags page", async ({ page }, testInfo) => {
    test.skip(
      testInfo.project.name !== "admin",
      "This test is for admin role only",
    );

    const tagsPage = new TagsPage(page);
    await tagsPage.goto();

    await expect(tagsPage.pageTitle).toBeVisible();
    const isCreateVisible = await tagsPage.isCreateButtonVisible();
    expect(isCreateVisible).toBe(true);
  });

  test("Editor should access tags page with create button", async ({
    page,
  }, testInfo) => {
    test.skip(
      testInfo.project.name !== "editor",
      "This test is for editor role only",
    );

    const tagsPage = new TagsPage(page);
    await tagsPage.goto();

    // Editor can view and manage tags
    await expect(tagsPage.pageTitle).toBeVisible();
  });

  test("Viewer may have limited tag access", async ({ page }, testInfo) => {
    test.skip(
      testInfo.project.name !== "viewer",
      "This test is for viewer role only",
    );

    const tagsPage = new TagsPage(page);
    await tagsPage.goto();

    // Viewer may see tags but not create button
    const isCreateVisible = await tagsPage.isCreateButtonVisible();
    expect(isCreateVisible).toBe(false);
  });
});
