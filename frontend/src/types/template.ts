import { Tag } from './tag';
import { Severity } from './incident';
import { User } from './incident';

export interface IncidentTemplate {
  id: number;
  name: string;
  description: string;
  title: string;
  content: string;
  severity: Severity;
  impact_scope: string;
  creator_id: number;
  is_public: boolean;
  usage_count: number;
  created_at: string;
  updated_at: string;
  creator?: User;
  tags: Tag[];
}

export interface CreateTemplateRequest {
  name: string;
  description: string;
  title: string;
  content: string;
  severity: Severity;
  impact_scope: string;
  is_public: boolean;
  tag_ids: number[];
}

export interface UpdateTemplateRequest {
  name: string;
  description: string;
  title: string;
  content: string;
  severity: Severity;
  impact_scope: string;
  is_public: boolean;
  tag_ids: number[];
}

export interface CreateIncidentFromTemplateRequest {
  template_id: number;
  assignee_id?: number;
  detected_at?: string;
}
