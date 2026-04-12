# Quickstart: Job Queue Feature

## How to Test

### 1. Upload a File (Encrypt Job)

```bash
# Upload file - creates job and returns immediately
curl -X POST http://localhost:6868/api/v1/jobs \
  -H "Authorization: Bearer <token>" \
  -F "file=@document.pdf" \
  -F "password=secret123"

# Response
{
  "job_id": "job-abc123",
  "status": "processing",
  "progress": 0
}
```

### 2. Poll for Job Status

```bash
# Poll every 1-2 seconds
curl http://localhost:6868/api/v1/jobs/job-abc123 \
  -H "Authorization: Bearer <token>"

# Response while processing
{
  "job_id": "job-abc123",
  "type": "encrypt",
  "status": "processing",
  "progress": 45
}

# Response when completed
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

### 3. Decrypt a File (Decrypt Job)

```bash
# Similar flow - send encrypted file for decryption
curl -X POST http://localhost:6868/api/v1/jobs \
  -H "Authorization: Bearer <token>" \
  -F "file_id=file-xyz789" \
  -F "password=secret123"

# Poll for progress
curl http://localhost:6868/api/v1/jobs/job-def456 \
  -H "Authorization: Bearer <token>"
```

## Expected Behavior

| Step | Expected Result |
|------|------------------|
| Upload file | Job created, returns job_id immediately |
| Poll 1 | Progress ~10-30% |
| Poll 2 | Progress ~50-80% |
| Poll 3 | Progress ~100%, status = completed |
| Error | status = failed, error message shown |

## Verification Checklist

- [X] Upload starts and returns job_id within 1 second
- [X] Progress bar shows incrementing percentage
- [X] Completed status returns file metadata
- [X] Failed status returns error details
- [X] Wrong password returns error immediately (not after full decryption)
- [X] Multiple concurrent uploads work correctly
- [X] Frontend shows progress bar, not spinner