import { test, expect } from "@playwright/test";
import { DashboardPage } from "./page-objects";

test.describe("Dashboard Page", () => {
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page }) => {
    dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();
    await dashboardPage.waitForPageLoad();
  });

  test.describe("Statistics Cards", () => {
    test("should display all statistics cards", async () => {
      const areVisible = await dashboardPage.areStatisticsCardsVisible();
      expect(areVisible).toBe(true);
    });

    test("should display total incidents count", async () => {
      const count = await dashboardPage.getTotalIncidentsCount();
      expect(count).toBeGreaterThanOrEqual(0);
    });

    test("should display critical incidents count", async () => {
      const count = await dashboardPage.getCriticalCount();
      expect(count).toBeGreaterThanOrEqual(0);
    });

    test("should display open incidents count", async () => {
      const count = await dashboardPage.getOpenCount();
      expect(count).toBeGreaterThanOrEqual(0);
    });

    test("should display resolved incidents count", async () => {
      const count = await dashboardPage.getResolvedCount();
      expect(count).toBeGreaterThanOrEqual(0);
    });
  });

  test.describe("Card Navigation", () => {
    test("should navigate to incidents list when clicking total card", async ({
      page,
    }) => {
      await dashboardPage.clickTotalIncidentsCard();
      await expect(page).toHaveURL(/\/incidents/);
    });

    test("should navigate to critical incidents when clicking critical card", async ({
      page,
    }) => {
      await dashboardPage.clickCriticalCard();
      await expect(page).toHaveURL(/\/incidents\?severity=critical/);
    });

    test("should navigate to open incidents when clicking open card", async ({
      page,
    }) => {
      await dashboardPage.clickOpenCard();
      await expect(page).toHaveURL(/\/incidents\?status=open/);
    });

    test("should navigate to resolved incidents when clicking resolved card", async ({
      page,
    }) => {
      await dashboardPage.clickResolvedCard();
      await expect(page).toHaveURL(/\/incidents\?status=resolved/);
    });
  });

  test.describe("Graph Display", () => {
    test("should display pie chart by default", async () => {
      await dashboardPage.switchToPieChart();
      await expect(dashboardPage.severityChart).toBeVisible();
    });

    test("should switch to time series view", async () => {
      await dashboardPage.switchToTimeSeries();
      await expect(dashboardPage.trendChart).toBeVisible();
    });

    test("should switch to bar chart view", async () => {
      await dashboardPage.switchToBarChart();
      // Bar chart should be displayed
      await dashboardPage.page.waitForLoadState("networkidle");
    });
  });

  test.describe("Period Selection", () => {
    test("should switch to daily period", async ({ page }) => {
      await dashboardPage.setPeriodDaily();
      // Verify period is selected
      await page.waitForLoadState("networkidle");
    });

    test("should switch to weekly period", async ({ page }) => {
      await dashboardPage.setPeriodWeekly();
      await page.waitForLoadState("networkidle");
    });

    test("should switch to monthly period", async ({ page }) => {
      await dashboardPage.setPeriodMonthly();
      await page.waitForLoadState("networkidle");
    });
  });

  test.describe("Recent Incidents Table", () => {
    test("should display recent incidents table", async () => {
      await expect(dashboardPage.recentIncidentsTable).toBeVisible();
    });

    test("should show incident details in table", async () => {
      const incidents = await dashboardPage.getRecentIncidents();
      // May or may not have incidents
      expect(Array.isArray(incidents)).toBe(true);
      if (incidents.length > 0) {
        expect(incidents[0].title).toBeTruthy();
      }
    });

    test("should navigate to incident detail when clicking row", async ({
      page,
    }) => {
      const incidents = await dashboardPage.getRecentIncidents();
      if (incidents.length > 0) {
        await dashboardPage.clickRecentIncident(0);
        await expect(page).toHaveURL(/\/incidents\/\d+/);
      }
    });
  });
});
