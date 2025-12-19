import { Tag } from './tag';

export type Severity = 'critical' | 'high' | 'medium' | 'low';
export type Status = 'open' | 'investigating' | 'resolved' | 'closed';

export interface User {
  id: number;
  name: string;
  email: string;
  role: string;
}

export interface Incident {
  id: number;
  title: string;
  description: string;
  summary: string | null;
  severity: Severity;
  status: Status;
  impact_scope: string;
  detected_at: string;
  resolved_at: string | null;
  assignee_id: number | null;
  creator_id: number;
  assignee: User | null;
  creator: User;
  tags: Tag[];
  created_at: string;
  updated_at: string;
}

export interface CreateIncidentRequest {
  title: string;
  description: string;
  severity: Severity;
  status: Status;
  impact_scope: string;
  detected_at: string;
  assignee_id?: number;
  tag_ids: number[];
}

export interface UpdateIncidentRequest {
  title: string;
  description: string;
  severity: Severity;
  status: Status;
  impact_scope: string;
  detected_at: string;
  resolved_at?: string | null;
  assignee_id?: number;
  tag_ids: number[];
}

export interface PaginationResult {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

export interface IncidentListResponse {
  incidents: Incident[];
  pagination: PaginationResult;
}

export interface IncidentFilters {
  page?: number;
  limit?: number;
  severity?: Severity;
  status?: Status;
  tag_ids?: string;
  search?: string;
  sort?: string;
  order?: string;
}
