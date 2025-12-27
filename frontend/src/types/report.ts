export interface MonthlyReport {
  period: ReportPeriod;
  summary: IncidentSummary;
  severity_breakdown: Record<string, number>;
  status_breakdown: Record<string, number>;
  daily_trend: DailyIncidentCount[];
  top_tags: TagStatistic[];
  performance_metrics: PerformanceMetrics;
  comparison?: PeriodComparison;
}

export interface ReportPeriod {
  start_date: string;
  end_date: string;
  month: number;
  year: number;
}

export interface IncidentSummary {
  total_incidents: number;
  new_incidents: number;
  resolved_incidents: number;
  open_incidents: number;
  critical_incidents: number;
}

export interface DailyIncidentCount {
  date: string;
  count: number;
}

export interface TagStatistic {
  tag_id: number;
  tag_name: string;
  count: number;
}

export interface PerformanceMetrics {
  average_resolution_time_hours: number;
}

export interface PeriodComparison {
  previous_period: ReportPeriod;
  total_incidents_change: number;
  total_incidents_change_percent: number;
  resolved_incidents_change: number;
  resolved_incidents_change_percent: number;
}
