import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import {
  Shield,
  LayoutDashboard,
  FolderOpen,
  UserCheck,
  Settings as SettingsIcon,
  Search,
  Bell,
} from "lucide-react";

const API_BASE_URL =
  import.meta.env.VITE_API_URL || "http://localhost:6888/api/v1";

export default function Settings() {
  const [user, setUser] = useState(null);
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
          <a
            className="flex items-center gap-3 px-3 py-2 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
            href="/dashboard"
          >
            <LayoutDashboard size={20} />
            <span>Dashboard</span>
          </a>
          <a
            className="flex items-center gap-3 px-3 py-2 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
            href="/files"
          >
            <FolderOpen size={20} />
            <span>Tệp tin</span>
          </a>
          <a
            className="flex items-center gap-3 px-3 py-2 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
            href="/whitelist"
          >
            <UserCheck size={20} />
            <span>Danh sách tin cậy</span>
          </a>
          <a
            className="flex items-center gap-3 px-3 py-2 rounded-lg bg-primary/10 text-primary font-medium"
            href="/settings"
          >
            <SettingsIcon size={20} />
            <span>Cài đặt tài khoản</span>
          </a>
        </nav>
        <div className="p-4 border-t border-primary/10">
          <div className="bg-primary/5 rounded-xl p-4">
            <p className="text-xs font-semibold text-slate-500 mb-2 uppercase tracking-wider">
              Dung lượng lưu trữ
            </p>
            <div className="w-full bg-slate-200 dark:bg-slate-700 h-1.5 rounded-full mb-2">
              <div
                className="bg-primary h-1.5 rounded-full"
                style={{ width: "64%" }}
              ></div>
            </div>
            <p className="text-xs text-slate-600 dark:text-slate-400">
              64.2 GB / 100 GB đã sử dụng
            </p>
          </div>
        </div>
      </aside>

      {/* Main Content Area */}
      <main className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <header className="h-16 border-b border-primary/10 bg-white dark:bg-slate-900 flex items-center justify-between px-8">
          <div className="flex items-center gap-4 flex-1 max-w-xl">
            <div className="relative w-full">
              <Search
                size={18}
                className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
              />
              <input
                className="w-full pl-10 pr-4 py-2 bg-slate-50 dark:bg-slate-800 border-none rounded-lg focus:ring-1 focus:ring-primary text-sm"
                placeholder="Tìm kiếm cài đặt..."
                type="text"
              />
            </div>
          </div>
          <div className="flex items-center gap-4">
            <button className="p-2 text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg relative">
              <Bell size={20} />
              <span className="absolute top-2 right-2 w-2 h-2 bg-red-500 rounded-full border-2 border-white dark:border-slate-900"></span>
            </button>
            <div className="h-8 w-px bg-primary/10 mx-2"></div>
            <div className="flex items-center gap-3">
              <div className="text-right hidden sm:block">
                <p className="text-sm font-semibold leading-none">
                  {user?.name || "User"}
                </p>
                <p className="text-xs text-slate-500">{user?.email || ""}</p>
              </div>
              <div
                className="w-10 h-10 rounded-full bg-slate-200 dark:bg-slate-700 border-2 border-primary/20 bg-cover bg-center"
                style={{
                  backgroundImage: user?.avatar
                    ? `url(${user.avatar})`
                    : "none",
                }}
              ></div>
            </div>
          </div>
        </header>

        {/* Scrollable Content */}
        <div className="flex-1 overflow-y-auto p-8">
          <div className="max-w-4xl mx-auto">
            <h1 className="text-2xl font-bold mb-8">Cài đặt tài khoản</h1>

            {/* Account Settings Card */}
            <div className="bg-white dark:bg-slate-900 rounded-xl border border-primary/10 shadow-sm p-6 mb-6">
              <h2 className="text-lg font-semibold mb-4">
                Thông tin tài khoản
              </h2>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
                    Tên hiển thị
                  </label>
                  <input
                    type="text"
                    defaultValue={user?.name || ""}
                    className="w-full px-3 py-2 border border-slate-300 dark:border-slate-600 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
                    Email
                  </label>
                  <input
                    type="email"
                    defaultValue={user?.email || ""}
                    disabled
                    className="w-full px-3 py-2 border border-slate-300 dark:border-slate-600 rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 cursor-not-allowed"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
                    Ảnh đại diện
                  </label>
                  <div className="flex items-center gap-4">
                    <div
                      className="w-12 h-12 rounded-full bg-slate-200 dark:bg-slate-700 border-2 border-primary/20 bg-cover bg-center"
                      style={{
                        backgroundImage: user?.avatar
                          ? `url(${user.avatar})`
                          : "none",
                      }}
                    ></div>
                    <button className="text-primary hover:underline text-sm">
                      Thay đổi ảnh
                    </button>
                  </div>
                </div>
              </div>
            </div>

            {/* Security Settings Card */}
            <div className="bg-white dark:bg-slate-900 rounded-xl border border-primary/10 shadow-sm p-6 mb-6">
              <h2 className="text-lg font-semibold mb-4">Bảo mật</h2>
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium">Mật khẩu</p>
                    <p className="text-sm text-slate-500 dark:text-slate-400">
                      Thay đổi mật khẩu tài khoản của bạn
                    </p>
                  </div>
                  <button className="bg-primary text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-orange-600 transition-colors">
                    Đổi mật khẩu
                  </button>
                </div>
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium">Xác thực hai yếu tố</p>
                    <p className="text-sm text-slate-500 dark:text-slate-400">
                      Thêm lớp bảo vệ bổ sung cho tài khoản
                    </p>
                  </div>
                  <button className="border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-300 px-4 py-2 rounded-lg text-sm font-medium hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
                    Thiết lập 2FA
                  </button>
                </div>
              </div>
            </div>

            {/* Danger Zone */}
            <div className="bg-white dark:bg-slate-900 rounded-xl border border-red-200 dark:border-red-800 shadow-sm p-6">
              <h2 className="text-lg font-semibold text-red-600 dark:text-red-400 mb-4">
                Khu vực nguy hiểm
              </h2>
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium text-red-600 dark:text-red-400">
                      Xóa tài khoản
                    </p>
                    <p className="text-sm text-slate-500 dark:text-slate-400">
                      Hành động này không thể hoàn tác
                    </p>
                  </div>
                  <button className="bg-red-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-red-700 transition-colors">
                    Xóa tài khoản
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
