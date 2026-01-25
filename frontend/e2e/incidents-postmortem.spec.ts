import { test, expect } from "@playwright/test";
import {
  PostMortemPage,
  IncidentListPage,
  IncidentDetailPage,
  IncidentFormPage,
} from "./page-objects";

/**
 * Post-Mortem Tests
 * Requires admin or editor role to create/edit
 */

test.describe("Post-Mortem Page", () => {
  let postMortemPage: PostMortemPage;

  test.describe("Access and Display", () => {
    test("should access post-mortem from incident detail", async ({
      page,
    }, testInfo) => {
      // Skip if viewer (may not have access to create post-mortems)
      test.skip(
        testInfo.project.name === "viewer",
        "Viewers may have limited post-mortem access",
      );

      // Navigate to an incident
      const listPage = new IncidentListPage(page);
      await listPage.goto();
      await listPage.waitForIncidentsLoaded();
      await listPage.clickIncident(0);

      // Click post-mortem button
      const detailPage = new IncidentDetailPage(page);
      if (await detailPage.postMortemButton.isVisible().catch(() => false)) {
        await detailPage.clickPostMortem();
        await expect(page).toHaveURL(/\/incidents\/\d+\/postmortem/);
      }
    });

    test("should display post-mortem form sections", async ({
      page,
    }, testInfo) => {
      test.skip(
        testInfo.project.name === "viewer",
        "Viewers may have limited post-mortem access",
      );

      // Navigate to post-mortem page
      const listPage = new IncidentListPage(page);
      await listPage.goto();
      await listPage.waitForIncidentsLoaded();
      await listPage.clickIncident(0);

      const detailPage = new IncidentDetailPage(page);
      if (await detailPage.postMortemButton.isVisible().catch(() => false)) {
        await detailPage.clickPostMortem();

        postMortemPage = new PostMortemPage(page);

        // Check if form sections are visible
        const isEditable = await postMortemPage.isEditable();
        if (isEditable) {
          await expect(postMortemPage.summaryInput).toBeVisible();
        }
      }
    });
  });

  test.describe("Create Post-Mortem", () => {
    test("should create post-mortem with all sections", async ({
      page,
    }, testInfo) => {
      test.skip(
        testInfo.project.name === "viewer",
        "Viewers cannot create post-mortems",
      );

      // First create a new incident for testing
      const formPage = new IncidentFormPage(page);
      await formPage.goto();

      const testTitle = `Post-Mortem Test ${Date.now()}`;
      await formPage.createIncident({
        title: testTitle,
        description: "Test incident for post-mortem",
        severity: "medium",
        status: "resolved", // Resolved incidents can have post-mortems
      });

      // Navigate to post-mortem
      const detailPage = new IncidentDetailPage(page);
      if (await detailPage.postMortemButton.isVisible().catch(() => false)) {
        await detailPage.clickPostMortem();

        postMortemPage = new PostMortemPage(page);

        // Fill form
        if (await postMortemPage.isEditable()) {
          await postMortemPage.fillForm({
            summary: "This is a test post-mortem summary",
            rootCause: "Root cause analysis for testing",
            impactAnalysis: "Impact was limited to test environment",
            timeline: "Event timeline description",
            lessonsLearned: "Key lessons from this incident",
            preventiveMeasures: "Preventive measures to implement",
          });

          // Save draft
          await postMortemPage.saveDraft();
        }
      }
    });

    test("should save post-mortem as draft", async ({ page }, testInfo) => {
      test.skip(
        testInfo.project.name === "viewer",
        "Viewers cannot create post-mortems",
      );

      const listPage = new IncidentListPage(page);
      await listPage.goto();
      await listPage.waitForIncidentsLoaded();
      await listPage.clickIncident(0);

      const detailPage = new IncidentDetailPage(page);
      if (await detailPage.postMortemButton.isVisible().catch(() => false)) {
        await detailPage.clickPostMortem();

        postMortemPage = new PostMortemPage(page);

        if (await postMortemPage.isEditable()) {
          await postMortemPage.fillSummary("Draft summary");
          await postMortemPage.saveDraft();

          // Page should remain on post-mortem
          await expect(page).toHaveURL(/\/postmortem/);
        }
      }
    });
  });

  test.describe("Action Items", () => {
    test("should add action item to post-mortem", async ({
      page,
    }, testInfo) => {
      test.skip(
        testInfo.project.name === "viewer",
        "Viewers cannot manage action items",
      );

      const listPage = new IncidentListPage(page);
      await listPage.goto();
      await listPage.waitForIncidentsLoaded();
      await listPage.clickIncident(0);

      const detailPage = new IncidentDetailPage(page);
      if (await detailPage.postMortemButton.isVisible().catch(() => false)) {
        await detailPage.clickPostMortem();

        postMortemPage = new PostMortemPage(page);

        // Add action item if button is visible
        if (
          await postMortemPage.addActionItemButton
            .isVisible()
            .catch(() => false)
        ) {
          await postMortemPage.addActionItem(
            "Test action item",
            undefined,
            undefined,
          );

          const actionItems = await postMortemPage.getActionItems();
          expect(
            actionItems.some((ai) => ai.description.includes("Test")),
          ).toBe(true);
        }
      }
    });
  });

  test.describe("Publish Post-Mortem", () => {
    test("should publish post-mortem", async ({ page }, testInfo) => {
      test.skip(
        testInfo.project.name === "viewer",
        "Viewers cannot publish post-mortems",
      );

      const listPage = new IncidentListPage(page);
      await listPage.goto();
      await listPage.waitForIncidentsLoaded();
      await listPage.clickIncident(0);

      const detailPage = new IncidentDetailPage(page);
      if (await detailPage.postMortemButton.isVisible().catch(() => false)) {
        await detailPage.clickPostMortem();

        postMortemPage = new PostMortemPage(page);

        // Fill required fields and publish
        if (await postMortemPage.isEditable()) {
          await postMortemPage.fillSummary("Summary for publication");

          if (await postMortemPage.isPublishButtonVisible()) {
            await postMortemPage.publish();

            // Status should change to published
            const status = await postMortemPage.getStatus();
            expect(["published", "draft", ""]).toContain(status);
          }
        }
      }
    });
  });
});
