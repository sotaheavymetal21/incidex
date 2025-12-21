import { User } from './incident';

export type ActivityType =
  | 'created'
  | 'comment'
  | 'status_change'
  | 'severity_change'
  | 'assignee_change'
  | 'resolved'
  | 'reopened';

export interface IncidentActivity {
  id: number;
  incident_id: number;
  user_id: number;
  activity_type: ActivityType;
  comment?: string;
  old_value?: string;
  new_value?: string;
  created_at: string;
  user?: User;
}

export interface AddCommentRequest {
  comment: string;
}
