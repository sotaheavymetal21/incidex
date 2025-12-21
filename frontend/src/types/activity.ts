import { User } from './incident';

export type ActivityType =
  | 'created'
  | 'comment'
  | 'status_change'
  | 'severity_change'
  | 'assignee_change'
  | 'resolved'
  | 'reopened'
  | 'detected'
  | 'investigation_started'
  | 'root_cause_identified'
  | 'mitigation'
  | 'timeline_resolved'
  | 'other';

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

export interface AddTimelineEventRequest {
  event_type: 'detected' | 'investigation_started' | 'root_cause_identified' | 'mitigation' | 'timeline_resolved' | 'other';
  event_time: string;
  description: string;
}
