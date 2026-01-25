import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import Tabs from '../Tabs';

describe('Tabs', () => {
  const defaultTabs = [
    { id: 'tab1', label: 'Tab 1', content: <div>Content 1</div> },
    { id: 'tab2', label: 'Tab 2', content: <div>Content 2</div> },
    { id: 'tab3', label: 'Tab 3', content: <div>Content 3</div> },
  ];

  describe('rendering', () => {
    it('renders all tab labels', () => {
      render(<Tabs tabs={defaultTabs} />);

      expect(screen.getByText('Tab 1')).toBeInTheDocument();
      expect(screen.getByText('Tab 2')).toBeInTheDocument();
      expect(screen.getByText('Tab 3')).toBeInTheDocument();
    });

    it('renders the first tab content by default', () => {
      render(<Tabs tabs={defaultTabs} />);

      expect(screen.getByText('Content 1')).toBeInTheDocument();
      expect(screen.queryByText('Content 2')).not.toBeInTheDocument();
      expect(screen.queryByText('Content 3')).not.toBeInTheDocument();
    });

    it('renders specified default tab content', () => {
      render(<Tabs tabs={defaultTabs} defaultTab="tab2" />);

      expect(screen.queryByText('Content 1')).not.toBeInTheDocument();
      expect(screen.getByText('Content 2')).toBeInTheDocument();
      expect(screen.queryByText('Content 3')).not.toBeInTheDocument();
    });

    it('renders tabs with icons', () => {
      const tabsWithIcons = [
        {
          id: 'tab1',
          label: 'Tab 1',
          content: <div>Content 1</div>,
          icon: <span data-testid="icon-1">Icon</span>,
        },
      ];

      render(<Tabs tabs={tabsWithIcons} />);

      expect(screen.getByTestId('icon-1')).toBeInTheDocument();
    });
  });

  describe('tab switching', () => {
    it('switches to another tab when clicked', () => {
      render(<Tabs tabs={defaultTabs} />);

      expect(screen.getByText('Content 1')).toBeInTheDocument();

      fireEvent.click(screen.getByText('Tab 2'));

      expect(screen.queryByText('Content 1')).not.toBeInTheDocument();
      expect(screen.getByText('Content 2')).toBeInTheDocument();
    });

    it('calls onChange callback when tab is switched', () => {
      const onChange = vi.fn();
      render(<Tabs tabs={defaultTabs} onChange={onChange} />);

      fireEvent.click(screen.getByText('Tab 2'));

      expect(onChange).toHaveBeenCalledWith('tab2');
    });

    it('does not call onChange when clicking the active tab', () => {
      const onChange = vi.fn();
      render(<Tabs tabs={defaultTabs} onChange={onChange} />);

      // Click the already active tab
      fireEvent.click(screen.getByText('Tab 1'));

      // Still called but with the same tab id
      expect(onChange).toHaveBeenCalledWith('tab1');
    });
  });

  describe('active tab styling', () => {
    it('applies active class to the current tab', () => {
      render(<Tabs tabs={defaultTabs} />);

      const tab1Button = screen.getByText('Tab 1').closest('button');
      const tab2Button = screen.getByText('Tab 2').closest('button');

      expect(tab1Button).toHaveAttribute('aria-current', 'page');
      expect(tab2Button).not.toHaveAttribute('aria-current');
    });

    it('updates active state when switching tabs', () => {
      render(<Tabs tabs={defaultTabs} />);

      fireEvent.click(screen.getByText('Tab 2'));

      const tab1Button = screen.getByText('Tab 1').closest('button');
      const tab2Button = screen.getByText('Tab 2').closest('button');

      expect(tab1Button).not.toHaveAttribute('aria-current');
      expect(tab2Button).toHaveAttribute('aria-current', 'page');
    });
  });

  describe('edge cases', () => {
    it('handles empty tabs array', () => {
      render(<Tabs tabs={[]} />);

      // Should render without crashing
      expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });

    it('handles single tab', () => {
      const singleTab = [{ id: 'only', label: 'Only Tab', content: <div>Only Content</div> }];

      render(<Tabs tabs={singleTab} />);

      expect(screen.getByText('Only Tab')).toBeInTheDocument();
      expect(screen.getByText('Only Content')).toBeInTheDocument();
    });
  });
});
