import { defineConfig, devices } from '@playwright/test';
import path from 'path';

const STORAGE_STATE_DIR = path.join(__dirname, 'e2e', 'storage-states');

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [
    ['html', { open: 'never' }],
    ['list'],
  ],
  use: {
    baseURL: process.env.BASE_URL || 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'on-first-retry',
    actionTimeout: 10000,
    navigationTimeout: 15000,
  },

  // Global timeout for each test
  timeout: 60000,

  // Expect timeout
  expect: {
    timeout: 10000,
  },

  projects: [
    // Authentication setup - runs first, MUST run serially to avoid rate limiting
    {
      name: 'auth-setup',
      testMatch: /auth\.setup\.ts/,
      fullyParallel: false, // Run serially to avoid rate limiting
      timeout: 120000, // 2 minutes for auth setup including delays
      use: { ...devices['Desktop Chrome'] },
    },

    // Admin role tests
    {
      name: 'admin',
      testMatch: /\.spec\.ts$/,
      testIgnore: [/auth\.setup\.ts/, /auth\.spec\.ts/], // Exclude auth tests - they run unauthenticated
      use: {
        ...devices['Desktop Chrome'],
        storageState: path.join(STORAGE_STATE_DIR, 'admin.json'),
      },
      dependencies: ['auth-setup'],
    },

    // Editor role tests
    {
      name: 'editor',
      testMatch: /\.spec\.ts$/,
      testIgnore: [/auth\.setup\.ts/, /auth\.spec\.ts/], // Exclude auth tests - they run unauthenticated
      use: {
        ...devices['Desktop Chrome'],
        storageState: path.join(STORAGE_STATE_DIR, 'editor.json'),
      },
      dependencies: ['auth-setup'],
    },

    // Viewer role tests
    {
      name: 'viewer',
      testMatch: /\.spec\.ts$/,
      testIgnore: [/auth\.setup\.ts/, /auth\.spec\.ts/], // Exclude auth tests - they run unauthenticated
      use: {
        ...devices['Desktop Chrome'],
        storageState: path.join(STORAGE_STATE_DIR, 'viewer.json'),
      },
      dependencies: ['auth-setup'],
    },

    // Unauthenticated tests (e.g., login page tests)
    // Depends on auth-setup to avoid concurrent login attempts and rate limiting
    {
      name: 'unauthenticated',
      testMatch: /auth\.spec\.ts$/,
      use: { ...devices['Desktop Chrome'] },
      dependencies: ['auth-setup'],
    },
  ],

  /* Run your local dev server before starting the tests */
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
    timeout: 120 * 1000,
  },
});
