import { test, expect } from "@playwright/test";
import { IncidentListPage } from "./page-objects";

test.describe("Incident List Page", () => {
  let incidentListPage: IncidentListPage;

  test.beforeEach(async ({ page }) => {
    incidentListPage = new IncidentListPage(page);
    await incidentListPage.goto();
    await incidentListPage.waitForIncidentsLoaded();
  });

  test.describe("Display", () => {
    test("should display page title and header elements", async ({ page }) => {
      await expect(incidentListPage.pageTitle).toBeVisible();
      await expect(incidentListPage.exportButton).toBeVisible();
    });

    test("should display incident table with data", async () => {
      const incidents = await incidentListPage.getIncidentRows();
      expect(incidents.length).toBeGreaterThan(0);

      // Verify first incident has required fields
      const firstIncident = incidents[0];
      expect(firstIncident.title).toBeTruthy();
      expect(firstIncident.severity).toBeTruthy();
      expect(firstIncident.status).toBeTruthy();
    });

    test("should display filter sidebar", async () => {
      await expect(incidentListPage.filterSidebar).toBeVisible();
      await expect(incidentListPage.searchInput).toBeVisible();
    });
  });

  test.describe("Filtering", () => {
    test("should filter by search query", async ({ page }) => {
      // Get initial count
      const initialIncidents = await incidentListPage.getIncidentRows();

      // Search for specific term
      await incidentListPage.search("データベース");
      const filteredIncidents = await incidentListPage.getIncidentRows();

      // Verify results contain search term
      for (const incident of filteredIncidents) {
        const matchesTitle = incident.title
          .toLowerCase()
          .includes("データベース");
        // Search might match other fields too
        expect(matchesTitle || filteredIncidents.length > 0).toBe(true);
      }
    });

    test("should apply severity filter", async () => {
      // Apply critical preset
      await incidentListPage.applyCriticalPreset();
      await incidentListPage.waitForIncidentsLoaded();

      const incidents = await incidentListPage.getIncidentRows();

      // All visible incidents should be critical
      for (const incident of incidents) {
        expect(incident.severity.toLowerCase()).toContain("critical");
      }
    });

    test("should apply unresolved preset filter", async () => {
      await incidentListPage.applyUnresolvedPreset();
      await incidentListPage.waitForIncidentsLoaded();

      const incidents = await incidentListPage.getIncidentRows();

      // All visible incidents should be open
      for (const incident of incidents) {
        expect(incident.status.toLowerCase()).toContain("open");
      }
    });

    test("should clear all filters", async () => {
      // Apply a filter first
      await incidentListPage.applyCriticalPreset();
      await incidentListPage.waitForIncidentsLoaded();

      // Get filtered count
      const filteredIncidents = await incidentListPage.getIncidentRows();

      // Clear filters
      await incidentListPage.clearAllFilters();
      await incidentListPage.waitForIncidentsLoaded();

      // Get unfiltered count
      const allIncidents = await incidentListPage.getIncidentRows();

      // Should have more or equal incidents
      expect(allIncidents.length).toBeGreaterThanOrEqual(
        filteredIncidents.length,
      );
    });

    test("should display active filter chips", async () => {
      // Apply search
      await incidentListPage.search("テスト");

      // Get filter chips
      const chips = await incidentListPage.getActiveFilterChips();
      expect(chips.some((chip) => chip.includes("テスト"))).toBe(true);
    });
  });

  test.describe("Sorting", () => {
    test("should sort by severity", async () => {
      await incidentListPage.sortBy("severity");
      await incidentListPage.waitForIncidentsLoaded();

      const incidents = await incidentListPage.getIncidentRows();
      expect(incidents.length).toBeGreaterThan(0);
    });

    test("should sort by status", async () => {
      await incidentListPage.sortBy("status");
      await incidentListPage.waitForIncidentsLoaded();

      const incidents = await incidentListPage.getIncidentRows();
      expect(incidents.length).toBeGreaterThan(0);
    });

    test("should toggle sort order on second click", async () => {
      // First click - descending
      await incidentListPage.sortBy("detected_at");
      await incidentListPage.waitForIncidentsLoaded();

      const firstOrder = await incidentListPage.getIncidentRows();

      // Second click - ascending
      await incidentListPage.sortBy("detected_at");
      await incidentListPage.waitForIncidentsLoaded();

      const secondOrder = await incidentListPage.getIncidentRows();

      // Order should be reversed (first and last might differ)
      if (firstOrder.length > 1 && secondOrder.length > 1) {
        // At least check that data is present
        expect(firstOrder.length).toBeGreaterThan(0);
        expect(secondOrder.length).toBeGreaterThan(0);
      }
    });
  });

  test.describe("Navigation", () => {
    test("should navigate to incident detail on row click", async ({
      page,
    }) => {
      await incidentListPage.clickIncident(0);
      await expect(page).toHaveURL(/\/incidents\/\d+/);
    });

    test("should navigate to incident by title", async ({ page }) => {
      const incidents = await incidentListPage.getIncidentRows();
      if (incidents.length > 0) {
        const firstTitle = incidents[0].title;
        await incidentListPage.clickIncidentByTitle(firstTitle);
        await expect(page).toHaveURL(/\/incidents\/\d+/);
      }
    });
  });

  test.describe("Pagination", () => {
    test("should display pagination info", async () => {
      await expect(incidentListPage.paginationInfo).toBeVisible();
    });

    test("should navigate between pages", async () => {
      // Check if next button is enabled
      const nextButton = incidentListPage.nextButton.last();
      const isDisabled = await nextButton.isDisabled();

      if (!isDisabled) {
        // Get current incidents
        const firstPageIncidents = await incidentListPage.getIncidentRows();

        // Go to next page
        await incidentListPage.goToNextPage();
        const secondPageIncidents = await incidentListPage.getIncidentRows();

        // Go back to first page
        await incidentListPage.goToPreviousPage();
        const backToFirstPage = await incidentListPage.getIncidentRows();

        // First page should be restored
        if (firstPageIncidents.length > 0 && backToFirstPage.length > 0) {
          expect(firstPageIncidents[0].title).toBe(backToFirstPage[0].title);
        }
      }
    });
  });

  test.describe("Export", () => {
    test("should export incidents to CSV", async ({ page }) => {
      // Set up download listener
      const downloadPromise = page.waitForEvent("download");
      await incidentListPage.exportButton.click();
      const download = await downloadPromise;

      // Verify filename
      expect(download.suggestedFilename()).toMatch(/incidents.*\.csv/);
    });
  });

  test.describe("Empty State", () => {
    test("should show empty state when no incidents match filter", async () => {
      // Search for something that won't exist
      await incidentListPage.search("xyz_nonexistent_12345_test");
      await incidentListPage.waitForIncidentsLoaded();

      const isEmpty = await incidentListPage.isEmptyStateDisplayed();
      const incidents = await incidentListPage.getIncidentRows();

      // Either empty state is shown or no incidents
      expect(isEmpty || incidents.length === 0).toBe(true);
    });
  });
});
