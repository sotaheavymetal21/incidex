const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

type RequestOptions = {
  method?: string;
  headers?: Record<string, string>;
  body?: unknown;
  token?: string;
};

import { Tag, CreateTagRequest, UpdateTagRequest } from '../types/tag';
import { Incident, IncidentListResponse, CreateIncidentRequest, UpdateIncidentRequest, IncidentFilters, User } from '../types/incident';

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
