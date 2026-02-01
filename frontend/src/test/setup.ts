import "@testing-library/jest-dom/vitest";
import { afterAll, afterEach, beforeAll, vi } from "vitest";
import { server } from "./mocks/server";

// AbortController のポリフィル（jsdom 互換性のため）
class MockAbortController {
  signal: AbortSignal;
  constructor() {
    this.signal = {
      aborted: false,
      reason: undefined,
      onabort: null,
      throwIfAborted: () => {},
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => true,
    } as unknown as AbortSignal;
  }
  abort() {
    (this.signal as unknown as { aborted: boolean }).aborted = true;
  }
}

// global.AbortController が正しく動作しない場合のフォールバック
if (typeof global.AbortController === "undefined" || process.env.VITEST) {
  global.AbortController =
    MockAbortController as unknown as typeof AbortController;
}

// fetch をラップして signal の互換性問題を回避
// undici (Node.js native fetch) は AbortSignal の instanceof チェックが厳密
const originalFetch = global.fetch;
global.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
  if (init?.signal) {
    // signal を除外して呼び出し
    const { signal, ...restInit } = init;
    return originalFetch(input, restInit);
  }
  return originalFetch(input, init);
};

// すべてのテスト前にサーバーを起動
beforeAll(() => server.listen({ onUnhandledRequest: "error" }));

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

Object.defineProperty(window, "localStorage", { value: localStorageMock });

// window.location のモック
const locationMock = {
  href: "",
  origin: "http://localhost:3000",
  pathname: "/",
  search: "",
  hash: "",
};

Object.defineProperty(window, "location", {
  value: locationMock,
  writable: true,
});
