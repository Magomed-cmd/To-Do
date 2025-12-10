# To-Do Platform — Frontend API Documentation

## Base URLs

| Service | URL | Purpose |
|---------|-----|---------|
| User Service | `http://localhost:8081` | Auth, Profile, Admin |
| Task Service | `http://localhost:8082` | Tasks, Categories, Comments, Export |
| Analytics Service | `http://localhost:8083` | Metrics |

## Authentication

All protected endpoints require JWT in header:
```
Authorization: Bearer <accessToken>
```

Token refresh when accessToken expires — use `/auth/refresh` with refreshToken.

---

# AUTH ENDPOINTS (User Service :8081)

## POST /auth/register
Create new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "name": "John Doe",
  "password": "SecurePass123!"
}
```

**Response 201:**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "role": "user",
    "isActive": true
  },
  "tokens": {
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "accessTokenExpiresAt": "2024-12-10T16:15:00Z",
    "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
    "refreshTokenExpiresAt": "2024-12-17T16:00:00Z"
  }
}
```

**Errors:** 400 (validation), 409 (email exists)

---

## POST /auth/login
Login with email/password.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response 200:** Same as register

**Errors:** 401 (wrong credentials), 403 (user inactive/locked)

---

## POST /auth/refresh
Get new tokens using refresh token.

**Request:**
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response 200:**
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "accessTokenExpiresAt": "2024-12-10T16:30:00Z",
  "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshTokenExpiresAt": "2024-12-17T16:15:00Z"
}
```

**Errors:** 401 (token revoked/expired)

---

## POST /auth/logout
Revoke refresh token.

**Request:**
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response:** 204 No Content

---

# PROFILE ENDPOINTS (User Service :8081)

## GET /users/profile
Get current user profile. **Requires auth.**

**Response 200:**
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "role": "user",
  "isActive": true
}
```

---

## PUT /users/profile
Update current user profile. **Requires auth.**

**Request:**
```json
{
  "name": "John Updated"
}
```

**Response 200:** Updated user object

---

## GET /users/preferences
Get user preferences. **Requires auth.**

**Response 200:**
```json
{
  "notificationsEnabled": true,
  "emailNotifications": true,
  "theme": "dark",
  "language": "ru",
  "timezone": "Europe/Moscow"
}
```

---

## PUT /users/preferences
Update user preferences. **Requires auth.**

**Request:**
```json
{
  "theme": "light",
  "language": "en"
}
```

**Response 200:** Updated preferences

---

# TASK ENDPOINTS (Task Service :8082)

## GET /tasks
Get user's tasks with filters. **Requires auth.**

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| status | string | `pending`, `in_progress`, `completed`, `archived` |
| priority | string | `low`, `medium`, `high` |
| categoryId | int64 | Filter by category |
| search | string | Search in title/description |
| dueFrom | datetime | Due date >= |
| dueTo | datetime | Due date <= |
| limit | int | Default 20, max 100 |
| offset | int | Default 0 |

**Example:** `GET /tasks?status=pending&priority=high&limit=10`

**Response 200:**
```json
[
  {
    "id": 1,
    "userId": 1,
    "title": "Complete project",
    "description": "Finish the frontend",
    "status": "pending",
    "priority": "high",
    "dueDate": "2024-12-15T10:00:00Z",
    "categoryId": 2,
    "category": {
      "id": 2,
      "name": "Work"
    },
    "createdAt": "2024-12-10T09:00:00Z",
    "updatedAt": "2024-12-10T09:00:00Z"
  }
]
```

---

## POST /tasks
Create new task. **Requires auth.**

**Request:**
```json
{
  "title": "Buy groceries",
  "description": "Milk, bread, eggs",
  "status": "pending",
  "priority": "medium",
  "dueDate": "2024-12-11T18:00:00Z",
  "categoryId": 1
}
```

**Required:** title (1-200 chars)
**Optional:** description, status (default: pending), priority (default: medium), dueDate, categoryId

**Response 201:** Created task object

---

## GET /tasks/:id
Get single task by ID. **Requires auth.**

**Response 200:** Task object

**Errors:** 404 (not found or no access)

---

## PUT /tasks/:id
Update task. **Requires auth.**

**Request:**
```json
{
  "title": "Updated title",
  "status": "in_progress",
  "clearDueDate": true
}
```

**Special fields:**
- `clearDueDate: true` — removes due date
- `clearCategory: true` — removes category

**Response 200:** Updated task object

---

## DELETE /tasks/:id
Delete task. **Requires auth.**

**Response:** 204 No Content

---

## PATCH /tasks/:id/status
Quick status update. **Requires auth.**

**Request:**
```json
{
  "status": "completed"
}
```

**Response 200:** Updated task object

---

## GET /tasks/:id/comments
Get task comments. **Requires auth.**

**Response 200:**
```json
[
  {
    "id": 1,
    "taskId": 1,
    "userId": 1,
    "content": "Need to review this",
    "createdAt": "2024-12-10T10:00:00Z"
  }
]
```

---

## POST /tasks/:id/comments
Add comment to task. **Requires auth.**

**Request:**
```json
{
  "content": "This is my comment"
}
```

**Response 201:** Created comment object

---

# CATEGORY ENDPOINTS (Task Service :8082)

## GET /categories
Get user's categories. **Requires auth.**

**Response 200:**
```json
[
  {
    "id": 1,
    "userId": 1,
    "name": "Personal",
    "createdAt": "2024-12-01T00:00:00Z",
    "updatedAt": "2024-12-01T00:00:00Z"
  },
  {
    "id": 2,
    "userId": 1,
    "name": "Work",
    "createdAt": "2024-12-01T00:00:00Z",
    "updatedAt": "2024-12-01T00:00:00Z"
  }
]
```

---

## POST /categories
Create category. **Requires auth.**

**Request:**
```json
{
  "name": "Shopping"
}
```

**Response 201:** Created category

---

## DELETE /categories/:id
Delete category. **Requires auth.**

**Response:** 204 No Content

---

# EXPORT ENDPOINTS (Task Service :8082)

## GET /export/csv
Download all tasks as CSV. **Requires auth.**

**Response:**
- Content-Type: `text/csv; charset=utf-8`
- Content-Disposition: `attachment; filename="tasks_2024-12-10.csv"`

**CSV Columns:** ID, Title, Description, Status, Priority, DueDate, Category, CreatedAt, UpdatedAt

---

## GET /export/ical
Download all tasks as iCalendar. **Requires auth.**

**Response:**
- Content-Type: `text/calendar; charset=utf-8`
- Content-Disposition: `attachment; filename="tasks_2024-12-10.ics"`

Use for import into Apple Calendar, Google Calendar, Outlook.

---

# ANALYTICS ENDPOINTS (Analytics Service :8083)

## GET /metrics/daily/:userId
Get daily task metrics.

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| date | string | Format: YYYY-MM-DD (default: today) |

**Response 200:**
```json
{
  "userId": 1,
  "date": "2024-12-10",
  "createdTasks": 5,
  "completedTasks": 3,
  "totalTasks": 15,
  "updatedAt": "2024-12-10T12:00:00Z"
}
```

---

# ERROR RESPONSE FORMAT

All errors return:
```json
{
  "error": "ERROR_CODE",
  "message": "Human readable message",
  "details": "Optional details"
}
```

**Common Error Codes:**
| Code | HTTP | Description |
|------|------|-------------|
| VALIDATION_FAILED | 400 | Invalid input data |
| UNAUTHORIZED | 401 | Missing/invalid token |
| INVALID_CREDENTIALS | 401 | Wrong email/password |
| TOKEN_EXPIRED | 401 | Access token expired |
| FORBIDDEN | 403 | No permission |
| USER_INACTIVE | 403 | Account disabled |
| NOT_FOUND | 404 | Resource not found |
| TASK_NOT_FOUND | 404 | Task not found |
| USER_ALREADY_EXISTS | 409 | Email taken |
| INTERNAL_ERROR | 500 | Server error |

---

# FRONTEND IMPLEMENTATION NOTES

## Token Storage
- Store `accessToken` in memory (not localStorage)
- Store `refreshToken` in httpOnly cookie or secure storage
- Refresh tokens ~1 min before expiry

## State Management
Recommended stores:
- `authStore` — user, tokens, isAuthenticated
- `taskStore` — tasks[], filters, pagination
- `categoryStore` — categories[]

## Recommended Flow
1. On app load: check for stored refreshToken
2. If exists: call `/auth/refresh`
3. On 401: redirect to login
4. On task mutations: optimistic updates

## Date Handling
All dates in ISO 8601 format: `2024-12-10T15:30:00Z`
Frontend should convert to user's timezone for display.

---

# QUICK REFERENCE

| Action | Method | Endpoint |
|--------|--------|----------|
| Register | POST | /auth/register |
| Login | POST | /auth/login |
| Refresh | POST | /auth/refresh |
| Logout | POST | /auth/logout |
| Get Profile | GET | /users/profile |
| List Tasks | GET | /tasks |
| Create Task | POST | /tasks |
| Update Task | PUT | /tasks/:id |
| Delete Task | DELETE | /tasks/:id |
| Complete Task | PATCH | /tasks/:id/status |
| List Categories | GET | /categories |
| Create Category | POST | /categories |
| Export CSV | GET | /export/csv |
| Export iCal | GET | /export/ical |
