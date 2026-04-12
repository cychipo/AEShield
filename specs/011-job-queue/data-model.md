# Data Model: Job Queue

## Entities

### Job

Represents a file processing task (encryption or decryption).

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| id | ObjectID | Unique job ID | Auto-generated |
| user_id | string | Owner of the job | Required |
| type | string | "encrypt" or "decrypt" | Required, enum |
| status | string | Current status | Required, enum |
| progress | int | 0-100 percentage | 0-100 |
| file_id | ObjectID | Reference to file (encrypt only) | Optional |
| file_metadata_id | ObjectID | Reference to file metadata | Optional |
| result | object | Job result (URL, preview data) | Optional |
| error | object | Error details if failed | Optional |
| created_at | datetime | Job creation time | Auto-generated |
| updated_at | datetime | Last update time | Auto-generated |
| completed_at | datetime | Completion time | Optional |

### JobStatus Enum

| Value | Description |
|-------|-------------|
| pending | Job created, waiting to be processed |
| processing | Job is actively running |
| completed | Job finished successfully |
| failed | Job encountered an error |
| cancelled | Job was cancelled by user |

### JobError

| Field | Type | Description |
|-------|------|-------------|
| code | string | Error code (e.g., "timeout", "invalid_password") |
| message | string | Human-readable error message |
| timestamp | datetime | When error occurred |

### JobResult

| Field | Type | Description |
|-------|------|-------------|
| file_id | string | ID of processed file |
| download_url | string | Presigned URL for download (encrypt) |
| preview_data | bytes | Decrypted content preview (decrypt) |

## MongoDB Collection

### Collection: jobs

```javascript
{
  _id: ObjectId,
  user_id: "user-123",
  type: "encrypt", // or "decrypt"
  status: "processing",
  progress: 45,
  file_id: ObjectId("..."),
  file_metadata_id: ObjectId("..."),
  result: { file_id: "...", download_url: "..." },
  error: null,
  created_at: ISODate("2026-04-12T10:00:00Z"),
  updated_at: ISODate("2026-04-12T10:00:30Z"),
  completed_at: null
}
```

## Indexes

| Index | Fields | Purpose |
|-------|--------|---------|
| user_id_1_created_at | user_id, created_at | List user's jobs |
| status_1 | status | Query by status |
| created_at_1 | created_at | Cleanup old jobs |

## State Transitions

```
pending → processing → completed
         ↘ failed
         ↘ cancelled
```

## Relationships

- Job (many) → User (one): Each job belongs to one user
- Job (many) → FileMetadata (one): Encrypt jobs reference file metadata