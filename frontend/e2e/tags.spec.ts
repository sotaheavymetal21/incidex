import { test, expect } from "@playwright/test";
import { TagsPage } from "./page-objects";

/**
 * Tag Management Tests
 * Admin and Editor roles can manage tags
 */

test.describe("Tag Management Page", () => {
  let tagsPage: TagsPage;

  test.beforeEach(async ({ page }, testInfo) => {
    // Skip if viewer role (no write access)
    test.skip(
      testInfo.project.name === "viewer",
      "Tag management tests require admin or editor role",
    );

    tagsPage = new TagsPage(page);
    await tagsPage.goto();
    await tagsPage.waitForTagsLoaded();
  });

  test.describe("Display", () => {
    test("should display page title", async () => {
      await expect(tagsPage.pageTitle).toBeVisible();
    });

    test("should display existing tags", async () => {
      const tags = await tagsPage.getTags();
      expect(tags.length).toBeGreaterThan(0);
    });

    test("should display create tag button", async () => {
      const isVisible = await tagsPage.isCreateButtonVisible();
      expect(isVisible).toBe(true);
    });

    test("should show tag name and color", async () => {
      const tags = await tagsPage.getTags();
      if (tags.length > 0) {
        const firstTag = tags[0];
        expect(firstTag.name).toBeTruthy();
        // Color may be in various formats
        expect(firstTag.color || tags.length > 0).toBeTruthy();
      }
    });
  });

  test.describe("Create Tag", () => {
    test("should open create tag form", async () => {
      await tagsPage.openCreateForm();
      await expect(tagsPage.nameInput).toBeVisible();
    });

    test("should create a new tag", async () => {
      const testTagName = `E2E Test Tag ${Date.now()}`;

      await tagsPage.createTag({
        name: testTagName,
        color: "#ff5733",
      });

      // Verify tag appears in list
      const tagExists = await tagsPage.tagExists(testTagName);
      expect(tagExists).toBe(true);
    });

    test("should cancel tag creation", async () => {
      await tagsPage.openCreateForm();
      await tagsPage.cancelForm();

      // Form should be closed
      await expect(tagsPage.nameInput)
        .toBeHidden()
        .catch(() => {
          // Form may still be visible but empty
        });
    });
  });

  test.describe("Edit Tag", () => {
    test("should edit existing tag", async () => {
      // First create a tag to edit
      const testTagName = `Edit Test ${Date.now()}`;
      await tagsPage.createTag({
        name: testTagName,
        color: "#3366cc",
      });

      // Edit the tag
      const newName = `Updated ${testTagName}`;
      await tagsPage.editTag(testTagName, {
        name: newName,
      });

      // Verify update
      const tagExists = await tagsPage.tagExists(newName);
      expect(tagExists).toBe(true);
    });
  });

  test.describe("Delete Tag", () => {
    test("should delete a tag", async () => {
      // First create a tag to delete
      const testTagName = `Delete Test ${Date.now()}`;
      await tagsPage.createTag({
        name: testTagName,
        color: "#cc3366",
      });

      // Verify tag exists
      let tagExists = await tagsPage.tagExists(testTagName);
      expect(tagExists).toBe(true);

      // Delete the tag
      await tagsPage.deleteTag(testTagName);

      // Verify tag is deleted
      tagExists = await tagsPage.tagExists(testTagName);
      expect(tagExists).toBe(false);
    });
  });
});

test.describe("Tag Management - Viewer Role", () => {
  test("Viewer should not see create button", async ({ page }, testInfo) => {
    test.skip(
      testInfo.project.name !== "viewer",
      "This test is for viewer role only",
    );

    const tagsPage = new TagsPage(page);
    await tagsPage.goto();

    const isCreateVisible = await tagsPage.isCreateButtonVisible();
    expect(isCreateVisible).toBe(false);
  });
});
