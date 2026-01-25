import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import SeverityGuide from '../SeverityGuide';

describe('SeverityGuide', () => {
  describe('initial state', () => {
    it('renders collapsed by default', () => {
      render(<SeverityGuide />);

      // Header should be visible
      expect(screen.getByText('Severity（深刻度）の設定基準')).toBeInTheDocument();

      // Content should be collapsed (severity levels not visible)
      expect(screen.queryByText('Critical（致命的）')).not.toBeInTheDocument();
    });

    it('displays toggle button', () => {
      render(<SeverityGuide />);

      const button = screen.getByRole('button');
      expect(button).toBeInTheDocument();
    });
  });

  describe('expand/collapse', () => {
    it('expands when clicked', () => {
      render(<SeverityGuide />);

      const button = screen.getByRole('button');
      fireEvent.click(button);

      // All severity levels should be visible
      expect(screen.getByText('Critical（致命的）')).toBeInTheDocument();
      expect(screen.getByText('High（高）')).toBeInTheDocument();
      expect(screen.getByText('Medium（中）')).toBeInTheDocument();
      expect(screen.getByText('Low（低）')).toBeInTheDocument();
    });

    it('collapses when clicked again', () => {
      render(<SeverityGuide />);

      const button = screen.getByRole('button');

      // Expand
      fireEvent.click(button);
      expect(screen.getByText('Critical（致命的）')).toBeInTheDocument();

      // Collapse
      fireEvent.click(button);
      expect(screen.queryByText('Critical（致命的）')).not.toBeInTheDocument();
    });
  });

  describe('severity content', () => {
    beforeEach(() => {
      render(<SeverityGuide />);
      // Expand the guide
      fireEvent.click(screen.getByRole('button'));
    });

    it('displays response times for each severity', () => {
      expect(screen.getByText(/即時対応（15分以内）/)).toBeInTheDocument();
      expect(screen.getByText(/1時間以内/)).toBeInTheDocument();
      expect(screen.getByText(/4時間以内/)).toBeInTheDocument();
      expect(screen.getByText(/1営業日以内/)).toBeInTheDocument();
    });

    it('displays criteria for critical severity', () => {
      expect(screen.getByText(/サービス全体が停止している/)).toBeInTheDocument();
      expect(screen.getByText(/データ損失や重大なセキュリティ侵害/)).toBeInTheDocument();
    });

    it('displays examples for critical severity', () => {
      expect(screen.getByText('データベース接続の完全な喪失')).toBeInTheDocument();
      expect(screen.getByText('決済システムの障害')).toBeInTheDocument();
    });

    it('displays criteria for high severity', () => {
      expect(screen.getByText(/主要機能が正常に動作していない/)).toBeInTheDocument();
      expect(screen.getByText(/パフォーマンスが著しく低下している/)).toBeInTheDocument();
    });

    it('displays examples for high severity', () => {
      expect(screen.getByText('API応答時間の大幅な遅延')).toBeInTheDocument();
      expect(screen.getByText('モバイルアプリの頻繁なクラッシュ')).toBeInTheDocument();
    });

    it('displays criteria for medium severity', () => {
      expect(screen.getByText(/一部の機能に問題がある/)).toBeInTheDocument();
      expect(screen.getByText(/代替手段が存在する/)).toBeInTheDocument();
    });

    it('displays examples for medium severity', () => {
      expect(screen.getByText('画像アップロード機能の間欠的なエラー')).toBeInTheDocument();
      expect(screen.getByText('メール通知の遅延')).toBeInTheDocument();
    });

    it('displays criteria for low severity', () => {
      expect(screen.getByText(/マイナーな問題や改善要望/)).toBeInTheDocument();
      expect(screen.getByText(/通常業務には支障がない/)).toBeInTheDocument();
    });

    it('displays examples for low severity', () => {
      expect(screen.getByText('UI表示の軽微な崩れ')).toBeInTheDocument();
      expect(screen.getByText('タイムゾーン表示のズレ')).toBeInTheDocument();
    });
  });

  describe('tips section', () => {
    it('displays tips when expanded', () => {
      render(<SeverityGuide />);
      fireEvent.click(screen.getByRole('button'));

      expect(screen.getByText('Tips')).toBeInTheDocument();
      expect(screen.getByText(/深刻度は状況に応じて変更できます/)).toBeInTheDocument();
      expect(screen.getByText(/影響範囲（ユーザー数）とビジネスへの影響度/)).toBeInTheDocument();
      expect(screen.getByText(/セキュリティに関連するインシデントは/)).toBeInTheDocument();
    });
  });

  describe('guidance text', () => {
    it('displays introductory guidance', () => {
      render(<SeverityGuide />);
      fireEvent.click(screen.getByRole('button'));

      expect(
        screen.getByText(/インシデントの深刻度は、以下の基準に基づいて設定してください/)
      ).toBeInTheDocument();
    });
  });

  describe('section headers', () => {
    it('displays criteria and example headers', () => {
      render(<SeverityGuide />);
      fireEvent.click(screen.getByRole('button'));

      // Each severity level should have these headers
      const criteriaHeaders = screen.getAllByText('設定基準:');
      const exampleHeaders = screen.getAllByText('具体例:');

      expect(criteriaHeaders.length).toBe(4); // 4 severity levels
      expect(exampleHeaders.length).toBe(4);
    });
  });
});
