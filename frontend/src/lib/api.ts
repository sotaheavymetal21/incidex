const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

type RequestOptions = {
  method?: string;
  headers?: Record<string, string>;
  body?: unknown;
  token?: string;
};

import { Tag, CreateTagRequest, UpdateTagRequest } from '../types/tag';
import { Incident, IncidentListResponse, CreateIncidentRequest, UpdateIncidentRequest, IncidentFilters, User as IncidentUser } from '../types/incident';
import { DashboardStats, TrendPeriod, SLAMetrics } from '../types/stats';
import { IncidentActivity, AddCommentRequest, AddTimelineEventRequest } from '../types/activity';
import { Attachment } from '../types/attachment';
import { NotificationSetting } from '../types/notification';
import { IncidentTemplate, CreateTemplateRequest, UpdateTemplateRequest, CreateIncidentFromTemplateRequest } from '../types/template';
import { PostMortem, CreatePostMortemRequest, UpdatePostMortemRequest } from '../types/postmortem';
import { ActionItem, CreateActionItemRequest, UpdateActionItemRequest } from '../types/actionitem';
import { User, UpdateUserRequest, UpdatePasswordRequest } from '../types/user';
import { AuditLog, AuditLogFilters, AuditLogResponse } from '../types/auditLog';
import { MonthlyReport } from '../types/report';

async function apiRequest<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
  const url = `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api'}${endpoint}`;

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  if (options.token) {
    headers['Authorization'] = `Bearer ${options.token}`;
  }

  const config: RequestInit = {
    method: options.method || 'GET',
    headers,
  };

  if (options.body) {
    config.body = JSON.stringify(options.body);
  }

  const response = await fetch(url, config);

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `Request failed with status ${response.status}`);
  }

  return response.json();
}

export const authApi = {
  // ... existing auth methods ...
  register: (name: string, email: string, password: string) => 
    apiRequest<{ token: string; user: any }>('/auth/register', {
      method: 'POST',
      body: { name, email, password },
    }),
  login: (email: string, password: string) =>
    apiRequest<{ token: string; user: any }>('/auth/login', {
      method: 'POST',
      body: { email, password },
    }),
};

export const tagApi = {
  getAll: (token: string) => apiRequest<Tag[]>('/tags', { token }),
  create: (token: string, data: CreateTagRequest) =>
    apiRequest<Tag>('/tags', {
      method: 'POST',
      body: data,
      token
    }),
  update: (token: string, id: number, data: UpdateTagRequest) =>
    apiRequest<Tag>(`/tags/${id}`, {
      method: 'PUT',
      body: data,
      token
    }),
  delete: (token: string, id: number) =>
    apiRequest<void>(`/tags/${id}`, {
      method: 'DELETE',
      token
    }),
};

export const incidentApi = {
  getAll: (token: string, params?: IncidentFilters) => {
    const queryParams = new URLSearchParams();
    if (params) {
      if (params.page) queryParams.append('page', params.page.toString());
      if (params.limit) queryParams.append('limit', params.limit.toString());
      if (params.severity) queryParams.append('severity', params.severity);
      if (params.status) queryParams.append('status', params.status);
      if (params.tag_ids) queryParams.append('tag_ids', params.tag_ids);
      if (params.search) queryParams.append('search', params.search);
      if (params.sort) queryParams.append('sort', params.sort);
      if (params.order) queryParams.append('order', params.order);
      if (params.assigned_to_id) queryParams.append('assigned_to_id', params.assigned_to_id.toString());
    }
    const queryString = queryParams.toString();
    return apiRequest<IncidentListResponse>(`/incidents${queryString ? `?${queryString}` : ''}`, { token });
  },
  getById: (token: string, id: number) =>
    apiRequest<Incident>(`/incidents/${id}`, { token }),
  create: (token: string, data: CreateIncidentRequest) =>
    apiRequest<Incident>('/incidents', {
      method: 'POST',
      body: data,
      token
    }),
  update: (token: string, id: number, data: UpdateIncidentRequest) =>
    apiRequest<Incident>(`/incidents/${id}`, {
      method: 'PUT',
      body: data,
      token
    }),
  delete: (token: string, id: number) =>
    apiRequest<void>(`/incidents/${id}`, {
      method: 'DELETE',
      token
    }),
  regenerateSummary: (token: string, id: number) =>
    apiRequest<{ summary: string; generated_at: string }>(`/incidents/${id}/summarize`, {
      method: 'POST',
      token
    }),
  assignIncident: (token: string, id: number, assigneeId: number | null) =>
    apiRequest<Incident>(`/incidents/${id}/assign`, {
      method: 'POST',
      token,
      body: { assignee_id: assigneeId },
    }),
};

export const userApi = {
  getAll: (token: string) => apiRequest<User[]>('/users', { token }),
  getById: (token: string, id: number) => apiRequest<User>(`/users/${id}`, { token }),
  update: (token: string, id: number, data: UpdateUserRequest) =>
    apiRequest<User>(`/users/${id}`, {
      method: 'PUT',
      token,
      body: data,
    }),
  updatePassword: (token: string, id: number, data: UpdatePasswordRequest) =>
    apiRequest<{ message: string }>(`/users/${id}/password`, {
      method: 'PUT',
      token,
      body: data,
    }),
  delete: (token: string, id: number) =>
    apiRequest<{ message: string }>(`/users/${id}`, {
      method: 'DELETE',
      token,
    }),
};

export const statsApi = {
  getDashboardStats: (token: string, period: TrendPeriod = 'daily') =>
    apiRequest<DashboardStats>(`/stats/dashboard?period=${period}`, { token }),

  getSLAMetrics: (token: string) =>
    apiRequest<SLAMetrics>('/stats/sla', { token }),
};

export const exportApi = {
  exportIncidentsCSV: async (token: string, params?: IncidentFilters): Promise<Blob> => {
    const queryParams = new URLSearchParams();
    if (params) {
      if (params.severity) queryParams.append('severity', params.severity);
      if (params.status) queryParams.append('status', params.status);
      if (params.tag_ids) queryParams.append('tag_ids', params.tag_ids);
      if (params.search) queryParams.append('search', params.search);
    }
    const queryString = queryParams.toString();
    const url = `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api'}/export/incidents${queryString ? `?${queryString}` : ''}`;

    const headers: Record<string, string> = {
      'Authorization': `Bearer ${token}`,
    };

    const response = await fetch(url, { headers });

    if (!response.ok) {
      throw new Error(`Export failed with status ${response.status}`);
    }

    return response.blob();
  },
};

export const activityApi = {
  getActivities: (token: string, incidentId: number, limit?: number) => {
    const queryParams = new URLSearchParams();
    if (limit) queryParams.append('limit', limit.toString());
    const queryString = queryParams.toString();
    return apiRequest<IncidentActivity[]>(
      `/incidents/${incidentId}/activities${queryString ? `?${queryString}` : ''}`,
      { token }
    );
  },
  addComment: (token: string, incidentId: number, data: AddCommentRequest) =>
    apiRequest<{ message: string }>(`/incidents/${incidentId}/comments`, {
      method: 'POST',
      body: data,
      token,
    }),
  addTimelineEvent: (token: string, incidentId: number, data: AddTimelineEventRequest) =>
    apiRequest<IncidentActivity>(`/incidents/${incidentId}/timeline`, {
      method: 'POST',
      body: data,
      token,
    }),
};

export const attachmentApi = {
  getAttachments: (token: string, incidentId: number) =>
    apiRequest<Attachment[]>(`/incidents/${incidentId}/attachments`, { token }),

  uploadAttachment: async (token: string, incidentId: number, file: File): Promise<Attachment> => {
    const url = `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api'}/incidents/${incidentId}/attachments`;

    const formData = new FormData();
    formData.append('file', file);

    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
      body: formData,
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `Upload failed with status ${response.status}`);
    }

    return response.json();
  },

  downloadAttachment: async (token: string, incidentId: number, attachmentId: number, fileName: string): Promise<void> => {
    const url = `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api'}/incidents/${incidentId}/attachments/${attachmentId}`;

    const response = await fetch(url, {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Download failed with status ${response.status}`);
    }

    const blob = await response.blob();
    const downloadUrl = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.download = fileName;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(downloadUrl);
  },

  deleteAttachment: (token: string, incidentId: number, attachmentId: number) =>
    apiRequest<{ message: string }>(`/incidents/${incidentId}/attachments/${attachmentId}`, {
      method: 'DELETE',
      token,
    }),
};

export const notificationApi = {
  getMySettings: (token: string) =>
    apiRequest<NotificationSetting>('/notifications/settings', {
      token,
    }),

  updateMySettings: (token: string, settings: Partial<NotificationSetting>) =>
    apiRequest<NotificationSetting>('/notifications/settings', {
      method: 'PUT',
      token,
      body: settings,
    }),

  getUserSettings: (token: string, userId: number) =>
    apiRequest<NotificationSetting>(`/notifications/settings/${userId}`, {
      token,
    }),
};

export const templateApi = {
  getAll: (token: string) =>
    apiRequest<IncidentTemplate[]>('/templates', { token }),

  getById: (token: string, id: number) =>
    apiRequest<IncidentTemplate>(`/templates/${id}`, { token }),

  create: (token: string, data: CreateTemplateRequest) =>
    apiRequest<IncidentTemplate>('/templates', {
      method: 'POST',
      token,
      body: data,
    }),

  update: (token: string, id: number, data: UpdateTemplateRequest) =>
    apiRequest<IncidentTemplate>(`/templates/${id}`, {
      method: 'PUT',
      token,
      body: data,
    }),

  delete: (token: string, id: number) =>
    apiRequest<{ message: string }>(`/templates/${id}`, {
      method: 'DELETE',
      token,
    }),

  createIncidentFromTemplate: (token: string, data: CreateIncidentFromTemplateRequest) =>
    apiRequest<Incident>('/templates/create-incident', {
      method: 'POST',
      token,
      body: data,
    }),
};

export const postMortemApi = {
  getAll: (token: string, params?: {
    status?: string;
    author_id?: number;
    page?: number;
    limit?: number;
  }) => {
    const queryParams = new URLSearchParams();
    if (params) {
      if (params.status) queryParams.append('status', params.status);
      if (params.author_id) queryParams.append('author_id', params.author_id.toString());
      if (params.page) queryParams.append('page', params.page.toString());
      if (params.limit) queryParams.append('limit', params.limit.toString());
    }
    const queryString = queryParams.toString();
    return apiRequest<{ post_mortems: PostMortem[]; pagination: any }>(
      `/post-mortems${queryString ? `?${queryString}` : ''}`,
      { token }
    );
  },

  getById: (token: string, id: number) =>
    apiRequest<PostMortem>(`/post-mortems/${id}`, { token }),

  getByIncidentId: (token: string, incidentId: number) =>
    apiRequest<PostMortem>(`/incidents/${incidentId}/postmortem`, { token }),

  create: (token: string, data: CreatePostMortemRequest) =>
    apiRequest<PostMortem>('/post-mortems', {
      method: 'POST',
      token,
      body: data,
    }),

  update: (token: string, id: number, data: UpdatePostMortemRequest) =>
    apiRequest<PostMortem>(`/post-mortems/${id}`, {
      method: 'PUT',
      token,
      body: data,
    }),

  publish: (token: string, id: number) =>
    apiRequest<PostMortem>(`/post-mortems/${id}/publish`, {
      method: 'POST',
      token,
    }),

  delete: (token: string, id: number) =>
    apiRequest<{ message: string }>(`/post-mortems/${id}`, {
      method: 'DELETE',
      token,
    }),

  generateAISuggestion: (token: string, incidentId: number) =>
    apiRequest<{ suggestion: string }>(`/incidents/${incidentId}/postmortem/ai-suggestion`, {
      method: 'POST',
      token,
    }),
};

export const actionItemApi = {
  getAll: (token: string, params?: {
    status?: string;
    priority?: string;
    assignee_id?: number;
    page?: number;
    limit?: number;
  }) => {
    const queryParams = new URLSearchParams();
    if (params) {
      if (params.status) queryParams.append('status', params.status);
      if (params.priority) queryParams.append('priority', params.priority);
      if (params.assignee_id) queryParams.append('assignee_id', params.assignee_id.toString());
      if (params.page) queryParams.append('page', params.page.toString());
      if (params.limit) queryParams.append('limit', params.limit.toString());
    }
    const queryString = queryParams.toString();
    return apiRequest<{ action_items: ActionItem[]; pagination: any }>(
      `/action-items${queryString ? `?${queryString}` : ''}`,
      { token }
    );
  },

  getById: (token: string, id: number) =>
    apiRequest<ActionItem>(`/action-items/${id}`, { token }),

  getByPostMortemId: (token: string, postMortemId: number) =>
    apiRequest<ActionItem[]>(`/post-mortems/${postMortemId}/action-items`, { token }),

  create: (token: string, data: CreateActionItemRequest) =>
    apiRequest<ActionItem>('/action-items', {
      method: 'POST',
      token,
      body: data,
    }),

  update: (token: string, id: number, data: UpdateActionItemRequest) =>
    apiRequest<ActionItem>(`/action-items/${id}`, {
      method: 'PUT',
      token,
      body: data,
    }),

  delete: (token: string, id: number) =>
    apiRequest<{ message: string }>(`/action-items/${id}`, {
      method: 'DELETE',
      token,
    }),
};

export const auditLogApi = {
  getAll: (token: string, filters?: AuditLogFilters) => {
    const queryParams = new URLSearchParams();
    if (filters) {
      if (filters.page) queryParams.append('page', filters.page.toString());
      if (filters.limit) queryParams.append('limit', filters.limit.toString());
      if (filters.user_id) queryParams.append('user_id', filters.user_id.toString());
      if (filters.action) queryParams.append('action', filters.action);
      if (filters.resource_type) queryParams.append('resource_type', filters.resource_type);
      if (filters.start_date) queryParams.append('start_date', filters.start_date);
      if (filters.end_date) queryParams.append('end_date', filters.end_date);
    }
    const queryString = queryParams.toString();
    return apiRequest<AuditLogResponse>(
      `/audit-logs${queryString ? `?${queryString}` : ''}`,
      { token }
    );
  },

  getById: (token: string, id: number) =>
    apiRequest<AuditLog>(`/audit-logs/${id}`, { token }),
};

export const reportApi = {
  getMonthlyReport: (token: string, year?: number, month?: number) => {
    const queryParams = new URLSearchParams();
    if (year) queryParams.append('year', year.toString());
    if (month) queryParams.append('month', month.toString());
    const queryString = queryParams.toString();
    return apiRequest<MonthlyReport>(
      `/reports/monthly${queryString ? `?${queryString}` : ''}`,
      { token }
    );
  },

  getCustomReport: (token: string, startDate: string, endDate: string) => {
    const queryParams = new URLSearchParams();
    queryParams.append('start_date', startDate);
    queryParams.append('end_date', endDate);
    return apiRequest<MonthlyReport>(
      `/reports/custom?${queryParams.toString()}`,
      { token }
    );
  },
};
