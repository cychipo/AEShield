import { useCallback, useEffect, useRef, useState } from "react";

const API_BASE_URL =
  import.meta.env.VITE_API_URL ||
  (import.meta.env.DEV ? "http://localhost:6888/api/v1" : "/api/v1");

export default function useJobPolling({ jobId, enabled = true, intervalMs = 1500 }) {
  const [job, setJob] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [isDisconnected, setIsDisconnected] = useState(false);
  const timerRef = useRef(null);

  const stop = useCallback(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  const poll = useCallback(async () => {
    if (!jobId || !enabled) {
      return;
    }
    const token = localStorage.getItem("aeshield_token");
    if (!token) {
      setError("Phiên đăng nhập đã hết hạn.");
      return;
    }

    setLoading(true);
    try {
      const response = await fetch(`${API_BASE_URL}/jobs/${jobId}`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      const payload = await response.json().catch(() => null);
      if (!response.ok) {
        setError(payload?.error || "Không thể tải trạng thái job.");
        stop();
        return;
      }
      setJob(payload);
      setError("");
      setIsDisconnected(false);
      if (["completed", "failed", "cancelled"].includes(payload?.status)) {
        stop();
        return;
      }
      timerRef.current = setTimeout(poll, intervalMs);
    } catch {
      setIsDisconnected(true);
      setError("Mất kết nối khi đang theo dõi tiến trình. Đang tự động kết nối lại...");
      timerRef.current = setTimeout(poll, intervalMs);
    } finally {
      setLoading(false);
    }
  }, [enabled, intervalMs, jobId, stop]);

  useEffect(() => {
    stop();
    if (jobId && enabled) {
      poll();
    }
    return stop;
  }, [jobId, enabled, poll, stop]);

  useEffect(() => {
    if (!jobId || !enabled) {
      return undefined;
    }

    const handleOnline = () => {
      if (!isDisconnected) {
        return;
      }
      stop();
      poll();
    };

    window.addEventListener("online", handleOnline);
    return () => window.removeEventListener("online", handleOnline);
  }, [enabled, isDisconnected, jobId, poll, stop]);

  return { job, loading, error, stop, isDisconnected };
}
