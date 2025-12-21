const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

type RequestOptions = {
  method?: string;
  headers?: Record<string, string>;
  body?: unknown;
  token?: string;
};

import { Tag, CreateTagRequest, UpdateTagRequest } from '../types/tag';
import { Incident, IncidentListResponse, CreateIncidentRequest, UpdateIncidentRequest, IncidentFilters, User } from '../types/incident';
import { DashboardStats, TrendPeriod, SLAMetrics } from '../types/stats';
import { IncidentActivity, AddCommentRequest } from '../types/activity';
import { Attachment } from '../types/attachment';
import { NotificationSetting } from '../types/notification';
import { IncidentTemplate, CreateTemplateRequest, UpdateTemplateRequest, CreateIncidentFromTemplateRequest } from '../types/template';

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
};

export const userApi = {
  getAll: (token: string) => apiRequest<User[]>('/users', { token }),
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
