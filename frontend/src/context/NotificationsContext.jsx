import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";

const API_BASE_URL =
  import.meta.env.VITE_API_URL || "http://localhost:6888/api/v1";

const NotificationsContext = createContext(null);

export function NotificationsProvider({ children }) {
  const navigate = useNavigate();
  const [items, setItems] = useState([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [nextCursor, setNextCursor] = useState("");
  const [loading, setLoading] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [markingAllRead, setMarkingAllRead] = useState(false);
  const [hasLoaded, setHasLoaded] = useState(false);
  const [error, setError] = useState("");

  const handleUnauthorized = useCallback(() => {
    localStorage.removeItem("aeshield_token");
    localStorage.removeItem("aeshield_user");
    navigate("/", { replace: true });
  }, [navigate]);

  const fetchNotifications = useCallback(
    async ({ cursor = "", append = false } = {}) => {
      const token = localStorage.getItem("aeshield_token");
      if (!token) {
        handleUnauthorized();
        return;
      }

      if (append) {
        setLoadingMore(true);
      } else {
        setLoading(true);
      }
      setError("");

      try {
        const params = new URLSearchParams({ limit: "5" });
        if (cursor) {
          params.set("cursor", cursor);
        }

        const response = await fetch(`${API_BASE_URL}/notifications?${params.toString()}`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (response.status === 401) {
          handleUnauthorized();
          return;
        }

        let payload = null;
        try {
          payload = await response.json();
        } catch {
          payload = null;
        }

        if (!response.ok) {
          setError(payload?.error || "Không thể tải thông báo.");
          return;
        }

        const nextItems = Array.isArray(payload?.items) ? payload.items : [];
        setItems((previous) => (append ? [...previous, ...nextItems] : nextItems));
        setUnreadCount(Number.isFinite(payload?.unread_count) ? payload.unread_count : 0);
        setHasMore(Boolean(payload?.has_more));
        setNextCursor(payload?.next_cursor || "");
        setHasLoaded(true);
      } catch (fetchError) {
        console.error("Error fetching notifications:", fetchError);
        setError("Có lỗi xảy ra khi tải thông báo.");
      } finally {
        setLoading(false);
        setLoadingMore(false);
      }
    },
    [handleUnauthorized]
  );

  const loadInitial = useCallback(async () => {
    if (hasLoaded || loading) {
      return;
    }
    await fetchNotifications({ append: false });
  }, [fetchNotifications, hasLoaded, loading]);

  const refresh = useCallback(async () => {
    await fetchNotifications({ append: false });
  }, [fetchNotifications]);

  const loadMore = useCallback(async () => {
    if (!hasMore || !nextCursor || loadingMore) {
      return;
    }
    await fetchNotifications({ cursor: nextCursor, append: true });
  }, [fetchNotifications, hasMore, loadingMore, nextCursor]);

  const markAllRead = useCallback(async () => {
    const token = localStorage.getItem("aeshield_token");
    if (!token) {
      handleUnauthorized();
      return;
    }

    setMarkingAllRead(true);
    setError("");

    try {
      const response = await fetch(`${API_BASE_URL}/notifications/read-all`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (response.status === 401) {
        handleUnauthorized();
        return;
      }

      let payload = null;
      try {
        payload = await response.json();
      } catch {
        payload = null;
      }

      if (!response.ok) {
        setError(payload?.error || "Không thể đánh dấu đã đọc tất cả.");
        return;
      }

      setUnreadCount(0);
      setItems((previous) => previous.map((item) => ({
        ...item,
        is_read: true,
        read_at: payload?.read_at || item.read_at,
      })));
    } catch (markError) {
      console.error("Error marking notifications as read:", markError);
      setError("Có lỗi xảy ra khi cập nhật thông báo.");
    } finally {
      setMarkingAllRead(false);
    }
  }, [handleUnauthorized]);

  useEffect(() => {
    const token = localStorage.getItem("aeshield_token");
    if (!token) {
      return;
    }
    fetchNotifications({ append: false });
  }, [fetchNotifications]);

  const value = useMemo(
    () => ({
      items,
      unreadCount,
      hasMore,
      nextCursor,
      loading,
      loadingMore,
      markingAllRead,
      hasLoaded,
      error,
      loadInitial,
      loadMore,
      markAllRead,
      refresh,
    }),
    [error, hasLoaded, hasMore, items, loadInitial, loadMore, loading, loadingMore, markingAllRead, nextCursor, refresh, unreadCount, markAllRead]
  );

  return (
    <NotificationsContext.Provider value={value}>
      {children}
    </NotificationsContext.Provider>
  );
}

export function useNotifications() {
  const context = useContext(NotificationsContext);
  if (!context) {
    throw new Error("useNotifications must be used within NotificationsProvider");
  }
  return context;
}
