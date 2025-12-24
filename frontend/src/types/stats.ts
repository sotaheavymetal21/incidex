import { Incident } from './incident';

export interface DashboardStats {
  total_incidents: number;
  by_severity: Record<string, number>;
  by_status: Record<string, number>;
  recent_incidents: Incident[];
  trend_data: TrendDataPoint[];
}

export interface TrendDataPoint {
  date: string;
  count: number;
}

export type TrendPeriod = 'daily' | 'weekly' | 'monthly';

export interface SLAMetrics {
  total_incidents: number;
  resolved_incidents: number;
  sla_violated_count: number;
  sla_compliance_rate: number;
  average_mttr: number;
  median_mttr: number;
  currently_overdue: number;
}

export interface TagStats {
  tag_id: number;
  tag_name: string;
  tag_color: string;
  count: number;
  percentage: number;
}
