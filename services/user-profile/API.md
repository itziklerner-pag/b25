# User Profile Service - API Documentation

## Overview

RESTful API for managing user profiles with authentication, privacy controls, and search capabilities.

**Base URL**: `http://localhost:9100/api/v1`

**Version**: 1.0.0

## Authentication

Most endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### JWT Token Structure

```json
{
  "userId": "string",
  "email": "string",
  "role": "string",
  "iat": 1234567890,
  "exp": 1234567890
}
```

## Endpoints

### Profiles

#### Create Profile

Creates a new user profile.

**Endpoint**: `POST /profiles`

**Authentication**: Required

**Request Body**:
```json
{
  "userId": "string (required, max 255)",
  "name": "string (required, min 1, max 255)",
  "bio": "string (optional, max 5000)",
  "avatarUrl": "string (optional, valid URL)",
  "preferences": {
    "theme": "light | dark | auto",
    "language": "string (2-10 chars)",
    "timezone": "string",
    "notifications": {
      "email": "boolean",
      "push": "boolean",
      "sms": "boolean"
    },
    "emailDigest": "daily | weekly | never"
  },
  "privacySettings": {
    "profileVisibility": "public | friends | private",
    "showEmail": "boolean",
    "showBio": "boolean",
    "showAvatar": "boolean",
    "allowMessaging": "boolean",
    "allowFollowing": "boolean"
  }
}
```

**Response**: `201 Created`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "userId": "string",
    "name": "string",
    "bio": "string | null",
    "avatarUrl": "string | null",
    "preferences": {},
    "privacySettings": {},
    "createdAt": "ISO 8601 date",
    "updatedAt": "ISO 8601 date"
  },
  "meta": {
    "timestamp": "ISO 8601 date",
    "requestId": "string"
  }
}
```

**Errors**:
- `400 Bad Request` - Invalid input data
- `401 Unauthorized` - Missing or invalid token
- `409 Conflict` - Profile already exists

---

#### Get Profile by ID

Retrieves a profile by its ID. Privacy settings apply based on authentication.

**Endpoint**: `GET /profiles/:id`

**Authentication**: Optional (affects privacy filtering)

**Parameters**:
- `id` (path) - Profile UUID

**Response**: `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "userId": "string",
    "name": "string",
    "bio": "string | null",
    "avatarUrl": "string | null",
    "preferences": {},
    "privacySettings": {},
    "createdAt": "ISO 8601 date",
    "updatedAt": "ISO 8601 date"
  },
  "meta": {
    "timestamp": "ISO 8601 date",
    "requestId": "string"
  }
}
```

**Errors**:
- `400 Bad Request` - Invalid UUID format
- `404 Not Found` - Profile not found

---

#### Get Profile by User ID

Retrieves a profile by user ID.

**Endpoint**: `GET /profiles/user/:userId`

**Authentication**: Optional

**Parameters**:
- `userId` (path) - User identifier

**Response**: `200 OK` (same as Get Profile by ID)

**Errors**:
- `404 Not Found` - Profile not found

---

#### Get Current User's Profile

Retrieves the authenticated user's profile with all private data.

**Endpoint**: `GET /profiles/me`

**Authentication**: Required

**Response**: `200 OK` (same as Get Profile by ID)

**Errors**:
- `401 Unauthorized` - Not authenticated
- `404 Not Found` - Profile not found

---

#### List Profiles

Lists all profiles with pagination. Privacy settings apply.

**Endpoint**: `GET /profiles`

**Authentication**: Optional

**Query Parameters**:
- `page` (optional) - Page number (default: 1, min: 1)
- `limit` (optional) - Items per page (default: 20, min: 1, max: 100)

**Response**: `200 OK`
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "userId": "string",
      "name": "string",
      "bio": "string | null",
      "avatarUrl": "string | null",
      "preferences": {},
      "privacySettings": {},
      "createdAt": "ISO 8601 date",
      "updatedAt": "ISO 8601 date"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "totalPages": 5
  },
  "meta": {
    "timestamp": "ISO 8601 date",
    "requestId": "string"
  }
}
```

**Errors**:
- `400 Bad Request` - Invalid query parameters

---

#### Search Profiles

Full-text search across profile names and bios.

**Endpoint**: `GET /profiles/search`

**Authentication**: Optional

**Query Parameters**:
- `q` (required) - Search query (min: 1, max: 255)
- `page` (optional) - Page number (default: 1)
- `limit` (optional) - Items per page (default: 20, max: 100)

**Response**: `200 OK` (same as List Profiles)

**Errors**:
- `400 Bad Request` - Missing or invalid query parameters

---

#### Update Profile

Updates an existing profile. Only the owner can update.

**Endpoint**: `PUT /profiles/:id`

**Authentication**: Required

**Parameters**:
- `id` (path) - Profile UUID

**Request Body**:
```json
{
  "name": "string (optional, min 1, max 255)",
  "bio": "string (optional, max 5000)",
  "avatarUrl": "string (optional, valid URL)",
  "preferences": {},
  "privacySettings": {}
}
```

**Response**: `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "userId": "string",
    "name": "string",
    "bio": "string | null",
    "avatarUrl": "string | null",
    "preferences": {},
    "privacySettings": {},
    "createdAt": "ISO 8601 date",
    "updatedAt": "ISO 8601 date"
  },
  "meta": {
    "timestamp": "ISO 8601 date",
    "requestId": "string"
  }
}
```

**Errors**:
- `400 Bad Request` - Invalid input data
- `401 Unauthorized` - Not authenticated
- `403 Forbidden` - Not the profile owner
- `404 Not Found` - Profile not found

---

#### Update Privacy Settings

Updates only the privacy settings of a profile.

**Endpoint**: `PATCH /profiles/:id/privacy`

**Authentication**: Required

**Parameters**:
- `id` (path) - Profile UUID

**Request Body**:
```json
{
  "profileVisibility": "public | friends | private",
  "showEmail": "boolean",
  "showBio": "boolean",
  "showAvatar": "boolean",
  "allowMessaging": "boolean",
  "allowFollowing": "boolean"
}
```

**Response**: `200 OK` (same as Update Profile)

**Errors**:
- `400 Bad Request` - Invalid privacy settings
- `401 Unauthorized` - Not authenticated
- `403 Forbidden` - Not the profile owner
- `404 Not Found` - Profile not found

---

#### Delete Profile

Deletes a profile. Only the owner can delete.

**Endpoint**: `DELETE /profiles/:id`

**Authentication**: Required

**Parameters**:
- `id` (path) - Profile UUID

**Response**: `200 OK`
```json
{
  "success": true,
  "data": {
    "message": "Profile deleted successfully"
  },
  "meta": {
    "timestamp": "ISO 8601 date",
    "requestId": "string"
  }
}
```

**Errors**:
- `401 Unauthorized` - Not authenticated
- `403 Forbidden` - Not the profile owner
- `404 Not Found` - Profile not found

---

### Health & Monitoring

#### Health Check

Basic health check endpoint.

**Endpoint**: `GET /health`

**Authentication**: None

**Response**: `200 OK`
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "service": "user-profile",
    "version": "1.0.0",
    "timestamp": "ISO 8601 date"
  },
  "meta": {
    "timestamp": "ISO 8601 date"
  }
}
```

---

#### Detailed Health Check

Detailed health check with dependency status.

**Endpoint**: `GET /health/detailed`

**Authentication**: None

**Response**: `200 OK` or `503 Service Unavailable`
```json
{
  "success": true,
  "data": {
    "status": "healthy | degraded",
    "service": "user-profile",
    "version": "1.0.0",
    "timestamp": "ISO 8601 date",
    "checks": {
      "database": {
        "status": "healthy | unhealthy",
        "connections": {
          "totalCount": 10,
          "idleCount": 5,
          "waitingCount": 0
        }
      },
      "memory": {
        "status": "healthy",
        "usage": {
          "rss": "100MB",
          "heapTotal": "50MB",
          "heapUsed": "30MB"
        }
      }
    }
  }
}
```

---

#### Readiness Probe

Kubernetes readiness probe.

**Endpoint**: `GET /health/ready`

**Response**: `200 OK` or `503 Service Unavailable`
```json
{
  "ready": true
}
```

---

#### Liveness Probe

Kubernetes liveness probe.

**Endpoint**: `GET /health/live`

**Response**: `200 OK`
```json
{
  "alive": true
}
```

---

#### Metrics

Prometheus metrics endpoint.

**Endpoint**: `GET /metrics`

**Authentication**: None

**Response**: `200 OK` (Prometheus text format)

## Error Codes

| Code | Description |
|------|-------------|
| VALIDATION_ERROR | Invalid input data |
| UNAUTHORIZED | Missing or invalid authentication |
| FORBIDDEN | Insufficient permissions |
| NOT_FOUND | Resource not found |
| PROFILE_EXISTS | Profile already exists |
| PROFILE_NOT_FOUND | Profile not found |
| RATE_LIMIT_EXCEEDED | Too many requests |
| INTERNAL_SERVER_ERROR | Server error |

## Rate Limiting

API requests are rate limited:
- **Window**: 15 minutes
- **Max Requests**: 100 per window

Rate limit headers included in responses:
```
RateLimit-Limit: 100
RateLimit-Remaining: 95
RateLimit-Reset: 1234567890
```

## Privacy Filtering

When accessing profiles, privacy settings determine what data is visible:

### Public Profiles
All fields visible to everyone

### Friends Profiles
Limited fields visible based on settings

### Private Profiles
Only visible to the owner

### Authenticated vs Unauthenticated

Unauthenticated requests see only public data. Authenticated requests may see more based on relationship to profile owner.

## Examples

### cURL Examples

**Create a profile:**
```bash
curl -X POST http://localhost:9100/api/v1/profiles \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-123",
    "name": "John Doe",
    "bio": "Software developer"
  }'
```

**Get a profile:**
```bash
curl http://localhost:9100/api/v1/profiles/PROFILE_UUID
```

**Search profiles:**
```bash
curl "http://localhost:9100/api/v1/profiles/search?q=john&limit=10"
```

**Update privacy settings:**
```bash
curl -X PATCH http://localhost:9100/api/v1/profiles/PROFILE_UUID/privacy \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "profileVisibility": "private",
    "showEmail": false
  }'
```

## Versioning

The API uses URL versioning (`/api/v1`). Future versions will be accessible via `/api/v2`, etc.

## Support

For issues or questions, please refer to the main README or create an issue in the repository.
