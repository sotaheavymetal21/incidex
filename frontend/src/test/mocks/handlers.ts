import { http, HttpResponse } from 'msw';
import { testUsers, testIncidents, testTags } from '../fixtures';

const API_URL = 'http://localhost:8080/api';

export const handlers = [
  // Auth handlers
  http.post(`${API_URL}/auth/login`, async ({ request }) => {
    const body = await request.json() as { email: string; password: string };

    if (body.email === 'test@example.com' && body.password === 'TestPassword123!') {
      return HttpResponse.json({
        access_token: 'mock-access-token',
        user: testUsers.viewer,
      });
    }

    return HttpResponse.json(
      { error: 'Invalid credentials' },
      { status: 401 }
    );
  }),

  http.post(`${API_URL}/auth/register`, async ({ request }) => {
    const body = await request.json() as { email: string; name: string; password: string };

    if (body.email === 'existing@example.com') {
      return HttpResponse.json(
        { error: 'Email already exists' },
        { status: 409 }
      );
    }

    return HttpResponse.json({
      access_token: 'mock-access-token',
      user: { ...testUsers.viewer, email: body.email, name: body.name },
    });
  }),

  http.post(`${API_URL}/auth/refresh`, () => {
    return HttpResponse.json({
      access_token: 'new-mock-access-token',
      user: testUsers.viewer,
    });
  }),

  http.post(`${API_URL}/auth/logout`, () => {
    return HttpResponse.json({ message: 'Logged out successfully' });
  }),

  // User handlers
  http.get(`${API_URL}/users`, () => {
    return HttpResponse.json(Object.values(testUsers));
  }),

  http.get(`${API_URL}/users/:id`, ({ params }) => {
    const id = Number(params.id);
    const user = Object.values(testUsers).find(u => u.id === id);

    if (!user) {
      return HttpResponse.json(
        { error: 'User not found' },
        { status: 404 }
      );
    }

    return HttpResponse.json(user);
  }),

  // Incident handlers
  http.get(`${API_URL}/incidents`, () => {
    return HttpResponse.json({
      incidents: Object.values(testIncidents),
      pagination: {
        page: 1,
        limit: 10,
        total: Object.values(testIncidents).length,
        total_pages: 1,
      },
    });
  }),

  http.get(`${API_URL}/incidents/:id`, ({ params }) => {
    const id = Number(params.id);
    const incident = Object.values(testIncidents).find(i => i.id === id);

    if (!incident) {
      return HttpResponse.json(
        { error: 'Incident not found' },
        { status: 404 }
      );
    }

    return HttpResponse.json(incident);
  }),

  http.post(`${API_URL}/incidents`, async ({ request }) => {
    const body = await request.json() as Record<string, unknown>;

    return HttpResponse.json({
      id: 100,
      ...body,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    }, { status: 201 });
  }),

  http.put(`${API_URL}/incidents/:id`, async ({ params, request }) => {
    const id = Number(params.id);
    const body = await request.json() as Record<string, unknown>;

    return HttpResponse.json({
      id,
      ...body,
      updated_at: new Date().toISOString(),
    });
  }),

  http.delete(`${API_URL}/incidents/:id`, () => {
    return HttpResponse.json({ message: 'Incident deleted' });
  }),

  // Tag handlers
  http.get(`${API_URL}/tags`, () => {
    return HttpResponse.json(Object.values(testTags));
  }),

  // Stats handlers
  http.get(`${API_URL}/stats/dashboard`, () => {
    return HttpResponse.json({
      total_incidents: 42,
      open_incidents: 10,
      critical_incidents: 3,
      sla_violations: 2,
      trends: [],
    });
  }),

  http.get(`${API_URL}/stats/sla`, () => {
    return HttpResponse.json({
      total_incidents: 42,
      resolved_incidents: 32,
      sla_violated_count: 5,
      sla_compliance_rate: 88.1,
      average_mttr: 4.5,
      median_mttr: 3.2,
      currently_overdue: 2,
    });
  }),
];
