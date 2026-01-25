import { setupServer } from 'msw/node';
import { handlers } from './handlers';

/**
 * テスト用モックサーバーのセットアップ
 * MSW (Mock Service Worker) を使用して API リクエストをインターセプトします
 */
export const server = setupServer(...handlers);
