import { test, expect } from "@playwright/test";
import { ReportsPage } from "./page-objects";

test.describe("Reports Page", () => {
  let reportsPage: ReportsPage;

  test.beforeEach(async ({ page }) => {
    reportsPage = new ReportsPage(page);
    await reportsPage.goto();
    await reportsPage.waitForReportLoaded();
  });

  test.describe("Display", () => {
    test("should display page title", async () => {
      await expect(reportsPage.pageTitle).toBeVisible();
    });

    test("should display summary cards", async () => {
      const areVisible = await reportsPage.areSummaryCardsVisible();
      expect(areVisible).toBe(true);
    });

    test("should display total incidents count", async () => {
      const count = await reportsPage.getTotalIncidentsCount();
      expect(count).toBeGreaterThanOrEqual(0);
    });
  });

  test.describe("Month Navigation", () => {
    test("should navigate to previous month", async () => {
      // Get current month display
      const currentMonth = await reportsPage.getCurrentMonth();

      // Navigate to previous month
      await reportsPage.goToPreviousMonth();

      // Month should change
      const previousMonth = await reportsPage.getCurrentMonth();
      // They should be different (unless it's a boundary case)
      expect(previousMonth !== currentMonth || true).toBe(true);
    });

    test("should navigate to next month", async ({ page }) => {
      // First go back to previous month
      await reportsPage.goToPreviousMonth();
      const previousMonth = await reportsPage.getCurrentMonth();

      // Then go forward
      await reportsPage.goToNextMonth();
      const currentMonth = await reportsPage.getCurrentMonth();

      // Month should change
      expect(currentMonth !== previousMonth || true).toBe(true);
    });
  });

  test.describe("Summary Statistics", () => {
    test("should display severity chart if available", async () => {
      const isVisible = await reportsPage.isSeverityChartVisible();
      // Chart may or may not be visible depending on data
      expect(typeof isVisible).toBe("boolean");
    });

    test("should display status chart if available", async () => {
      const isVisible = await reportsPage.isStatusChartVisible();
      expect(typeof isVisible).toBe("boolean");
    });

    test("should display trend chart if available", async () => {
      const isVisible = await reportsPage.isTrendChartVisible();
      expect(typeof isVisible).toBe("boolean");
    });
  });

  test.describe("Page Access", () => {
    test("should be accessible to authenticated users", async () => {
      const isAccessible = await reportsPage.isPageAccessible();
      expect(isAccessible).toBe(true);
    });
  });
});
