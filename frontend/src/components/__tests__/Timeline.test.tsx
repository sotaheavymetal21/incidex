import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import Timeline from '../Timeline';
import { IncidentActivity } from '@/types/activity';

describe('Timeline', () => {
  const createActivity = (
    overrides: Partial<IncidentActivity> = {}
  ): IncidentActivity => ({
    id: 1,
    incident_id: 1,
    user_id: 1,
    activity_type: 'comment',
    comment: 'Test comment',
    old_value: '',
    new_value: '',
    created_at: '2025-01-20T10:00:00Z',
    user: {
      id: 1,
      email: 'test@example.com',
      name: 'Test User',
      role: 'admin',
      is_active: true,
      created_at: '2025-01-01T00:00:00Z',
      updated_at: '2025-01-01T00:00:00Z',
    },
    ...overrides,
  });

  describe('rendering', () => {
    it('shows empty state when no activities', () => {
      render(<Timeline activities={[]} />);

      expect(screen.getByText('アクティビティがありません')).toBeInTheDocument();
    });

    it('renders activity list', () => {
      const activities = [
        createActivity({ id: 1, comment: 'First comment' }),
        createActivity({ id: 2, comment: 'Second comment' }),
      ];

      render(<Timeline activities={activities} />);

      expect(screen.getByText('First comment')).toBeInTheDocument();
      expect(screen.getByText('Second comment')).toBeInTheDocument();
    });

    it('displays user name', () => {
      const activities = [createActivity()];

      render(<Timeline activities={activities} />);

      expect(screen.getByText('Test User')).toBeInTheDocument();
    });

    it('displays formatted date', () => {
      const activities = [createActivity({ created_at: '2025-01-20T10:30:45Z' })];

      render(<Timeline activities={activities} />);

      // Japanese date format
      expect(screen.getByText(/2025/)).toBeInTheDocument();
    });
  });

  describe('activity types', () => {
    it('renders comment activity', () => {
      const activities = [
        createActivity({
          activity_type: 'comment',
          comment: 'This is a comment',
        }),
      ];

      render(<Timeline activities={activities} />);

      expect(screen.getByText('This is a comment')).toBeInTheDocument();
    });

    it('renders created activity', () => {
      const activities = [
        createActivity({
          activity_type: 'created',
          comment: '',
        }),
      ];

      render(<Timeline activities={activities} />);

      expect(screen.getByText(/インシデントを作成しました/)).toBeInTheDocument();
    });

    it('renders status_change activity', () => {
      const activities = [
        createActivity({
          activity_type: 'status_change',
          old_value: 'open',
          new_value: 'investigating',
        }),
      ];

      render(<Timeline activities={activities} />);

      expect(screen.getByText(/ステータスを open から investigating に変更しました/)).toBeInTheDocument();
    });

    it('renders severity_change activity', () => {
      const activities = [
        createActivity({
          activity_type: 'severity_change',
          old_value: 'medium',
          new_value: 'critical',
        }),
      ];

      render(<Timeline activities={activities} />);

      expect(screen.getByText(/重要度を medium から critical に変更しました/)).toBeInTheDocument();
    });

    it('renders assignee_change activity', () => {
      const activities = [
        createActivity({
          activity_type: 'assignee_change',
          old_value: 'User A',
          new_value: 'User B',
        }),
      ];

      render(<Timeline activities={activities} />);

      expect(screen.getByText(/担当者を User A から User B に変更しました/)).toBeInTheDocument();
    });

    it('renders resolved activity', () => {
      const activities = [
        createActivity({
          activity_type: 'resolved',
        }),
      ];

      render(<Timeline activities={activities} />);

      expect(screen.getByText(/インシデントを解決しました/)).toBeInTheDocument();
    });

    it('renders reopened activity', () => {
      const activities = [
        createActivity({
          activity_type: 'reopened',
        }),
      ];

      render(<Timeline activities={activities} />);

      expect(screen.getByText(/インシデントを再オープンしました/)).toBeInTheDocument();
    });

    it('renders detected activity', () => {
      const activities = [
        createActivity({
          activity_type: 'detected',
        }),
      ];

      render(<Timeline activities={activities} />);

      expect(screen.getByText(/インシデントを検知しました/)).toBeInTheDocument();
    });

    it('renders investigation_started activity', () => {
      const activities = [
        createActivity({
          activity_type: 'investigation_started',
        }),
      ];

      render(<Timeline activities={activities} />);

      expect(screen.getByText(/調査を開始しました/)).toBeInTheDocument();
    });

    it('renders root_cause_identified activity', () => {
      const activities = [
        createActivity({
          activity_type: 'root_cause_identified',
        }),
      ];

      render(<Timeline activities={activities} />);

      expect(screen.getByText(/根本原因を特定しました/)).toBeInTheDocument();
    });

    it('renders mitigation activity', () => {
      const activities = [
        createActivity({
          activity_type: 'mitigation',
        }),
      ];

      render(<Timeline activities={activities} />);

      expect(screen.getByText(/緩和策を実施しました/)).toBeInTheDocument();
    });

    it('renders other activity with comment', () => {
      const activities = [
        createActivity({
          activity_type: 'other',
          comment: 'Other activity description',
        }),
      ];

      render(<Timeline activities={activities} />);

      expect(screen.getByText('Other activity description')).toBeInTheDocument();
    });
  });

  describe('timeline styling', () => {
    it('renders connecting lines between activities', () => {
      const activities = [
        createActivity({ id: 1 }),
        createActivity({ id: 2 }),
        createActivity({ id: 3 }),
      ];

      render(<Timeline activities={activities} />);

      // There should be connecting lines between activities (not on the last one)
      const lines = document.querySelectorAll('.bg-gray-200');
      expect(lines.length).toBe(2); // 3 activities = 2 connecting lines
    });

    it('renders activity icons', () => {
      const activities = [createActivity({ activity_type: 'comment' })];

      render(<Timeline activities={activities} />);

      // Check for SVG icon presence
      const svg = document.querySelector('svg');
      expect(svg).toBeInTheDocument();
    });
  });

  describe('unknown user handling', () => {
    it('displays Unknown User when user is missing', () => {
      const activities = [
        createActivity({
          user: undefined,
        }),
      ];

      render(<Timeline activities={activities} />);

      expect(screen.getByText('Unknown User')).toBeInTheDocument();
    });
  });

  describe('comment display', () => {
    it('preserves whitespace in comments', () => {
      const activities = [
        createActivity({
          activity_type: 'comment',
          comment: 'Line 1\nLine 2\nLine 3',
        }),
      ];

      render(<Timeline activities={activities} />);

      const commentElement = screen.getByText(/Line 1/);
      expect(commentElement).toHaveClass('whitespace-pre-wrap');
    });
  });
});
