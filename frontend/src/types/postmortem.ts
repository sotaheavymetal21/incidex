import { User } from './incident';

export type PMStatus = 'draft' | 'published';

export interface FiveWhysAnalysis {
  why1: string;
  why2: string;
  why3: string;
  why4: string;
  why5: string;
}

export interface PostMortem {
  id: number;
  incident_id: number;
  author_id: number;
  root_cause: string;
  impact_analysis: string;
  what_went_well: string;
  what_went_wrong: string;
  lessons_learned: string;
  five_whys_analysis: string; // JSON string
  ai_root_cause_suggestion: string;
  status: PMStatus;
  created_at: string;
  updated_at: string;
  published_at: string | null;
  author?: User;
}

export interface CreatePostMortemRequest {
  incident_id: number;
  root_cause?: string;
  impact_analysis?: string;
  what_went_well?: string;
  what_went_wrong?: string;
  lessons_learned?: string;
  five_whys_analysis?: FiveWhysAnalysis;
}

export interface UpdatePostMortemRequest {
  root_cause?: string;
  impact_analysis?: string;
  what_went_well?: string;
  what_went_wrong?: string;
  lessons_learned?: string;
  five_whys_analysis?: FiveWhysAnalysis;
}
