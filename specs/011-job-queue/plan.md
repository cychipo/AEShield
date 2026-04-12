# Implementation Plan: Job Queue for File Encryption and Decryption

**Branch**: `011-job-queue` | **Date**: 2026-04-12 | **Spec**: specs/011-job-queue/spec.md
**Input**: Feature specification from `/specs/011-job-queue/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Implement a job queue system for file upload/encryption and decryption operations. When users upload or decrypt files, the system creates a job with a unique ID and tracks progress. The frontend displays a real-time progress bar instead of a loading spinner. Users can poll for job status and receive immediate error feedback for wrong passwords.

## Technical Context

**Language/Version**: Go 1.25 (backend), React 18.2 + Vite 5.2 (frontend)  
**Primary Dependencies**: Fiber v2 (backend), Ant Design 6.x (frontend), MongoDB (existing), Cloudflare R2 (existing)  
**Storage**: MongoDB (already has files and user storage - can reuse for job status)  
**Testing**: go test (backend), no frontend testing framework yet  
**Target Platform**: Linux server (VPS deployment)  
**Project Type**: Web application with backend API + React SPA  
**Performance Goals**: Handle 10+ concurrent jobs, progress updates every 1-2 seconds  
**Constraints**: Existing MongoDB, no new external dependencies preferred, maintain timeout handling  
**Scale**: Single user session based (no multi-tenant yet)

## Constitution Check

*No Constitution file found - skipping gate checks.*

## Project Structure

### Documentation (this feature)

```text
specs/011-job-queue/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── files/
│   │   ├── service.go    # Existing - add job tracking
│   │   ├── handler.go   # Existing - add job endpoints
│   │   └── job.go      # NEW - Job queue service
│   └── storage/
│       └── job_repo.go # NEW - MongoDB repo for jobs
frontend/src/
├── components/
│   └── ProgressJob.jsx # NEW - Progress bar component
├── hooks/
│   └── useJobPolling.js # NEW - Polling hook for job status
└── pages/
    └── Files.jsx       # Existing - update to use job polling
```

**Structure Decision**: Add job queue as new package under existing backend structure, reusing MongoDB for persistence. Frontend gets progress component and polling hook.

## Complexity Tracking

No complexity violations identified.

## Phase 0: Research

### Unknowns to Research

1. **Real-time progress from Go to frontend**: Should use WebSocket vs long-polling vs server-sent events?
2. **MongoDB for job status**: Store in existing `jobs` collection or extend `files` collection?
3. **File encryption progress tracking**: How to track percentage during AES-GCM streaming?

### Technologies to Research

1. **Go Fiber SSE (Server-Sent Events)**: Simple unidirectional updates - good for progress
2. **WebSocket libraries for Go**: gorilla/websocket or nhooyr.io/websocket
3. **Frontend polling**: setInterval with abort controller

### Best Practices

1. **Job queue patterns**: Redis Bull, Asynq, or custom MongoDB-based queue
2. **Progress percentage**: Calculate from bytes processed vs total bytes
3. **Immediate password error**: Derive key first, if fails return immediately

### Decision: Phase 0 Summary

- Use **long-polling** (every 1-2 seconds) for progress updates - simpler than WebSocket, works reliably
- Store job status in **new `jobs` collection** in MongoDB - clean separation from files
- Track progress by **bytes processed counter** in encryption loop - update every chunk

## Phase 1: Design & Contracts

### Data Model

Create `specs/011-job-queue/data-model.md`:

- **Job**: id, user_id, type (encrypt/decrypt), status, progress (0-100), file_id, result, error, created_at, updated_at
- **JobStatus**: pending → processing → completed (or failed/cancelled)

### Interface Contracts

Create `specs/011-job-queue/contracts/`:

- `POST /api/v1/jobs` - Create job (returns job_id)
- `GET /api/v1/jobs/:id` - Get job status
- `DELETE /api/v1/jobs/:id` - Cancel job

### Quickstart

Create `specs/011-job-queue/quickstart.md`:

- How to test: Upload a file and poll for status
- How to verify: Progress bar updates correctly

## Output Checklist

- [x] Research resolved all unknowns
- [x] Data model defined
- [x] API contracts defined
- [x] Quickstart guide written
- [x] Agent context update scripted

## Notes

- No external dependencies added (using existing MongoDB)
- Long-polling selected over WebSocket for simplicity
- Job storage in separate collection for clean separation