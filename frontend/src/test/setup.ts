import '@testing-library/jest-dom/vitest';
import { afterAll, afterEach, beforeAll } from 'vitest';
import { server } from './mocks/server';

// すべてのテスト前にサーバーを起動
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));

// 各テスト後にハンドラーをリセット
afterEach(() => server.resetHandlers());

// すべてのテスト後にサーバーを閉じる
afterAll(() => server.close());

// localStorage のモック
const localStorageMock = {
  getItem: (key: string) => localStorageMock.store[key] || null,
  setItem: (key: string, value: string) => {
    localStorageMock.store[key] = value;
  },
  removeItem: (key: string) => {
    delete localStorageMock.store[key];
  },
  clear: () => {
    localStorageMock.store = {};
  },
  store: {} as Record<string, string>,
};

Object.defineProperty(window, 'localStorage', { value: localStorageMock });

// window.location のモック
const locationMock = {
  href: '',
  origin: 'http://localhost:3000',
  pathname: '/',
  search: '',
  hash: '',
};

Object.defineProperty(window, 'location', {
  value: locationMock,
  writable: true,
});
