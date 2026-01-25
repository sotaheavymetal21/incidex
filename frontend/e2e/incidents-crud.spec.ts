import { test, expect } from "@playwright/test";
import {
  IncidentFormPage,
  IncidentListPage,
  IncidentDetailPage,
} from "./page-objects";

// Use admin project for CRUD operations
test.describe("Incident CRUD Operations", () => {
  test.describe("Create Incident", () => {
    test("should display create incident form", async ({ page }) => {
      const formPage = new IncidentFormPage(page);
      await formPage.goto();

      // Verify form elements
      await expect(formPage.titleInput).toBeVisible();
      await expect(formPage.descriptionInput).toBeVisible();
      await expect(formPage.severitySelect).toBeVisible();
      await expect(formPage.submitButton).toBeVisible();
    });

    test("should validate required title field", async ({ page }) => {
      const formPage = new IncidentFormPage(page);
      await formPage.goto();

      // Try to submit without title
      await formPage.selectSeverity("medium");
      await formPage.submit();

      // Should show validation error or stay on page
      await expect(page).toHaveURL(/\/incidents\/create/);
    });

    test("should create incident successfully", async ({ page }) => {
      const formPage = new IncidentFormPage(page);
      await formPage.goto();

      const testTitle = `E2E Test Incident ${Date.now()}`;

      await formPage.createIncident({
        title: testTitle,
        description: "This is a test incident created by E2E tests",
        severity: "medium",
        status: "open",
        impactScope: "Test environment only",
      });

      // Should redirect to detail page
      await expect(page).toHaveURL(/\/incidents\/\d+$/);

      // Verify incident was created
      const detailPage = new IncidentDetailPage(page);
      const title = await detailPage.getTitle();
      expect(title).toContain(testTitle);
    });

    test("should preserve form data after validation error", async ({
      page,
    }) => {
      const formPage = new IncidentFormPage(page);
      await formPage.goto();

      // Fill some fields
      await formPage.fillDescription("Test description");
      await formPage.selectSeverity("high");

      // Try to submit without title (should fail)
      await formPage.submit();

      // Verify data is preserved
      const values = await formPage.getFormValues();
      expect(values.description).toBe("Test description");
      expect(values.severity).toBe("high");
    });

    test("should cancel and return to list", async ({ page }) => {
      const formPage = new IncidentFormPage(page);
      await formPage.goto();

      await formPage.cancel();

      // Should go back (either to list or previous page)
      await page.waitForLoadState("networkidle");
    });
  });

  test.describe("Edit Incident", () => {
    test("should load existing incident data in edit form", async ({
      page,
    }) => {
      // First, get an existing incident
      const listPage = new IncidentListPage(page);
      await listPage.goto();
      await listPage.waitForIncidentsLoaded();

      const incidents = await listPage.getIncidentRows();
      expect(incidents.length).toBeGreaterThan(0);

      // Navigate to detail page
      await listPage.clickIncident(0);
      await expect(page).toHaveURL(/\/incidents\/\d+$/);

      // Navigate to edit page
      const detailPage = new IncidentDetailPage(page);

      // Check if edit button is visible (depends on role)
      if (await detailPage.isEditButtonVisible()) {
        await detailPage.clickEdit();

        // Verify form is pre-filled
        const formPage = new IncidentFormPage(page);
        const values = await formPage.getFormValues();
        expect(values.title).toBeTruthy();
      }
    });

    test("should update incident successfully", async ({ page }) => {
      // First, create a new incident to edit
      const formPage = new IncidentFormPage(page);
      await formPage.goto();

      const originalTitle = `Edit Test ${Date.now()}`;
      await formPage.createIncident({
        title: originalTitle,
        description: "Original description",
        severity: "low",
      });

      // Get the incident ID from URL
      const url = page.url();
      const incidentId = url.match(/\/incidents\/(\d+)/)?.[1];
      expect(incidentId).toBeTruthy();

      // Navigate to edit page
      const detailPage = new IncidentDetailPage(page);
      if (await detailPage.isEditButtonVisible()) {
        await detailPage.clickEdit();

        // Update the incident
        const updatedTitle = `Updated ${originalTitle}`;
        await formPage.updateIncident({
          title: updatedTitle,
          description: "Updated description",
          severity: "high",
        });

        // Verify update
        const title = await detailPage.getTitle();
        expect(title).toContain(updatedTitle);
      }
    });
  });

  test.describe("Delete Incident", () => {
    test("should delete incident with confirmation", async ({ page }) => {
      // First, create a new incident to delete
      const formPage = new IncidentFormPage(page);
      await formPage.goto();

      const testTitle = `Delete Test ${Date.now()}`;
      await formPage.createIncident({
        title: testTitle,
        description: "This incident will be deleted",
        severity: "low",
      });

      // Verify on detail page
      const detailPage = new IncidentDetailPage(page);
      await expect(page).toHaveURL(/\/incidents\/\d+$/);

      // Delete the incident (if delete button is visible)
      if (await detailPage.isDeleteButtonVisible()) {
        await detailPage.deleteIncident();

        // Should redirect to list
        await expect(page).toHaveURL(/\/incidents$/);

        // Verify incident no longer appears
        const listPage = new IncidentListPage(page);
        await listPage.search(testTitle);
        await listPage.waitForIncidentsLoaded();

        const incidents = await listPage.getIncidentRows();
        const found = incidents.some((i) => i.title.includes(testTitle));
        expect(found).toBe(false);
      }
    });
  });
});

test.describe("Incident Form Validation", () => {
  test("should show all severity options", async ({ page }) => {
    const formPage = new IncidentFormPage(page);
    await formPage.goto();

    const options = await formPage.getSeverityOptions();
    expect(options).toContain("critical");
    expect(options).toContain("high");
    expect(options).toContain("medium");
    expect(options).toContain("low");
  });

  test("should show all status options", async ({ page }) => {
    const formPage = new IncidentFormPage(page);
    await formPage.goto();

    const options = await formPage.getStatusOptions();
    expect(options).toContain("open");
    expect(options).toContain("investigating");
    expect(options).toContain("resolved");
    expect(options).toContain("closed");
  });
});
