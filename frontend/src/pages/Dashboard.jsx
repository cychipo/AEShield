import { useState, useEffect } from "react";
import { useNavigate, Link } from "react-router-dom";
import {
  Shield,
  LayoutDashboard,
  FolderOpen,
  Settings,
  Lock,
  Database,
  ShieldAlert,
  FileText,
  CloudUpload,
  Radar,
  ShieldCheck,
  User,
  Link as LinkIcon,
} from "lucide-react";
import AppHeader from "../components/AppHeader";

const API_BASE_URL =
  import.meta.env.VITE_API_URL ||
  (import.meta.env.DEV ? "http://localhost:6888/api/v1" : "/api/v1");

export default function Dashboard() {
  const [user, setUser] = useState(null);
  const [storageUsage, setStorageUsage] = useState({
    used_bytes: 0,
    quota_bytes: 10 * 1024 * 1024 * 1024,
    used_gb: 0,
    quota_gb: 10,
    percent_used: 0,
    file_count: 0,
    available_bytes: 10 * 1024 * 1024 * 1024,
  });
  const [ownedFiles, setOwnedFiles] = useState([]);
  const [sharedFiles, setSharedFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    fetchUser();
  }, []);

  const fetchUser = async () => {
    const token = localStorage.getItem("aeshield_token");

    if (!token) {
      navigate("/", { replace: true });
      return;
    }

    try {
      const response = await fetch(`${API_BASE_URL}/auth/me`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const userData = await response.json();
        setUser(userData);

        const [storageResponse, filesResponse] = await Promise.all([
          fetch(`${API_BASE_URL}/storage/me`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }),
          fetch(`${API_BASE_URL}/files`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }),
        ]);

        if (storageResponse.ok) {
          const storageData = await storageResponse.json();
          setStorageUsage(storageData);
        } else if (storageResponse.status === 401) {
          localStorage.removeItem("aeshield_token");
          localStorage.removeItem("aeshield_user");
          navigate("/", { replace: true });
          return;
        }

        if (filesResponse.ok) {
          const filesData = await filesResponse.json();
          // API trả về { owned_files, shared_with_me }
          const owned = Array.isArray(filesData.owned_files) ? filesData.owned_files : [];
          const shared = Array.isArray(filesData.shared_with_me) ? filesData.shared_with_me : [];
          // Gắn flag isOwned để phân biệt
          const ownedWithFlag = owned.map((f) => ({ ...f, isOwned: true }));
          const sharedWithFlag = shared.map((f) => ({ ...f, isOwned: false }));
          setOwnedFiles(ownedWithFlag);
          setSharedFiles(sharedWithFlag);
        } else if (filesResponse.status === 401) {
          localStorage.removeItem("aeshield_token");
          localStorage.removeItem("aeshield_user");
          navigate("/", { replace: true });
          return;
        }
      } else {
        localStorage.removeItem("aeshield_token");
        localStorage.removeItem("aeshield_user");
        navigate("/", { replace: true });
      }
    } catch (error) {
      console.error("Error fetching user:", error);
      localStorage.removeItem("aeshield_token");
      localStorage.removeItem("aeshield_user");
      navigate("/", { replace: true });
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background-light dark:bg-background-dark">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-charcoal dark:text-slate-100">Đang tải...</p>
        </div>
      </div>
    );
  }

  const isNearQuota = storageUsage.percent_used >= 90;

  const formatStorageValue = (bytes) => {
    if (!Number.isFinite(bytes) || bytes <= 0) {
      return "0 KB";
    }

    if (bytes >= 1024 * 1024 * 1024) {
      return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
    }

    if (bytes >= 1024 * 1024) {
      return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
    }

    return `${(bytes / 1024).toFixed(2)} KB`;
  };

  const recentFiles = [...ownedFiles, ...sharedFiles].slice(0, 10);

  return (
    <div className="flex h-screen overflow-hidden bg-background-light dark:bg-background-dark">
      {/* Sidebar Navigation */}
      <aside className="w-64 border-r border-primary/10 bg-white dark:bg-slate-900 flex flex-col">
        <div className="p-6">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center text-white">
              <Shield size={18} />
            </div>
            <div>
              <h1 className="text-lg font-bold leading-none">AEShield</h1>
              <p className="text-xs text-primary font-medium">
                Bảo mật Doanh nghiệp
              </p>
            </div>
          </div>
        </div>
        <nav className="flex-1 px-4 space-y-1">
          <Link
            to="/dashboard"
            className="flex items-center gap-3 px-3 py-2 rounded-lg bg-primary/10 text-primary font-medium hover:bg-primary/20 hover:text-primary dark:hover:bg-primary/25 dark:hover:text-primary hover:shadow-sm transition-all"
          >
            <LayoutDashboard size={20} />
            <span>Dashboard</span>
          </Link>
          <Link
            to="/files"
            className="flex items-center gap-3 px-3 py-2 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-primary/5 hover:text-primary dark:hover:bg-primary/10 dark:hover:text-primary transition-colors"
          >
            <FolderOpen size={20} />
            <span>Tệp tin</span>
          </Link>
          <Link
            to="/settings"
            className="flex items-center gap-3 px-3 py-2 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-primary/5 hover:text-primary dark:hover:bg-primary/10 dark:hover:text-primary transition-colors"
          >
            <Settings size={20} />
            <span>Cài đặt tài khoản</span>
          </Link>
        </nav>
        <div className="p-4 border-t border-primary/10">
          <div className="bg-primary/5 rounded-xl p-4">
            <p className="text-xs font-semibold text-slate-500 mb-2 uppercase tracking-wider">
              Dung lượng lưu trữ
            </p>
            <div className="w-full bg-slate-200 dark:bg-slate-700 h-1.5 rounded-full mb-2">
              <div
                className={`${isNearQuota ? "bg-red-500" : "bg-primary"} h-1.5 rounded-full`}
                style={{ width: `${Math.min(storageUsage.percent_used || 0, 100)}%` }}
              ></div>
            </div>
            <p className="text-xs text-slate-600 dark:text-slate-400">
              {formatStorageValue(storageUsage.used_bytes)} / {formatStorageValue(storageUsage.quota_bytes)} đã sử dụng
            </p>
          </div>
        </div>
      </aside>

      {/* Main Content Area */}
      <main className="flex-1 flex flex-col overflow-hidden">
        <AppHeader
          user={user}
          searchPlaceholder="Tìm kiếm tệp tin đã mã hóa..."
        />

        {/* Scrollable Content */}
        <div className="flex-1 overflow-y-auto p-8">
          <div className="max-w-6xl mx-auto space-y-8">
            {isNearQuota && (
              <div className="border border-red-300 bg-red-50 text-red-700 dark:bg-red-950/30 dark:text-red-300 rounded-xl px-4 py-3 text-sm font-medium">
                Cảnh báo: Dung lượng lưu trữ đã đạt {storageUsage.percent_used.toFixed(2)}%. Vui lòng xóa bớt tệp để tránh gián đoạn upload.
              </div>
            )}

            {/* Stats Grid */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="bg-white dark:bg-slate-900 p-6 rounded-xl border border-primary/10 shadow-sm">
                <div className="flex justify-between items-start mb-4">
                  <div className="p-2 bg-primary/10 rounded-lg text-primary">
                    <Lock size={22} />
                  </div>
                  <span className="text-emerald-500 text-xs font-bold bg-emerald-500/10 px-2 py-1 rounded">
                    +12%
                  </span>
                </div>
                <h3 className="text-slate-500 text-sm font-medium">
                  Tổng tệp tin đã mã hóa
                </h3>
                <p className="text-3xl font-bold mt-1">{storageUsage.file_count || 0}</p>
              </div>
              <div className="bg-white dark:bg-slate-900 p-6 rounded-xl border border-primary/10 shadow-sm">
                <div className="flex justify-between items-start mb-4">
                  <div className="p-2 bg-primary/10 rounded-lg text-primary">
                    <Database size={22} />
                  </div>
                  <span className="text-emerald-500 text-xs font-bold bg-emerald-500/10 px-2 py-1 rounded">
                    +5%
                  </span>
                </div>
                <h3 className="text-slate-500 text-sm font-medium">
                  Dung lượng đã sử dụng
                </h3>
                <p className="text-3xl font-bold mt-1">{formatStorageValue(storageUsage.used_bytes)}</p>
              </div>
              <div className="bg-white dark:bg-slate-900 p-6 rounded-xl border border-primary/10 shadow-sm">
                <div className="flex justify-between items-start mb-4">
                  <div className="p-2 bg-primary/10 rounded-lg text-primary">
                    <ShieldAlert size={22} />
                  </div>
                  <span className="text-slate-400 text-xs font-bold bg-slate-400/10 px-2 py-1 rounded">
                    Bảo mật
                  </span>
                </div>
                <h3 className="text-slate-500 text-sm font-medium">
                  Mối đe dọa bị chặn
                </h3>
                <p className="text-3xl font-bold mt-1">0</p>
              </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
              {/* Recent Activity Table */}
              <div className="lg:col-span-2 bg-white dark:bg-slate-900 rounded-xl border border-primary/10 shadow-sm overflow-hidden">
                <div className="p-6 border-b border-primary/10 flex items-center justify-between">
                  <h3 className="font-bold text-lg">Hoạt động gần đây</h3>
                  <button
                    onClick={() => navigate("/files")}
                    className="text-primary text-sm font-semibold hover:underline"
                  >
                    Xem tất cả
                  </button>
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-left">
                    <thead className="bg-slate-50 dark:bg-slate-800/50 text-slate-500 text-xs uppercase tracking-wider">
                      <tr>
                        <th className="px-6 py-4 font-semibold">Tên tệp</th>
                        <th className="px-6 py-4 font-semibold">Loại</th>
                        <th className="px-6 py-4 font-semibold">Trạng thái</th>
                        <th className="px-6 py-4 font-semibold">Thời gian</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-primary/5">
                      {recentFiles.length === 0 ? (
                        <tr>
                          <td
                            className="px-6 py-6 text-sm text-slate-500"
                            colSpan={4}
                          >
                            Chưa có dữ liệu tệp tin.
                          </td>
                        </tr>
                      ) : (
                        recentFiles.map((file) => (
                          <tr
                            className="hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
                            key={file.id || file.storage_path || file.filename}
                          >
                            <td className="px-6 py-4 font-medium flex items-center gap-2">
                              <FileText size={18} className="text-slate-400" />
                              {file.filename || "-"}
                            </td>
                            <td className="px-6 py-4 text-sm">
                              {file.isOwned ? (
                                <span className="inline-flex items-center gap-1 text-blue-600 dark:text-blue-400">
                                  <User size={14} />
                                  Sở hữu
                                </span>
                              ) : (
                                <span className="inline-flex items-center gap-1 text-amber-600 dark:text-amber-400">
                                  <LinkIcon size={14} />
                                  Được chia sẻ
                                </span>
                              )}
                            </td>
                            <td className="px-6 py-4">
                              <span className="px-2 py-1 bg-emerald-500/10 text-emerald-500 text-xs rounded-full font-medium">
                                {file.encryption_type || "AES"}
                              </span>
                            </td>
                            <td className="px-6 py-4 text-sm text-slate-500">
                              {file.updated_at
                                ? new Date(file.updated_at).toLocaleString("vi-VN")
                                : "-"}
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              </div>

              {/* Quick Actions */}
              <div className="space-y-6">
                <h3 className="font-bold text-lg">Thao tác nhanh</h3>
                <div
                  onClick={() => navigate("/files?upload=true")}
                  className="group bg-white dark:bg-slate-900 p-6 rounded-xl border border-primary/10 shadow-sm hover:border-primary transition-all cursor-pointer"
                >
                  <div className="flex items-center gap-4">
                    <div className="p-3 bg-primary text-white rounded-xl">
                      <CloudUpload size={22} />
                    </div>
                    <div>
                      <h4 className="font-bold">Tải lên tệp mới</h4>
                      <p className="text-sm text-slate-500">
                        Mã hóa và lưu trữ an toàn
                      </p>
                    </div>
                  </div>
                </div>
                <div className="group bg-white dark:bg-slate-900 p-6 rounded-xl border border-primary/10 shadow-sm hover:border-primary transition-all cursor-pointer">
                  <div className="flex items-center gap-4">
                    <div className="p-3 bg-slate-900 dark:bg-slate-800 text-white rounded-xl group-hover:bg-primary transition-colors">
                      <Radar size={22} />
                    </div>
                    <div>
                      <h4 className="font-bold">Quét hệ thống</h4>
                      <p className="text-sm text-slate-500">
                        Quét sâu lỗ hổng bảo mật
                      </p>
                    </div>
                  </div>
                </div>
                {/* Additional Feature Card */}
                <div className="relative overflow-hidden bg-primary/10 p-6 rounded-xl border border-primary/20">
                  <div className="relative z-10">
                    <h4 className="font-bold text-primary mb-1">
                      Kiểm tra bảo mật hàng tuần
                    </h4>
                    <p className="text-sm text-slate-600 dark:text-slate-400 mb-4">
                      Hệ thống của bạn được bảo vệ 98% tuần này.
                    </p>
                    <button className="bg-white dark:bg-slate-900 text-slate-900 dark:text-white px-4 py-2 rounded-lg text-sm font-bold shadow-sm">
                      Xem báo cáo
                    </button>
                  </div>
                  <div className="absolute -bottom-4 -right-4 opacity-10">
                    <ShieldCheck size={100} />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
