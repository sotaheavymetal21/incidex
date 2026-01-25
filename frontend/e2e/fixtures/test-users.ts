/**
 * E2E Test User Definitions
 * Default password for all seeded users: admin1234 (TEST_USER_PASSWORD env var)
 */
export const testUsers = {
  admin: {
    email: 'admin@example.com',
    password: process.env.TEST_USER_PASSWORD || 'admin1234',
    role: 'admin' as const,
    name: '管理者ユーザー',
  },
  editor: {
    email: 'editor1@example.com',
    password: process.env.TEST_USER_PASSWORD || 'admin1234',
    role: 'editor' as const,
    name: '編集者 太郎',
  },
  viewer: {
    email: 'viewer1@example.com',
    password: process.env.TEST_USER_PASSWORD || 'admin1234',
    role: 'viewer' as const,
    name: '閲覧者 一郎',
  },
} as const;

export type TestUserRole = keyof typeof testUsers;
export type TestUser = (typeof testUsers)[TestUserRole];
