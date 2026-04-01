import { useEffect, useMemo, useRef, useState } from "react";
import { Bell } from "lucide-react";
import { useNotifications } from "../context/NotificationsContext";

function formatNotificationTime(value) {
  if (!value) {
    return "";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }

  return date.toLocaleString("vi-VN", {
    dateStyle: "short",
    timeStyle: "short",
  });
}

export default function NotificationsBell() {
  const {
    items,
    unreadCount,
    hasMore,
    loading,
    loadingMore,
    markingAllRead,
    hasLoaded,
    error,
    loadInitial,
    loadMore,
    markAllRead,
  } = useNotifications();
  const [open, setOpen] = useState(false);
  const containerRef = useRef(null);

  useEffect(() => {
    if (!open || hasLoaded) {
      return;
    }
    loadInitial();
  }, [hasLoaded, loadInitial, open]);

  useEffect(() => {
    if (!open) {
      return undefined;
    }

    const handlePointerDown = (event) => {
      if (containerRef.current && !containerRef.current.contains(event.target)) {
        setOpen(false);
      }
    };

    document.addEventListener("mousedown", handlePointerDown);
    return () => document.removeEventListener("mousedown", handlePointerDown);
  }, [open]);

  const unreadLabel = useMemo(() => {
    if (unreadCount <= 0) {
      return "";
    }
    if (unreadCount > 99) {
      return "99+";
    }
    return `${unreadCount}`;
  }, [unreadCount]);

  return (
    <div className="relative" ref={containerRef}>
      <button
        className="p-2 text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg relative"
        onClick={() => setOpen((current) => !current)}
        type="button"
      >
        <Bell size={20} />
        {unreadCount > 0 && (
          <>
            <span className="absolute top-1.5 right-1.5 w-2.5 h-2.5 bg-red-500 rounded-full border-2 border-white dark:border-slate-900"></span>
            <span className="absolute -top-1 -right-1 min-w-[18px] h-[18px] px-1 rounded-full bg-red-500 text-white text-[10px] font-semibold flex items-center justify-center">
              {unreadLabel}
            </span>
          </>
        )}
      </button>

      {open && (
        <div className="absolute right-0 mt-2 w-[360px] max-w-[calc(100vw-2rem)] rounded-xl border border-primary/10 bg-white dark:bg-slate-900 shadow-xl z-50 overflow-hidden">
          <div className="px-4 py-3 border-b border-primary/10 flex items-center justify-between gap-4">
            <div>
              <h3 className="text-sm font-semibold">Thông báo</h3>
              <p className="text-xs text-slate-500 dark:text-slate-400">
                {unreadCount > 0 ? `${unreadCount} chưa đọc` : "Tất cả đã đọc"}
              </p>
            </div>
            <button
              className="text-xs font-medium text-primary disabled:text-slate-400"
              disabled={markingAllRead || items.length === 0 || unreadCount === 0}
              onClick={markAllRead}
              type="button"
            >
              {markingAllRead ? "Đang cập nhật..." : "Đánh dấu đã đọc tất cả"}
            </button>
          </div>

          <div className="max-h-[420px] overflow-y-auto">
            {loading && !hasLoaded ? (
              <div className="p-4 text-sm text-slate-500 dark:text-slate-400">
                Đang tải thông báo...
              </div>
            ) : error ? (
              <div className="p-4 text-sm text-red-600 dark:text-red-400">{error}</div>
            ) : items.length === 0 ? (
              <div className="p-4 text-sm text-slate-500 dark:text-slate-400">
                Chưa có thông báo nào.
              </div>
            ) : (
              <div className="divide-y divide-primary/10">
                {items.map((item) => (
                  <div
                    key={item.id}
                    className={`px-4 py-3 ${item.is_read ? "bg-white dark:bg-slate-900" : "bg-primary/5 dark:bg-primary/10"}`}
                  >
                    <p className="text-sm font-medium text-slate-800 dark:text-slate-100 leading-6">
                      <span>{item.actor?.name || item.actor?.email || "Một người dùng"}</span>{" "}
                      <span className="font-normal">đã thêm bạn vào tệp</span>{" "}
                      <span className="break-words">{item.file?.filename || "Không xác định"}</span>
                    </p>
                    <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">
                      {formatNotificationTime(item.created_at)}
                    </p>
                  </div>
                ))}
              </div>
            )}
          </div>

          {hasMore && !loading && !error && (
            <div className="px-4 py-3 border-t border-primary/10">
              <button
                className="w-full text-sm font-medium text-primary disabled:text-slate-400"
                disabled={loadingMore}
                onClick={loadMore}
                type="button"
              >
                {loadingMore ? "Đang tải thêm..." : "Xem thêm"}
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
