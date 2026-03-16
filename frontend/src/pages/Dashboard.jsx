import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Shield, LayoutDashboard, FolderOpen, UserCheck, Settings,
  Search, Bell, Lock, Database, ShieldAlert, FileText, Image,
  FolderArchive, CloudUpload, Radar, ShieldCheck
} from 'lucide-react';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:6868/api/v1';

export default function Dashboard() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    fetchUser();
  }, []);

  const fetchUser = async () => {
    const token = localStorage.getItem('aeshield_token');

    if (!token) {
      navigate('/', { replace: true });
      return;
    }

    try {
      const response = await fetch(`${API_BASE_URL}/auth/me`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (response.ok) {
        const userData = await response.json();
        setUser(userData);
      } else {
        localStorage.removeItem('aeshield_token');
        localStorage.removeItem('aeshield_user');
        navigate('/', { replace: true });
      }
    } catch (error) {
      console.error('Error fetching user:', error);
      localStorage.removeItem('aeshield_token');
      localStorage.removeItem('aeshield_user');
      navigate('/', { replace: true });
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('aeshield_token');
    localStorage.removeItem('aeshield_user');
    navigate('/', { replace: true });
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
              <p className="text-xs text-primary font-medium">Bảo mật Doanh nghiệp</p>
            </div>
          </div>
        </div>
        <nav className="flex-1 px-4 space-y-1">
          <a className="flex items-center gap-3 px-3 py-2 rounded-lg bg-primary/10 text-primary font-medium" href="/dashboard">
            <LayoutDashboard size={20} />
            <span>Dashboard</span>
          </a>
          <a className="flex items-center gap-3 px-3 py-2 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors" href="/files">
            <FolderOpen size={20} />
            <span>Tệp tin</span>
          </a>
          <a className="flex items-center gap-3 px-3 py-2 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors" href="/whitelist">
            <UserCheck size={20} />
            <span>Danh sách tin cậy</span>
          </a>
          <a className="flex items-center gap-3 px-3 py-2 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors" href="/settings">
            <Settings size={20} />
            <span>Cài đặt tài khoản</span>
          </a>
        </nav>
        <div className="p-4 border-t border-primary/10">
          <div className="bg-primary/5 rounded-xl p-4">
            <p className="text-xs font-semibold text-slate-500 mb-2 uppercase tracking-wider">Dung lượng lưu trữ</p>
            <div className="w-full bg-slate-200 dark:bg-slate-700 h-1.5 rounded-full mb-2">
              <div className="bg-primary h-1.5 rounded-full" style={{width: "64%"}}></div>
            </div>
            <p className="text-xs text-slate-600 dark:text-slate-400">64.2 GB / 100 GB đã sử dụng</p>
          </div>
        </div>
      </aside>

      {/* Main Content Area */}
      <main className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <header className="h-16 border-b border-primary/10 bg-white dark:bg-slate-900 flex items-center justify-between px-8">
          <div className="flex items-center gap-4 flex-1 max-w-xl">
            <div className="relative w-full">
              <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
              <input
                className="w-full pl-10 pr-4 py-2 bg-slate-50 dark:bg-slate-800 border-none rounded-lg focus:ring-1 focus:ring-primary text-sm"
                placeholder="Tìm kiếm tệp tin đã mã hóa..."
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
                <p className="text-sm font-semibold leading-none">{user?.name || 'User'}</p>
                <p className="text-xs text-slate-500">{user?.email || ''}</p>
              </div>
              <div
                className="w-10 h-10 rounded-full bg-slate-200 dark:bg-slate-700 border-2 border-primary/20 bg-cover bg-center"
                style={{
                  backgroundImage: user?.avatar ? `url(${user.avatar})` : 'none'
                }}
              ></div>
            </div>
          </div>
        </header>

        {/* Scrollable Content */}
        <div className="flex-1 overflow-y-auto p-8">
          <div className="max-w-6xl mx-auto space-y-8">
            {/* Stats Grid */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="bg-white dark:bg-slate-900 p-6 rounded-xl border border-primary/10 shadow-sm">
                <div className="flex justify-between items-start mb-4">
                  <div className="p-2 bg-primary/10 rounded-lg text-primary">
                    <Lock size={22} />
                  </div>
                  <span className="text-emerald-500 text-xs font-bold bg-emerald-500/10 px-2 py-1 rounded">+12%</span>
                </div>
                <h3 className="text-slate-500 text-sm font-medium">Tổng tệp tin đã mã hóa</h3>
                <p className="text-3xl font-bold mt-1">12,482</p>
              </div>
              <div className="bg-white dark:bg-slate-900 p-6 rounded-xl border border-primary/10 shadow-sm">
                <div className="flex justify-between items-start mb-4">
                  <div className="p-2 bg-primary/10 rounded-lg text-primary">
                    <Database size={22} />
                  </div>
                  <span className="text-emerald-500 text-xs font-bold bg-emerald-500/10 px-2 py-1 rounded">+5%</span>
                </div>
                <h3 className="text-slate-500 text-sm font-medium">Dung lượng đã sử dụng</h3>
                <p className="text-3xl font-bold mt-1">64.2 GB</p>
              </div>
              <div className="bg-white dark:bg-slate-900 p-6 rounded-xl border border-primary/10 shadow-sm">
                <div className="flex justify-between items-start mb-4">
                  <div className="p-2 bg-primary/10 rounded-lg text-primary">
                    <ShieldAlert size={22} />
                  </div>
                  <span className="text-slate-400 text-xs font-bold bg-slate-400/10 px-2 py-1 rounded">Bảo mật</span>
                </div>
                <h3 className="text-slate-500 text-sm font-medium">Mối đe dọa bị chặn</h3>
                <p className="text-3xl font-bold mt-1">0</p>
              </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
              {/* Recent Activity Table */}
              <div className="lg:col-span-2 bg-white dark:bg-slate-900 rounded-xl border border-primary/10 shadow-sm overflow-hidden">
                <div className="p-6 border-b border-primary/10 flex items-center justify-between">
                  <h3 className="font-bold text-lg">Hoạt động gần đây</h3>
                  <button className="text-primary text-sm font-semibold hover:underline">Xem tất cả</button>
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-left">
                    <thead className="bg-slate-50 dark:bg-slate-800/50 text-slate-500 text-xs uppercase tracking-wider">
                      <tr>
                        <th className="px-6 py-4 font-semibold">Tên tệp</th>
                        <th className="px-6 py-4 font-semibold">Hành động</th>
                        <th className="px-6 py-4 font-semibold">Trạng thái</th>
                        <th className="px-6 py-4 font-semibold">Thời gian</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-primary/5">
                      <tr className="hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
                        <td className="px-6 py-4 font-medium flex items-center gap-2">
                          <FileText size={18} className="text-slate-400" />
                          bao_cao_nam_2023.pdf
                        </td>
                        <td className="px-6 py-4 text-sm">Mã hóa</td>
                        <td className="px-6 py-4">
                          <span className="px-2 py-1 bg-emerald-500/10 text-emerald-500 text-xs rounded-full font-medium">Hoàn thành</span>
                        </td>
                        <td className="px-6 py-4 text-sm text-slate-500">2 phút trước</td>
                      </tr>
                      <tr className="hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
                        <td className="px-6 py-4 font-medium flex items-center gap-2">
                          <Image size={18} className="text-slate-400" />
                          sao_luu_server.iso
                        </td>
                        <td className="px-6 py-4 text-sm">Tải lên Cloud</td>
                        <td className="px-6 py-4">
                          <span className="px-2 py-1 bg-blue-500/10 text-blue-500 text-xs rounded-full font-medium">Đang xử lý</span>
                        </td>
                        <td className="px-6 py-4 text-sm text-slate-500">14 phút trước</td>
                      </tr>
                      <tr className="hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
                        <td className="px-6 py-4 font-medium flex items-center gap-2">
                          <Lock size={18} className="text-slate-400" />
                          thong_tin_dang_nhap.txt
                        </td>
                        <td className="px-6 py-4 text-sm">Xoay khóa</td>
                        <td className="px-6 py-4">
                          <span className="px-2 py-1 bg-emerald-500/10 text-emerald-500 text-xs rounded-full font-medium">Hoàn thành</span>
                        </td>
                        <td className="px-6 py-4 text-sm text-slate-500">1 giờ trước</td>
                      </tr>
                      <tr className="hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
                        <td className="px-6 py-4 font-medium flex items-center gap-2">
                          <FolderArchive size={18} className="text-slate-400" />
                          ma_nguon_du_an.zip
                        </td>
                        <td className="px-6 py-4 text-sm">Quét lỗ hổng</td>
                        <td className="px-6 py-4">
                          <span className="px-2 py-1 bg-emerald-500/10 text-emerald-500 text-xs rounded-full font-medium">Hoàn thành</span>
                        </td>
                        <td className="px-6 py-4 text-sm text-slate-500">3 giờ trước</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              {/* Quick Actions */}
              <div className="space-y-6">
                <h3 className="font-bold text-lg">Thao tác nhanh</h3>
                <div className="group bg-white dark:bg-slate-900 p-6 rounded-xl border border-primary/10 shadow-sm hover:border-primary transition-all cursor-pointer">
                  <div className="flex items-center gap-4">
                    <div className="p-3 bg-primary text-white rounded-xl">
                      <CloudUpload size={22} />
                    </div>
                    <div>
                      <h4 className="font-bold">Tải lên tệp mới</h4>
                      <p className="text-sm text-slate-500">Mã hóa và lưu trữ an toàn</p>
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
                      <p className="text-sm text-slate-500">Quét sâu lỗ hổng bảo mật</p>
                    </div>
                  </div>
                </div>
                {/* Additional Feature Card */}
                <div className="relative overflow-hidden bg-primary/10 p-6 rounded-xl border border-primary/20">
                  <div className="relative z-10">
                    <h4 className="font-bold text-primary mb-1">Kiểm tra bảo mật hàng tuần</h4>
                    <p className="text-sm text-slate-600 dark:text-slate-400 mb-4">Hệ thống của bạn được bảo vệ 98% tuần này.</p>
                    <button className="bg-white dark:bg-slate-900 text-slate-900 dark:text-white px-4 py-2 rounded-lg text-sm font-bold shadow-sm">Xem báo cáo</button>
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
