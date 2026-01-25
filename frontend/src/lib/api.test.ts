import { describe, it, expect, beforeEach, vi } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '@/test/mocks/server';
import {
  authApi,
  tagApi,
  incidentApi,
  userApi,
  statsApi,
  activityApi,
  postMortemApi,
  actionItemApi,
  auditLogApi,
  reportApi,
  notificationApi,
} from './api';
import { testUsers, testIncidents, testTags } from '@/test/fixtures';

const API_URL = 'http://localhost:8080/api';

// logger をモック
vi.mock('./logger', () => ({
  logger: {
    warn: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
    debug: vi.fn(),
    apiRequest: vi.fn(),
    apiResponse: vi.fn(),
  },
}));

describe('API Client', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  describe('authApi', () => {
    describe('register', () => {
      it('sends registration request with correct data', async () => {
        server.use(
          http.post(`${API_URL}/auth/register`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual({
              name: 'Test User',
              email: 'test@example.com',
              password: 'password123',
              employee_number: 'EMP001',
              department: 'Engineering',
            });
            return HttpResponse.json({
              access_token: 'new-token',
              user: testUsers.viewer,
            });
          })
        );

        const result = await authApi.register(
          'Test User',
          'test@example.com',
          'password123',
          'EMP001',
          'Engineering'
        );

        expect(result.access_token).toBe('new-token');
        expect(result.user).toEqual(testUsers.viewer);
      });

      it('handles registration error', async () => {
        server.use(
          http.post(`${API_URL}/auth/register`, () => {
            return HttpResponse.json(
              { error: 'Email already exists' },
              { status: 409 }
            );
          })
        );

        await expect(
          authApi.register('Test', 'existing@example.com', 'pass', 'EMP001', 'Eng')
        ).rejects.toThrow('Email already exists');
      });
    });

    describe('login', () => {
      it('sends login request and returns token', async () => {
        server.use(
          http.post(`${API_URL}/auth/login`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual({
              email: 'test@example.com',
              password: 'password123',
            });
            return HttpResponse.json({
              access_token: 'test-token',
              user: testUsers.viewer,
            });
          })
        );

        const result = await authApi.login('test@example.com', 'password123');

        expect(result.access_token).toBe('test-token');
        expect(result.user).toEqual(testUsers.viewer);
      });

      it('handles invalid credentials', async () => {
        server.use(
          http.post(`${API_URL}/auth/login`, () => {
            return HttpResponse.json(
              { error: 'Invalid credentials' },
              { status: 401 }
            );
          })
        );

        await expect(
          authApi.login('wrong@example.com', 'wrongpass')
        ).rejects.toThrow('Invalid credentials');
      });
    });

    describe('logout', () => {
      it('sends logout request', async () => {
        server.use(
          http.post(`${API_URL}/auth/logout`, () => {
            return HttpResponse.json({ message: 'Logged out successfully' });
          })
        );

        const result = await authApi.logout();
        expect(result.message).toBe('Logged out successfully');
      });
    });

    describe('refresh', () => {
      it('sends refresh request', async () => {
        server.use(
          http.post(`${API_URL}/auth/refresh`, () => {
            return HttpResponse.json({
              access_token: 'refreshed-token',
              user: testUsers.viewer,
            });
          })
        );

        const result = await authApi.refresh();
        expect(result.access_token).toBe('refreshed-token');
      });
    });

    describe('requestPasswordReset', () => {
      it('sends password reset request', async () => {
        server.use(
          http.post(`${API_URL}/auth/forgot-password`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual({ email: 'test@example.com' });
            return HttpResponse.json({ message: 'Email sent' });
          })
        );

        const result = await authApi.requestPasswordReset('test@example.com');
        expect(result.message).toBe('Email sent');
      });
    });

    describe('resetPassword', () => {
      it('sends reset password request with token', async () => {
        server.use(
          http.post(`${API_URL}/auth/reset-password`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual({
              token: 'reset-token',
              new_password: 'newpassword123',
            });
            return HttpResponse.json({ message: 'Password reset successfully' });
          })
        );

        const result = await authApi.resetPassword('reset-token', 'newpassword123');
        expect(result.message).toBe('Password reset successfully');
      });
    });

    describe('validateResetToken', () => {
      it('validates reset token', async () => {
        server.use(
          http.get(`${API_URL}/auth/validate-reset-token`, ({ request }) => {
            const url = new URL(request.url);
            expect(url.searchParams.get('token')).toBe('valid-token');
            return HttpResponse.json({ valid: true });
          })
        );

        const result = await authApi.validateResetToken('valid-token');
        expect(result.valid).toBe(true);
      });

      it('returns false for invalid token', async () => {
        server.use(
          http.get(`${API_URL}/auth/validate-reset-token`, () => {
            return HttpResponse.json({ valid: false });
          })
        );

        const result = await authApi.validateResetToken('invalid-token');
        expect(result.valid).toBe(false);
      });
    });
  });

  describe('tagApi', () => {
    const token = 'test-token';

    describe('getAll', () => {
      it('fetches all tags with authorization', async () => {
        server.use(
          http.get(`${API_URL}/tags`, ({ request }) => {
            expect(request.headers.get('Authorization')).toBe(`Bearer ${token}`);
            return HttpResponse.json(Object.values(testTags));
          })
        );

        const result = await tagApi.getAll(token);
        expect(result).toEqual(Object.values(testTags));
      });
    });

    describe('create', () => {
      it('creates a new tag', async () => {
        const newTag = { name: 'New Tag', color: '#FF0000' };
        server.use(
          http.post(`${API_URL}/tags`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual(newTag);
            return HttpResponse.json({ id: 10, ...newTag });
          })
        );

        const result = await tagApi.create(token, newTag);
        expect(result.name).toBe('New Tag');
        expect(result.id).toBe(10);
      });
    });

    describe('update', () => {
      it('updates an existing tag', async () => {
        const updateData = { name: 'Updated Tag' };
        server.use(
          http.put(`${API_URL}/tags/1`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual(updateData);
            return HttpResponse.json({ id: 1, name: 'Updated Tag', color: '#000000' });
          })
        );

        const result = await tagApi.update(token, 1, updateData);
        expect(result.name).toBe('Updated Tag');
      });
    });

    describe('delete', () => {
      it('deletes a tag', async () => {
        server.use(
          http.delete(`${API_URL}/tags/1`, ({ request }) => {
            expect(request.headers.get('Authorization')).toBe(`Bearer ${token}`);
            return HttpResponse.json({ message: 'Tag deleted' });
          })
        );

        await tagApi.delete(token, 1);
        // void返却のため、エラーなく完了すれば成功
      });
    });
  });

  describe('incidentApi', () => {
    const token = 'test-token';

    describe('getAll', () => {
      it('fetches incidents without filters', async () => {
        server.use(
          http.get(`${API_URL}/incidents`, () => {
            return HttpResponse.json({
              incidents: Object.values(testIncidents),
              pagination: { page: 1, limit: 10, total: 4, total_pages: 1 },
            });
          })
        );

        const result = await incidentApi.getAll(token);
        expect(result.incidents).toHaveLength(Object.values(testIncidents).length);
      });

      it('fetches incidents with filters', async () => {
        server.use(
          http.get(`${API_URL}/incidents`, ({ request }) => {
            const url = new URL(request.url);
            expect(url.searchParams.get('severity')).toBe('critical');
            expect(url.searchParams.get('status')).toBe('open');
            expect(url.searchParams.get('page')).toBe('2');
            expect(url.searchParams.get('limit')).toBe('20');
            return HttpResponse.json({
              incidents: [],
              pagination: { page: 2, limit: 20, total: 0, total_pages: 0 },
            });
          })
        );

        await incidentApi.getAll(token, {
          severity: 'critical',
          status: 'open',
          page: 2,
          limit: 20,
        });
      });

      it('includes search and sort parameters', async () => {
        server.use(
          http.get(`${API_URL}/incidents`, ({ request }) => {
            const url = new URL(request.url);
            expect(url.searchParams.get('search')).toBe('database');
            expect(url.searchParams.get('sort')).toBe('created_at');
            expect(url.searchParams.get('order')).toBe('desc');
            return HttpResponse.json({
              incidents: [],
              pagination: { page: 1, limit: 10, total: 0, total_pages: 0 },
            });
          })
        );

        await incidentApi.getAll(token, {
          search: 'database',
          sort: 'created_at',
          order: 'desc',
        });
      });
    });

    describe('getById', () => {
      it('fetches a single incident', async () => {
        const incident = testIncidents.critical;
        server.use(
          http.get(`${API_URL}/incidents/${incident.id}`, () => {
            return HttpResponse.json(incident);
          })
        );

        const result = await incidentApi.getById(token, incident.id);
        expect(result.id).toBe(incident.id);
        expect(result.title).toBe(incident.title);
      });

      it('handles not found error', async () => {
        server.use(
          http.get(`${API_URL}/incidents/9999`, () => {
            return HttpResponse.json(
              { error: 'Incident not found' },
              { status: 404 }
            );
          })
        );

        await expect(incidentApi.getById(token, 9999)).rejects.toThrow('Incident not found');
      });
    });

    describe('create', () => {
      it('creates a new incident', async () => {
        const newIncident = {
          title: 'New Incident',
          description: 'Test description',
          severity: 'high' as const,
        };

        server.use(
          http.post(`${API_URL}/incidents`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual(newIncident);
            return HttpResponse.json({
              id: 100,
              ...newIncident,
              status: 'open',
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
            });
          })
        );

        const result = await incidentApi.create(token, newIncident);
        expect(result.title).toBe('New Incident');
        expect(result.id).toBe(100);
      });
    });

    describe('update', () => {
      it('updates an existing incident', async () => {
        const updateData = { title: 'Updated Title', status: 'investigating' as const };

        server.use(
          http.put(`${API_URL}/incidents/1`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual(updateData);
            return HttpResponse.json({
              id: 1,
              ...updateData,
              updated_at: new Date().toISOString(),
            });
          })
        );

        const result = await incidentApi.update(token, 1, updateData);
        expect(result.title).toBe('Updated Title');
      });
    });

    describe('delete', () => {
      it('deletes an incident', async () => {
        server.use(
          http.delete(`${API_URL}/incidents/1`, () => {
            return HttpResponse.json({ message: 'Incident deleted' });
          })
        );

        await incidentApi.delete(token, 1);
        // void返却のため、エラーなく完了すれば成功
      });
    });

    describe('assignIncident', () => {
      it('assigns user to incident', async () => {
        server.use(
          http.post(`${API_URL}/incidents/1/assign`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual({ assignee_id: 5 });
            return HttpResponse.json({
              id: 1,
              assigned_to_id: 5,
            });
          })
        );

        const result = await incidentApi.assignIncident(token, 1, 5);
        expect(result.assigned_to_id).toBe(5);
      });

      it('unassigns user from incident', async () => {
        server.use(
          http.post(`${API_URL}/incidents/1/assign`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual({ assignee_id: null });
            return HttpResponse.json({
              id: 1,
              assigned_to_id: null,
            });
          })
        );

        const result = await incidentApi.assignIncident(token, 1, null);
        expect(result.assigned_to_id).toBeNull();
      });
    });
  });

  describe('userApi', () => {
    const token = 'test-token';

    describe('getAll', () => {
      it('fetches all users', async () => {
        server.use(
          http.get(`${API_URL}/users`, () => {
            return HttpResponse.json(Object.values(testUsers));
          })
        );

        const result = await userApi.getAll(token);
        expect(result).toHaveLength(Object.values(testUsers).length);
      });
    });

    describe('getById', () => {
      it('fetches a single user', async () => {
        server.use(
          http.get(`${API_URL}/users/1`, () => {
            return HttpResponse.json(testUsers.admin);
          })
        );

        const result = await userApi.getById(token, 1);
        expect(result.id).toBe(1);
        expect(result.role).toBe('admin');
      });
    });

    describe('create', () => {
      it('creates a new user', async () => {
        const newUser = {
          name: 'New User',
          email: 'new@example.com',
          password: 'password123',
          role: 'editor' as const,
        };

        server.use(
          http.post(`${API_URL}/users`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual(newUser);
            return HttpResponse.json({
              id: 10,
              ...newUser,
              is_active: true,
            });
          })
        );

        const result = await userApi.create(token, newUser);
        expect(result.email).toBe('new@example.com');
        expect(result.id).toBe(10);
      });
    });

    describe('update', () => {
      it('updates user information', async () => {
        const updateData = { name: 'Updated Name' };

        server.use(
          http.put(`${API_URL}/users/1`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual(updateData);
            return HttpResponse.json({
              ...testUsers.admin,
              name: 'Updated Name',
            });
          })
        );

        const result = await userApi.update(token, 1, updateData);
        expect(result.name).toBe('Updated Name');
      });
    });

    describe('toggleActive', () => {
      it('activates a user', async () => {
        server.use(
          http.patch(`${API_URL}/users/1/status`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual({ is_active: true });
            return HttpResponse.json({ message: 'User activated' });
          })
        );

        const result = await userApi.toggleActive(token, 1, true);
        expect(result.message).toBe('User activated');
      });

      it('deactivates a user', async () => {
        server.use(
          http.patch(`${API_URL}/users/1/status`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual({ is_active: false });
            return HttpResponse.json({ message: 'User deactivated' });
          })
        );

        const result = await userApi.toggleActive(token, 1, false);
        expect(result.message).toBe('User deactivated');
      });
    });

    describe('updatePassword', () => {
      it('updates user password', async () => {
        const passwordData = {
          current_password: 'oldpass',
          new_password: 'newpass123',
        };

        server.use(
          http.put(`${API_URL}/users/1/password`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual(passwordData);
            return HttpResponse.json({ message: 'Password updated' });
          })
        );

        const result = await userApi.updatePassword(token, 1, passwordData);
        expect(result.message).toBe('Password updated');
      });
    });

    describe('adminResetPassword', () => {
      it('resets user password as admin', async () => {
        server.use(
          http.post(`${API_URL}/users/1/admin-reset-password`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual({ new_password: 'newpassword123' });
            return HttpResponse.json({ message: 'Password reset successfully' });
          })
        );

        const result = await userApi.adminResetPassword(token, 1, 'newpassword123');
        expect(result.message).toBe('Password reset successfully');
      });
    });

    describe('delete', () => {
      it('deletes a user', async () => {
        server.use(
          http.delete(`${API_URL}/users/1`, () => {
            return HttpResponse.json({ message: 'User deleted' });
          })
        );

        const result = await userApi.delete(token, 1);
        expect(result.message).toBe('User deleted');
      });
    });
  });

  describe('statsApi', () => {
    const token = 'test-token';

    describe('getDashboardStats', () => {
      it('fetches dashboard stats with default period', async () => {
        server.use(
          http.get(`${API_URL}/stats/dashboard`, ({ request }) => {
            const url = new URL(request.url);
            expect(url.searchParams.get('period')).toBe('daily');
            return HttpResponse.json({
              total_incidents: 100,
              open_incidents: 10,
              critical_incidents: 2,
            });
          })
        );

        const result = await statsApi.getDashboardStats(token);
        expect(result.total_incidents).toBe(100);
      });

      it('fetches dashboard stats with custom period', async () => {
        server.use(
          http.get(`${API_URL}/stats/dashboard`, ({ request }) => {
            const url = new URL(request.url);
            expect(url.searchParams.get('period')).toBe('weekly');
            return HttpResponse.json({
              total_incidents: 100,
              open_incidents: 10,
            });
          })
        );

        await statsApi.getDashboardStats(token, 'weekly');
      });
    });

    describe('getSLAMetrics', () => {
      it('fetches SLA metrics', async () => {
        server.use(
          http.get(`${API_URL}/stats/sla`, () => {
            return HttpResponse.json({
              sla_compliance_rate: 95.5,
              average_mttr: 4.2,
            });
          })
        );

        const result = await statsApi.getSLAMetrics(token);
        expect(result.sla_compliance_rate).toBe(95.5);
      });
    });

    describe('getTagStats', () => {
      it('fetches tag statistics', async () => {
        server.use(
          http.get(`${API_URL}/stats/tags`, () => {
            return HttpResponse.json({
              tag_stats: [
                { tag_id: 1, tag_name: 'Database', incident_count: 10 },
                { tag_id: 2, tag_name: 'Network', incident_count: 5 },
              ],
            });
          })
        );

        const result = await statsApi.getTagStats(token);
        expect(result.tag_stats).toHaveLength(2);
      });
    });
  });

  describe('activityApi', () => {
    const token = 'test-token';

    describe('getActivities', () => {
      it('fetches activities for an incident', async () => {
        server.use(
          http.get(`${API_URL}/incidents/1/activities`, () => {
            return HttpResponse.json([
              { id: 1, type: 'comment', content: 'Test comment' },
            ]);
          })
        );

        const result = await activityApi.getActivities(token, 1);
        expect(result).toHaveLength(1);
      });

      it('fetches activities with limit', async () => {
        server.use(
          http.get(`${API_URL}/incidents/1/activities`, ({ request }) => {
            const url = new URL(request.url);
            expect(url.searchParams.get('limit')).toBe('5');
            return HttpResponse.json([]);
          })
        );

        await activityApi.getActivities(token, 1, 5);
      });
    });

    describe('addComment', () => {
      it('adds a comment to an incident', async () => {
        server.use(
          http.post(`${API_URL}/incidents/1/comments`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual({ content: 'New comment' });
            return HttpResponse.json({ message: 'Comment added' });
          })
        );

        const result = await activityApi.addComment(token, 1, { content: 'New comment' });
        expect(result.message).toBe('Comment added');
      });
    });

    describe('addTimelineEvent', () => {
      it('adds a timeline event', async () => {
        server.use(
          http.post(`${API_URL}/incidents/1/timeline`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual({
              event_type: 'status_change',
              description: 'Status changed',
            });
            return HttpResponse.json({
              id: 1,
              event_type: 'status_change',
            });
          })
        );

        const result = await activityApi.addTimelineEvent(token, 1, {
          event_type: 'status_change',
          description: 'Status changed',
        });
        expect(result.event_type).toBe('status_change');
      });
    });
  });

  describe('postMortemApi', () => {
    const token = 'test-token';

    describe('getAll', () => {
      it('fetches post mortems with filters', async () => {
        server.use(
          http.get(`${API_URL}/post-mortems`, ({ request }) => {
            const url = new URL(request.url);
            expect(url.searchParams.get('status')).toBe('published');
            expect(url.searchParams.get('page')).toBe('1');
            return HttpResponse.json({
              post_mortems: [],
              pagination: { page: 1, limit: 10, total: 0 },
            });
          })
        );

        await postMortemApi.getAll(token, { status: 'published', page: 1 });
      });
    });

    describe('getById', () => {
      it('fetches a single post mortem', async () => {
        server.use(
          http.get(`${API_URL}/post-mortems/1`, () => {
            return HttpResponse.json({
              id: 1,
              title: 'Test Post Mortem',
              status: 'draft',
            });
          })
        );

        const result = await postMortemApi.getById(token, 1);
        expect(result.id).toBe(1);
      });
    });

    describe('create', () => {
      it('creates a new post mortem', async () => {
        const data = {
          incident_id: 1,
          title: 'New Post Mortem',
          summary: 'Summary text',
        };

        server.use(
          http.post(`${API_URL}/post-mortems`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual(data);
            return HttpResponse.json({ id: 1, ...data, status: 'draft' });
          })
        );

        const result = await postMortemApi.create(token, data);
        expect(result.title).toBe('New Post Mortem');
      });
    });

    describe('publish', () => {
      it('publishes a post mortem', async () => {
        server.use(
          http.post(`${API_URL}/post-mortems/1/publish`, () => {
            return HttpResponse.json({ id: 1, status: 'published' });
          })
        );

        const result = await postMortemApi.publish(token, 1);
        expect(result.status).toBe('published');
      });
    });

    describe('unpublish', () => {
      it('unpublishes a post mortem', async () => {
        server.use(
          http.post(`${API_URL}/post-mortems/1/unpublish`, () => {
            return HttpResponse.json({ id: 1, status: 'draft' });
          })
        );

        const result = await postMortemApi.unpublish(token, 1);
        expect(result.status).toBe('draft');
      });
    });
  });

  describe('actionItemApi', () => {
    const token = 'test-token';

    describe('getAll', () => {
      it('fetches action items with filters', async () => {
        server.use(
          http.get(`${API_URL}/action-items`, ({ request }) => {
            const url = new URL(request.url);
            expect(url.searchParams.get('status')).toBe('open');
            expect(url.searchParams.get('priority')).toBe('high');
            return HttpResponse.json({
              action_items: [],
              pagination: { page: 1, limit: 10, total: 0 },
            });
          })
        );

        await actionItemApi.getAll(token, { status: 'open', priority: 'high' });
      });
    });

    describe('getByPostMortemId', () => {
      it('fetches action items for a post mortem', async () => {
        server.use(
          http.get(`${API_URL}/post-mortems/1/action-items`, () => {
            return HttpResponse.json([
              { id: 1, title: 'Action Item 1' },
            ]);
          })
        );

        const result = await actionItemApi.getByPostMortemId(token, 1);
        expect(result).toHaveLength(1);
      });
    });

    describe('create', () => {
      it('creates a new action item', async () => {
        const data = {
          post_mortem_id: 1,
          title: 'New Action Item',
          description: 'Description',
          priority: 'high' as const,
        };

        server.use(
          http.post(`${API_URL}/action-items`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual(data);
            return HttpResponse.json({ id: 1, ...data });
          })
        );

        const result = await actionItemApi.create(token, data);
        expect(result.title).toBe('New Action Item');
      });
    });
  });

  describe('auditLogApi', () => {
    const token = 'test-token';

    describe('getAll', () => {
      it('fetches audit logs with filters', async () => {
        server.use(
          http.get(`${API_URL}/audit-logs`, ({ request }) => {
            const url = new URL(request.url);
            expect(url.searchParams.get('action')).toBe('create');
            expect(url.searchParams.get('resource_type')).toBe('incident');
            return HttpResponse.json({
              audit_logs: [],
              pagination: { page: 1, limit: 10, total: 0 },
            });
          })
        );

        await auditLogApi.getAll(token, {
          action: 'create',
          resource_type: 'incident',
        });
      });
    });

    describe('getById', () => {
      it('fetches a single audit log', async () => {
        server.use(
          http.get(`${API_URL}/audit-logs/1`, () => {
            return HttpResponse.json({
              id: 1,
              action: 'create',
              resource_type: 'incident',
            });
          })
        );

        const result = await auditLogApi.getById(token, 1);
        expect(result.id).toBe(1);
      });
    });
  });

  describe('reportApi', () => {
    const token = 'test-token';

    describe('getMonthlyReport', () => {
      it('fetches monthly report with default date', async () => {
        server.use(
          http.get(`${API_URL}/reports/monthly`, () => {
            return HttpResponse.json({
              total_incidents: 50,
              average_resolution_time: 4.5,
            });
          })
        );

        const result = await reportApi.getMonthlyReport(token);
        expect(result.total_incidents).toBe(50);
      });

      it('fetches monthly report with specific date', async () => {
        server.use(
          http.get(`${API_URL}/reports/monthly`, ({ request }) => {
            const url = new URL(request.url);
            expect(url.searchParams.get('year')).toBe('2025');
            expect(url.searchParams.get('month')).toBe('6');
            return HttpResponse.json({ total_incidents: 30 });
          })
        );

        await reportApi.getMonthlyReport(token, 2025, 6);
      });
    });

    describe('getCustomReport', () => {
      it('fetches custom date range report', async () => {
        server.use(
          http.get(`${API_URL}/reports/custom`, ({ request }) => {
            const url = new URL(request.url);
            expect(url.searchParams.get('start_date')).toBe('2025-01-01');
            expect(url.searchParams.get('end_date')).toBe('2025-06-30');
            return HttpResponse.json({ total_incidents: 100 });
          })
        );

        const result = await reportApi.getCustomReport(token, '2025-01-01', '2025-06-30');
        expect(result.total_incidents).toBe(100);
      });
    });
  });

  describe('notificationApi', () => {
    const token = 'test-token';

    describe('getMySettings', () => {
      it('fetches current user notification settings', async () => {
        server.use(
          http.get(`${API_URL}/notifications/settings`, () => {
            return HttpResponse.json({
              email_enabled: true,
              slack_enabled: false,
            });
          })
        );

        const result = await notificationApi.getMySettings(token);
        expect(result.email_enabled).toBe(true);
      });
    });

    describe('updateMySettings', () => {
      it('updates notification settings', async () => {
        server.use(
          http.put(`${API_URL}/notifications/settings`, async ({ request }) => {
            const body = await request.json() as Record<string, unknown>;
            expect(body).toEqual({ email_enabled: false });
            return HttpResponse.json({
              email_enabled: false,
              slack_enabled: false,
            });
          })
        );

        const result = await notificationApi.updateMySettings(token, { email_enabled: false });
        expect(result.email_enabled).toBe(false);
      });
    });

    describe('getUserSettings', () => {
      it('fetches specific user notification settings', async () => {
        server.use(
          http.get(`${API_URL}/notifications/settings/5`, () => {
            return HttpResponse.json({
              email_enabled: true,
              slack_enabled: true,
            });
          })
        );

        const result = await notificationApi.getUserSettings(token, 5);
        expect(result.slack_enabled).toBe(true);
      });
    });
  });

  describe('Error handling', () => {
    it('handles server errors', async () => {
      server.use(
        http.get(`${API_URL}/tags`, () => {
          return HttpResponse.json(
            { error: 'Internal server error' },
            { status: 500 }
          );
        })
      );

      await expect(tagApi.getAll('token')).rejects.toThrow('Internal server error');
    });

    it('handles generic error without message', async () => {
      server.use(
        http.get(`${API_URL}/tags`, () => {
          return new HttpResponse('', { status: 503 });
        })
      );

      await expect(tagApi.getAll('token')).rejects.toThrow('Request failed with status 503');
    });
  });

  describe('Token refresh', () => {
    it('refreshes token on 401 and retries request', async () => {
      let callCount = 0;

      server.use(
        http.get(`${API_URL}/tags`, ({ request }) => {
          callCount++;
          const auth = request.headers.get('Authorization');

          if (auth === 'Bearer old-token') {
            return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
          }

          if (auth === 'Bearer new-token') {
            return HttpResponse.json([{ id: 1, name: 'Tag 1' }]);
          }

          return HttpResponse.json({ error: 'Unexpected token' }, { status: 401 });
        }),
        http.post(`${API_URL}/auth/refresh`, () => {
          return HttpResponse.json({
            access_token: 'new-token',
            user: testUsers.viewer,
          });
        })
      );

      const result = await tagApi.getAll('old-token');

      expect(result).toHaveLength(1);
      expect(callCount).toBe(2); // 最初の401 + リフレッシュ後の成功
      expect(localStorage.getItem('token')).toBe('new-token');
    });
  });
});
