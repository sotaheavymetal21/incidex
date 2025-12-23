export type AuditAction = 'create' | 'read' | 'update' | 'delete' | 'login' | 'logout';

export interface AuditLog {
  id: number;
  user_id?: number;
  user_name: string;
  user_email: string;
  action: AuditAction;
  resource_type: string;
  resource_id?: number;
  method: string;
  path: string;
  ip_address: string;
  user_agent: string;
  status_code: number;
  details: string;
  created_at: string;
}

export interface AuditLogFilters {
  page?: number;
  limit?: number;
  user_id?: number;
  action?: AuditAction;
  resource_type?: string;
  start_date?: string;
  end_date?: string;
}

export interface AuditLogResponse {
  logs: AuditLog[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
}
