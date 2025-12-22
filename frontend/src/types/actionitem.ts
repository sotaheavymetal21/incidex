import { User } from './incident';

export type Priority = 'high' | 'medium' | 'low';
export type ActionStatus = 'pending' | 'in_progress' | 'completed';

export interface ActionItem {
  id: number;
  post_mortem_id: number;
  title: string;
  description: string;
  assignee_id: number | null;
  priority: Priority;
  status: ActionStatus;
  due_date: string | null;
  related_links: string;
  created_at: string;
  updated_at: string;
  completed_at: string | null;
  assignee?: User;
}

export interface CreateActionItemRequest {
  post_mortem_id: number;
  title: string;
  description?: string;
  assignee_id?: number;
  priority: Priority;
  due_date?: string; // RFC3339 format
  related_links?: string;
}

export interface UpdateActionItemRequest {
  title: string;
  description?: string;
  assignee_id?: number;
  priority: Priority;
  status: ActionStatus;
  due_date?: string;
  related_links?: string;
}
