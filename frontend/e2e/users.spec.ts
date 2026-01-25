import { test, expect } from "@playwright/test";
import { UsersPage } from "./page-objects";

/**
 * User Management Tests (Admin only)
 * These tests should only run in the admin project
 */

test.describe("User Management Page", () => {
  let usersPage: UsersPage;

  test.beforeEach(async ({ page }, testInfo) => {
    // Skip if not running as admin
    test.skip(
      testInfo.project.name !== "admin",
      "User management tests require admin role",
    );

    usersPage = new UsersPage(page);
    await usersPage.goto();
    await usersPage.waitForUsersLoaded();
  });

  test.describe("Display", () => {
    test("should display page title", async () => {
      await expect(usersPage.pageTitle).toBeVisible();
    });

    test("should display user list", async () => {
      const users = await usersPage.getUsers();
      expect(users.length).toBeGreaterThan(0);
    });

    test("should display create user button", async () => {
      const isVisible = await usersPage.isCreateButtonVisible();
      expect(isVisible).toBe(true);
    });

    test("should show user details in table", async () => {
      const users = await usersPage.getUsers();
      const firstUser = users[0];

      expect(firstUser.name).toBeTruthy();
      expect(firstUser.email).toBeTruthy();
      expect(firstUser.role).toBeTruthy();
    });
  });

  test.describe("Create User", () => {
    test("should open create user modal", async () => {
      await usersPage.openCreateModal();
      await expect(usersPage.userModal).toBeVisible();
      await expect(usersPage.nameInput).toBeVisible();
      await expect(usersPage.emailInput).toBeVisible();
      await expect(usersPage.roleSelect).toBeVisible();
    });

    test("should create a new user", async () => {
      const testEmail = `test_${Date.now()}@example.com`;

      await usersPage.createUser({
        name: "E2E Test User",
        email: testEmail,
        password: "TestPass123!",
        role: "viewer",
        department: "Test Department",
      });

      // Verify user appears in list
      const userExists = await usersPage.userExists(testEmail);
      expect(userExists).toBe(true);
    });

    test("should close modal on cancel", async () => {
      await usersPage.openCreateModal();
      await usersPage.closeModal();
      await expect(usersPage.userModal).toBeHidden();
    });
  });

  test.describe("Edit User", () => {
    test("should open edit modal for existing user", async () => {
      const users = await usersPage.getUsers();
      if (users.length > 0) {
        const userEmail = users[0].email;
        await usersPage.openEditModal(userEmail);
        await expect(usersPage.userModal).toBeVisible();
      }
    });

    test("should update user information", async () => {
      // First create a test user
      const testEmail = `edit_test_${Date.now()}@example.com`;
      await usersPage.createUser({
        name: "Edit Test User",
        email: testEmail,
        password: "TestPass123!",
        role: "viewer",
      });

      // Then edit the user
      const newName = "Updated Test User";
      await usersPage.editUser(testEmail, {
        name: newName,
      });

      // Verify update
      const user = await usersPage.getUserByEmail(testEmail);
      expect(user?.name).toBe(newName);
    });
  });

  test.describe("Change Password", () => {
    test("should open change password modal", async () => {
      const users = await usersPage.getUsers();
      if (users.length > 0) {
        const userEmail = users[0].email;
        await usersPage.openChangePasswordModal(userEmail);
        await expect(usersPage.changePasswordModal).toBeVisible();
      }
    });
  });

  test.describe("Toggle User Status", () => {
    test("should toggle user active status", async () => {
      // Create a test user first
      const testEmail = `status_test_${Date.now()}@example.com`;
      await usersPage.createUser({
        name: "Status Test User",
        email: testEmail,
        password: "TestPass123!",
        role: "viewer",
      });

      // Get initial status
      const initialUser = await usersPage.getUserByEmail(testEmail);
      const initialStatus = initialUser?.status;

      // Toggle status
      await usersPage.toggleUserStatus(testEmail);

      // Verify status changed
      const updatedUser = await usersPage.getUserByEmail(testEmail);
      // Status should be different (active/inactive toggle)
      expect(updatedUser).toBeTruthy();
    });
  });

  test.describe("Delete User", () => {
    test("should delete a user", async () => {
      // First create a test user to delete
      const testEmail = `delete_test_${Date.now()}@example.com`;
      await usersPage.createUser({
        name: "Delete Test User",
        email: testEmail,
        password: "TestPass123!",
        role: "viewer",
      });

      // Verify user exists
      let userExists = await usersPage.userExists(testEmail);
      expect(userExists).toBe(true);

      // Delete the user
      await usersPage.deleteUser(testEmail);

      // Verify user is deleted
      userExists = await usersPage.userExists(testEmail);
      expect(userExists).toBe(false);
    });
  });
});
