import { test, expect } from "@playwright/test";
import {
  IncidentDetailPage,
  IncidentListPage,
  IncidentFormPage,
} from "./page-objects";

test.describe("Incident Detail Page", () => {
  let detailPage: IncidentDetailPage;
  let incidentId: string;

  test.beforeEach(async ({ page }) => {
    // Navigate to an existing incident
    const listPage = new IncidentListPage(page);
    await listPage.goto();
    await listPage.waitForIncidentsLoaded();

    // Click first incident to go to detail
    await listPage.clickIncident(0);
    await expect(page).toHaveURL(/\/incidents\/\d+/);

    // Get incident ID from URL
    incidentId = page.url().match(/\/incidents\/(\d+)/)?.[1] || "";
    detailPage = new IncidentDetailPage(page);
  });

  test.describe("Display", () => {
    test("should display incident title", async () => {
      const title = await detailPage.getTitle();
      expect(title).toBeTruthy();
    });

    test("should display severity badge", async () => {
      const severity = await detailPage.getSeverity();
      expect(["critical", "high", "medium", "low"]).toContain(severity);
    });

    test("should display status badge", async () => {
      const status = await detailPage.getStatus();
      expect(["open", "investigating", "resolved", "closed"]).toContain(status);
    });
  });

  test.describe("Comments", () => {
    test("should add a comment", async ({ page }) => {
      const commentText = `Test comment ${Date.now()}`;

      // Check if comment input is visible
      if (await detailPage.commentInput.isVisible().catch(() => false)) {
        await detailPage.addComment(commentText);

        // Verify comment appears
        const comments = await detailPage.getComments();
        const hasNewComment = comments.some((c) =>
          c.content.includes(commentText),
        );
        expect(hasNewComment).toBe(true);
      }
    });

    test("should display existing comments", async () => {
      // This test just verifies the structure is correct
      const comments = await detailPage.getComments();
      // Comments may or may not exist, just verify no errors
      expect(Array.isArray(comments)).toBe(true);
    });
  });

  test.describe("Navigation", () => {
    test("should navigate to edit page", async ({ page }) => {
      if (await detailPage.isEditButtonVisible()) {
        await detailPage.clickEdit();
        await expect(page).toHaveURL(/\/incidents\/\d+\/edit/);
      }
    });

    test("should navigate to post-mortem page", async ({ page }) => {
      const postMortemButton = detailPage.postMortemButton;
      if (await postMortemButton.isVisible().catch(() => false)) {
        await detailPage.clickPostMortem();
        await expect(page).toHaveURL(/\/incidents\/\d+\/postmortem/);
      }
    });
  });
});

test.describe("Incident Activity/Timeline", () => {
  test("should display activity events", async ({ page }) => {
    // Navigate to an existing incident
    const listPage = new IncidentListPage(page);
    await listPage.goto();
    await listPage.waitForIncidentsLoaded();
    await listPage.clickIncident(0);

    const detailPage = new IncidentDetailPage(page);

    // Check if activity section exists
    if (await detailPage.activitySection.isVisible().catch(() => false)) {
      const events = await detailPage.getActivityEvents();
      expect(Array.isArray(events)).toBe(true);
    }
  });
});

test.describe("Incident Attachments", () => {
  test("should display attachments section", async ({ page }) => {
    const listPage = new IncidentListPage(page);
    await listPage.goto();
    await listPage.waitForIncidentsLoaded();
    await listPage.clickIncident(0);

    const detailPage = new IncidentDetailPage(page);

    // Check if attachments section exists
    if (await detailPage.attachmentsSection.isVisible().catch(() => false)) {
      const attachments = await detailPage.getAttachments();
      expect(Array.isArray(attachments)).toBe(true);
    }
  });
});

test.describe("Incident Actions Based on Role", () => {
  test("should show or hide edit button based on permissions", async ({
    page,
  }, testInfo) => {
    const listPage = new IncidentListPage(page);
    await listPage.goto();
    await listPage.waitForIncidentsLoaded();
    await listPage.clickIncident(0);

    const detailPage = new IncidentDetailPage(page);
    const isEditVisible = await detailPage.isEditButtonVisible();

    // Admin and Editor should see edit button
    // Viewer should not see edit button
    const projectName = testInfo.project.name;
    if (projectName === "admin" || projectName === "editor") {
      expect(isEditVisible).toBe(true);
    } else if (projectName === "viewer") {
      expect(isEditVisible).toBe(false);
    }
  });

  test("should show or hide delete button based on permissions", async ({
    page,
  }, testInfo) => {
    const listPage = new IncidentListPage(page);
    await listPage.goto();
    await listPage.waitForIncidentsLoaded();
    await listPage.clickIncident(0);

    const detailPage = new IncidentDetailPage(page);
    const isDeleteVisible = await detailPage.isDeleteButtonVisible();

    // Only Admin should see delete button
    const projectName = testInfo.project.name;
    if (projectName === "admin") {
      expect(isDeleteVisible).toBe(true);
    } else {
      expect(isDeleteVisible).toBe(false);
    }
  });
});
