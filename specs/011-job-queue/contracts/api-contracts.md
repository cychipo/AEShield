# API Contracts: Job Queue

## Endpoints

### 1. Create Job
**POST** `/api/v1/jobs`

Create a new job for file encryption or decryption.

**Request (multipart/form-data for encrypt)**:
```
file: binary
password: string
encryption_type: string (optional, default: AES-256)
access_mode: string (optional, default: private)
```

**Response (202 Accepted)**:
```json
{
  "job_id": "job-abc123",
  "status": "processing",
  "progress": 0
}
```

### 2. Get Job Status
**GET** `/api/v1/jobs/:id`

Poll for job status and progress.

**Response (200 OK)**:
```json
{
  "job_id": "job-abc123",
  "type": "encrypt",
  "status": "processing",
  "progress": 45,
  "file_id": "file-xyz789"
}
```

**Response (200 Completed)**:
```json
{
  "job_id": "job-abc123",
  "type": "encrypt",
  "status": "completed",
  "progress": 100,
  "result": {
    "file_id": "file-xyz789",
    "filename": "document.pdf",
    "download_url": "https://..."
  }
}
```

**Response (200 Failed)**:
```json
{
  "job_id": "job-abc123",
  "type": "encrypt",
  "status": "failed",
  "progress": 30,
  "error": {
    "code": "timeout",
    "message": "File upload exceeded time limit"
  }
}
```

### 3. Cancel Job
**DELETE** `/api/v1/jobs/:id`

Cancel a running job.

**Response (200 OK)**:
```json
{
  "job_id": "job-abc123",
  "status": "cancelled"
}
```

### 4. List User Jobs
**GET** `/api/v1/jobs`

List all jobs for the current user.

**Query Parameters**:
- `status`: Filter by status (optional)
- `limit`: Number of results (optional, default: 20)
- `offset`: Pagination offset (optional)

**Response (200 OK)**:
```json
{
  "jobs": [
    {
      "job_id": "job-abc123",
      "type": "encrypt",
      "status": "completed",
      "progress": 100,
      "created_at": "2026-04-12T10:00:00Z"
    }
  ],
  "total": 10,
  "limit": 20,
  "offset": 0
}
```

## Polling Strategy

Frontend polls every 1-2 seconds:

```
GET /api/v1/jobs/:id → { progress: 45, status: "processing" }
GET /api/v1/jobs/:id → { progress: 100, status: "completed", result: {...} }
```

Stop polling when status is:
- `completed` - success
- `failed` - error (show error message)
- `cancelled` - user cancelled

## Authentication

All endpoints require JWT Bearer token in header:
```
Authorization: Bearer <token>
```

## Error Handling

| Status Code | Description |
|-------------|-------------|
| 400 | Invalid request (missing fields, invalid type) |
| 401 | Unauthorized (missing or invalid token) |
| 404 | Job not found or doesn't belong to user |
| 422 | Job validation error (e.g., invalid file) |
| 500 | Server error |