/**
 * Profile Controller Tests
 */
import request from 'supertest';
import { Application } from 'express';
import createApp from '../app';
import db from '../db';
import jwt from 'jsonwebtoken';
import config from '../config';

let app: Application;
let testToken: string;
let testUserId: string;
let testProfileId: string;

beforeAll(async () => {
  app = createApp();

  // Generate test JWT token
  testUserId = 'test-user-123';
  testToken = jwt.sign(
    { userId: testUserId, email: 'test@example.com' },
    config.auth.jwtSecret,
    { expiresIn: '1h' }
  );

  // Connect to test database
  try {
    await db.connect();
  } catch (error) {
    console.warn('Database connection failed, tests may fail');
  }
});

afterAll(async () => {
  // Clean up test data
  if (testProfileId) {
    try {
      await db.query('DELETE FROM user_profiles WHERE id = $1', [testProfileId]);
    } catch (error) {
      // Ignore cleanup errors
    }
  }

  await db.close();
});

describe('Profile API', () => {
  describe('POST /api/v1/profiles', () => {
    it('should create a new profile with authentication', async () => {
      const profileData = {
        userId: testUserId,
        name: 'Test User',
        bio: 'Test bio',
        preferences: {
          theme: 'dark',
          language: 'en',
        },
      };

      const response = await request(app)
        .post('/api/v1/profiles')
        .set('Authorization', `Bearer ${testToken}`)
        .send(profileData)
        .expect(201);

      expect(response.body.success).toBe(true);
      expect(response.body.data).toHaveProperty('id');
      expect(response.body.data.name).toBe(profileData.name);
      expect(response.body.data.bio).toBe(profileData.bio);

      testProfileId = response.body.data.id;
    });

    it('should reject profile creation without authentication', async () => {
      const profileData = {
        userId: 'test-user-456',
        name: 'Test User',
      };

      await request(app)
        .post('/api/v1/profiles')
        .send(profileData)
        .expect(401);
    });

    it('should reject invalid profile data', async () => {
      const profileData = {
        userId: 'test-user-456',
        // Missing required 'name' field
      };

      await request(app)
        .post('/api/v1/profiles')
        .set('Authorization', `Bearer ${testToken}`)
        .send(profileData)
        .expect(400);
    });
  });

  describe('GET /api/v1/profiles/:id', () => {
    it('should get profile by ID', async () => {
      if (!testProfileId) {
        return; // Skip if profile wasn't created
      }

      const response = await request(app)
        .get(`/api/v1/profiles/${testProfileId}`)
        .expect(200);

      expect(response.body.success).toBe(true);
      expect(response.body.data.id).toBe(testProfileId);
    });

    it('should return 404 for non-existent profile', async () => {
      const fakeId = '00000000-0000-0000-0000-000000000000';

      await request(app)
        .get(`/api/v1/profiles/${fakeId}`)
        .expect(404);
    });
  });

  describe('PUT /api/v1/profiles/:id', () => {
    it('should update profile with authentication', async () => {
      if (!testProfileId) {
        return; // Skip if profile wasn't created
      }

      const updateData = {
        name: 'Updated Name',
        bio: 'Updated bio',
      };

      const response = await request(app)
        .put(`/api/v1/profiles/${testProfileId}`)
        .set('Authorization', `Bearer ${testToken}`)
        .send(updateData)
        .expect(200);

      expect(response.body.success).toBe(true);
      expect(response.body.data.name).toBe(updateData.name);
      expect(response.body.data.bio).toBe(updateData.bio);
    });

    it('should reject update without authentication', async () => {
      if (!testProfileId) {
        return; // Skip if profile wasn't created
      }

      await request(app)
        .put(`/api/v1/profiles/${testProfileId}`)
        .send({ name: 'Updated Name' })
        .expect(401);
    });
  });

  describe('GET /api/v1/profiles', () => {
    it('should list profiles with pagination', async () => {
      const response = await request(app)
        .get('/api/v1/profiles')
        .query({ page: 1, limit: 10 })
        .expect(200);

      expect(response.body.success).toBe(true);
      expect(response.body.data).toBeInstanceOf(Array);
      expect(response.body.pagination).toHaveProperty('page');
      expect(response.body.pagination).toHaveProperty('limit');
      expect(response.body.pagination).toHaveProperty('total');
    });
  });

  describe('PATCH /api/v1/profiles/:id/privacy', () => {
    it('should update privacy settings', async () => {
      if (!testProfileId) {
        return; // Skip if profile wasn't created
      }

      const privacyData = {
        profileVisibility: 'private',
        showEmail: false,
      };

      const response = await request(app)
        .patch(`/api/v1/profiles/${testProfileId}/privacy`)
        .set('Authorization', `Bearer ${testToken}`)
        .send(privacyData)
        .expect(200);

      expect(response.body.success).toBe(true);
      expect(response.body.data.privacySettings.profileVisibility).toBe('private');
    });
  });

  describe('GET /api/v1/profiles/search', () => {
    it('should search profiles', async () => {
      const response = await request(app)
        .get('/api/v1/profiles/search')
        .query({ q: 'Test', page: 1, limit: 10 })
        .expect(200);

      expect(response.body.success).toBe(true);
      expect(response.body.data).toBeInstanceOf(Array);
    });

    it('should require search query', async () => {
      await request(app)
        .get('/api/v1/profiles/search')
        .expect(400);
    });
  });
});

describe('Health Checks', () => {
  describe('GET /health', () => {
    it('should return health status', async () => {
      const response = await request(app)
        .get('/health')
        .expect(200);

      expect(response.body.success).toBe(true);
      expect(response.body.data).toHaveProperty('status');
    });
  });

  describe('GET /health/ready', () => {
    it('should return readiness status', async () => {
      await request(app)
        .get('/health/ready')
        .expect((res) => {
          expect([200, 503]).toContain(res.status);
        });
    });
  });
});

describe('Metrics', () => {
  describe('GET /metrics', () => {
    it('should return Prometheus metrics', async () => {
      const response = await request(app)
        .get('/metrics')
        .expect(200);

      expect(response.text).toContain('http_requests_total');
    });
  });
});
