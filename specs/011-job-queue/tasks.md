# Tasks: Job Queue for File Encryption and Decryption

**Feature**: Job Queue for File Processing | **Branch**: 011-job-queue  
**Spec**: specs/011-job-queue/spec.md | **Plan**: specs/011-job-queue/plan.md

---

## Phase 1: Setup (Foundation)

Tasks to initialize the job queue feature infrastructure.

- [X] T001 Create Job model in backend/models/entities.go with all fields (ID, UserID, Type, Status, Progress, FileMetadataID, Result, Error, CreatedAt, UpdatedAt, CompletedAt)
- [X] T002 [P] Add JobStatus and JobType constants in backend/models/entities.go
- [X] T003 Create JobRepository interface in backend/internal/storage/job_repo.go
- [X] T004 Implement JobRepository MongoDB methods (Create, FindByID, FindByUserID, Update, Delete) in backend/internal/storage/job_repo.go
- [X] T005 Create indexes for jobs collection (user_id + created_at, status, created_at) in backend/internal/storage/job_repo.go
- [X] T006 [P] Create job service package in backend/internal/files/job_service.go with JobService struct and methods (CreateJob, GetJob, UpdateProgress, Complete, Fail, Cancel)

---

## Phase 2: Foundational (Blocking Prerequisites)

Core components required before any user story can be implemented.

- [X] T007 Integrate job creation into existing file upload flow - modify backend/internal/files/service.go to create job before encryption starts
- [X] T008 Add progress tracking callback to encryption process - update crypto.Encrypt to report bytes processed
- [X] T009 Update job service to update progress based on bytes encrypted in backend/internal/files/job_service.go
- [X] T010 Create job handler in backend/internal/files/job_handler.go with endpoints (POST /jobs, GET /jobs/:id, DELETE /jobs/:id, GET /jobs)
- [X] T011 Register job routes in backend/cmd/main.go under /api/v1/jobs with JWT middleware

---

## Phase 3: User Story 1 - Upload File with Progress (P1)

Users upload files and see real-time progress bar.

**Independent Test**: Upload any file and observe progress bar from 0% to 100%

- [X] T012 [P] [US1] Update frontend upload form in frontend/src/pages/Files.jsx to call new job API instead of synchronous upload
- [X] T013 [US1] Create JobProgress component in frontend/src/components/JobProgress.jsx with progress bar, percentage display, and status text
- [X] T014 [US1] Create useJobPolling hook in frontend/src/hooks/useJobPolling.js to poll job status every 1-2 seconds
- [X] T015 [US1] Integrate JobProgress component into Files.jsx upload flow showing progress during encryption
- [X] T016 [US1] Handle job completion state - redirect to file list or show download link in frontend/src/pages/Files.jsx
- [X] T017 [US1] Handle job failure state - show error message and allow retry in frontend/src/pages/Files.jsx
- [X] T018 [US1] Verify progress updates appear at least every 2 seconds during upload

---

## Phase 4: User Story 2 - Decrypt File with Progress (P1)

Users decrypt files and see real-time progress bar.

**Independent Test**: Decrypt any encrypted file and observe progress bar from 0% to 100%

- [X] T019 [P] [US2] Create decrypt job endpoint in backend/internal/files/job_handler.go (POST /jobs with type=decrypt)
- [X] T020 [US2] Add progress tracking to decryption process - update crypto.Decrypt to report bytes processed
- [X] T021 [US2] Update job service to update progress during decryption in backend/internal/files/job_service.go
- [X] T022 [US2] Implement fail-fast password validation - derive key first before full decryption in backend/internal/crypto/crypto.go
- [X] T023 [US2] Update frontend decrypt flow in frontend/src/pages/Files.jsx to use job polling instead of synchronous decrypt
- [X] T024 [US2] Integrate JobProgress component into decrypt modal in frontend/src/pages/Files.jsx
- [X] T025 [US2] Handle wrong password error immediately (not after full decrypt) - show error within 5 seconds
- [X] T026 [US2] Handle decryption completed state - show preview or download in frontend/src/pages/Files.jsx

---

## Phase 5: User Story 3 - Check Job Status (P2)

Users can query job status via API.

**Independent Test**: Call GET /api/v1/jobs/:id and receive current status with percentage

- [X] T027 [P] [US3] Implement GET /api/v1/jobs/:id endpoint in backend/internal/files/job_handler.go returning full job details
- [X] T028 [US3] Implement GET /api/v1/jobs endpoint to list user's jobs with filters (status, limit, offset) in backend/internal/files/job_handler.go
- [X] T029 [US3] Add job status display in frontend - show list of recent jobs in Files.jsx or new JobsList component
- [X] T030 [US3] Allow users to view completed job results (download URL for encrypt, preview for decrypt)

---

## Phase 6: User Story 4 - Job Queue Management (P3)

System manages multiple concurrent jobs efficiently.

**Independent Test**: Send multiple upload requests simultaneously and verify each gets unique job ID

- [X] T031 [P] [US4] Implement DELETE /api/v1/jobs/:id endpoint to cancel running jobs in backend/internal/files/job_handler.go
- [X] T032 [US4] Add cancellation flag to job processing - check cancel status during encryption/decryption loops
- [X] T033 [US4] Implement job cleanup for old completed jobs (keep for 24 hours per spec) in backend/internal/storage/job_repo.go
- [X] T034 [US4] Verify system handles 10+ concurrent jobs without performance degradation
- [X] T035 [US4] Add frontend UI to show cancel button for running jobs in frontend/src/components/JobProgress.jsx

---

## Phase 7: Polish & Cross-Cutting Concerns

Final improvements and edge case handling.

- [X] T036 Add timeout handling - if job takes too long, mark as failed with timeout error
- [X] T037 Handle network disconnection during polling - allow resume when connection restored
- [X] T038 Add job history in frontend - show past 20 jobs with status in Files.jsx
- [X] T039 Optimize progress calculation - ensure percentage is accurate based on total file size
- [X] T040 Add loading states for job list - show skeleton while fetching in frontend/src/pages/Files.jsx

---

## Dependencies

```
Phase 1 (Setup)
  └── Phase 2 (Foundational)
        ├── Phase 3 (US1 - Upload)
        ├── Phase 4 (US2 - Decrypt) depends on Phase 3
        ├── Phase 5 (US3 - Status) depends on Phase 2
        └── Phase 6 (US4 - Queue) depends on Phase 3
              └── Phase 7 (Polish)
```

---

## Parallel Opportunities

| Tasks | Reason |
|-------|--------|
| T002, T003 | Different files (models, repo interface) |
| T012, T013 | Frontend components can be created in parallel |
| T019, T020 | Backend changes can be parallel |
| T027, T029 | Both use job data, can parallel with T019 |

---

## Summary

| Metric | Count |
|--------|-------|
| Total Tasks | 40 |
| User Story 1 (Upload) | 7 tasks |
| User Story 2 (Decrypt) | 7 tasks |
| User Story 3 (Status) | 4 tasks |
| User Story 4 (Queue) | 5 tasks |
| Setup/Foundational | 11 tasks |
| Polish | 6 tasks |

---

## MVP Scope

**Minimum Viable Product**: Tasks for User Story 1 only

- T001-T006: Setup
- T007-T011: Foundational  
- T012-T018: User Story 1 (Upload with Progress)

This enables basic upload with progress bar. User can then extend to decrypt (T019-T026) and other stories as needed.