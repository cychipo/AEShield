export default function JobProgress({ job, error, onCancel, cancelling = false }) {
  if (!job && !error) {
    return null;
  }

  const progress = Math.max(0, Math.min(job?.progress ?? 0, 100));
  const status = job?.status || "processing";

  return (
    <div className="rounded-xl border border-primary/15 bg-primary/5 p-4 space-y-3">
      <div className="flex items-center justify-between gap-4">
        <div>
          <p className="text-sm font-semibold text-slate-800 dark:text-slate-100">
            {job?.filename || "Đang xử lý tệp tin"}
          </p>
          <p className="text-xs text-slate-500 dark:text-slate-400">
            Trạng thái: {status}
          </p>
        </div>
        {status === "processing" && onCancel && (
          <button
            type="button"
            onClick={onCancel}
            disabled={cancelling}
            className="text-xs font-semibold text-red-600 hover:underline disabled:opacity-50"
          >
            {cancelling ? "Đang hủy..." : "Hủy"}
          </button>
        )}
      </div>

      <div className="w-full bg-slate-200 dark:bg-slate-700 h-2 rounded-full overflow-hidden">
        <div
          className="h-2 bg-primary transition-all duration-300"
          style={{ width: `${progress}%` }}
        />
      </div>

      <div className="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400">
        <span>{progress}%</span>
        <span>
          {status === "completed"
            ? "Hoàn thành"
            : status === "failed"
              ? "Thất bại"
              : status === "cancelled"
                ? "Đã hủy"
                : "Đang xử lý"}
        </span>
      </div>

      {error && <p className="text-sm text-red-600">{error}</p>}
      {job?.error?.message && <p className="text-sm text-red-600">{job.error.message}</p>}
    </div>
  );
}
