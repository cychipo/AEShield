# AEShield - Claude Context

<!-- Note: This file is auto-updated. Add manual entries between START/END markers. -->
<!-- START_MANUAL -->

## Current Feature: Job Queue for File Processing (011-job-queue)

### Architecture
- **Storage**: MongoDB new `jobs` collection for job status
- **Progress**: Long-polling every 1-2 seconds (no WebSocket)
- **Track**: Bytes processed counter during AES-GCM streaming
- **Immediate feedback**: Derive key first to fail fast on wrong password

### New Files
- backend/internal/files/job.go - Job queue service
- backend/internal/storage/job_repo.go - MongoDB repository
- frontend/src/components/ProgressJob.jsx - Progress bar component  
- frontend/src/hooks/useJobPolling.js - Polling hook

### Data Model
```go
type Job {
  ID, UserID, Type, Status, Progress int
  FileMetadataID primitive.ObjectID
  Result, Error interface{}
  CreatedAt, UpdatedAt, CompletedAt time.Time
}
```

### API Contracts
- POST /api/v1/jobs - Create job
- GET /api/v1/jobs/:id - Get status
- DELETE /api/v1/jobs/:id - Cancel job  
- GET /api/v1/jobs - List user's jobs

### Phase Status
- [x] Plan complete
- [ ] Implement (tasks.md needed)
- [ ] Test
- [ ] Review

<!-- END_MANUAL -->