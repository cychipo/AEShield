import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { ShieldCheck, Lock, UserCheck, KeyRound } from "lucide-react";

const API_BASE_URL =
  import.meta.env.VITE_API_URL || "http://localhost:6888/api/v1";

export default function Login() {
  const navigate = useNavigate();

  useEffect(() => {
    checkExistingToken();
  }, []);

  const checkExistingToken = () => {
    const token = localStorage.getItem("aeshield_token");
    if (token) {
      navigate("/dashboard", { replace: true });
    }
  };

  const handleLogin = async (provider) => {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/urls`);
      const data = await response.json();

      if (data[provider]) {
        window.location.href = data[provider];
      } else {
        throw new Error("Failed to get authorization URL");
      }
    } catch (err) {
      console.error("Login error:", err);
      alert("Có lỗi xảy ra. Vui lòng thử lại.");
    }
  };

  return (
    <div className="bg-background-light dark:bg-background-dark min-h-screen flex items-center justify-center p-6 font-display">
      <div className="w-full max-w-md">
        {/* Brand Identity */}
        <div className="flex flex-col items-center mb-10">
          <div className="bg-primary/10 p-3 rounded-xl mb-4">
            <ShieldCheck size={48} className="text-primary" />
          </div>
          <h1 className="text-3xl font-bold text-charcoal dark:text-slate-100 tracking-tight">
            AEShield
          </h1>
          <p className="text-charcoal/60 dark:text-slate-400 mt-2 text-center text-sm">
            Cổng bảo mật cấp doanh nghiệp
          </p>
        </div>

        {/* Login Card */}
        <div className="bg-surface dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-xl shadow-sm p-8">
          <div className="mb-8">
            <h2 className="text-xl font-semibold text-charcoal dark:text-slate-100">
              Đăng nhập
            </h2>
            <p className="text-charcoal/60 dark:text-slate-400 text-sm mt-1">
              Xác thực để truy cập bảng điều khiển an toàn
            </p>
          </div>

          <div className="flex flex-col gap-4">
            {/* Google Provider */}
            <button
              onClick={() => handleLogin("google")}
              className="flex items-center justify-center gap-3 w-full h-12 px-5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors text-charcoal dark:text-slate-200 font-medium"
            >
              <svg className="w-5 h-5" viewBox="0 0 24 24">
                <path
                  d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                  fill="#4285F4"
                ></path>
                <path
                  d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                  fill="#34A853"
                ></path>
                <path
                  d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z"
                  fill="#FBBC05"
                ></path>
                <path
                  d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 12-4.53z"
                  fill="#EA4335"
                ></path>
              </svg>
              <span>Tiếp tục với Google</span>
            </button>

            {/* GitHub Provider */}
            <button
              onClick={() => handleLogin("github")}
              className="flex items-center justify-center gap-3 w-full h-12 px-5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors text-charcoal dark:text-slate-200 font-medium"
            >
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"></path>
              </svg>
              <span>Tiếp tục với GitHub</span>
            </button>
          </div>

          {/* Visual Separator */}
          <div className="mt-8 pt-6 border-t border-slate-100 dark:border-slate-800 flex items-center justify-center gap-2 text-charcoal/40 dark:text-slate-500 text-xs uppercase tracking-widest font-semibold">
            <Lock size={14} />
            Bảo vệ bởi AEShield
          </div>
        </div>

        {/* Footer Info */}
        <div className="mt-8 text-center space-y-4">
          <div className="flex items-center justify-center gap-6">
            <div className="flex items-center gap-1.5 text-charcoal/50 dark:text-slate-500 text-xs">
              <UserCheck size={14} />
              Tuân thủ SOC2
            </div>
            <div className="flex items-center gap-1.5 text-charcoal/50 dark:text-slate-500 text-xs">
              <KeyRound size={14} />
              Mã hóa đầu cuối
            </div>
          </div>
          <p className="text-charcoal/40 dark:text-slate-600 text-[11px] leading-relaxed max-w-[280px] mx-auto">
            Bằng việc đăng nhập, bạn đồng ý với{" "}
            <a href="/terms" className="text-primary hover:underline">
              điều khoản dịch vụ
            </a>{" "}
            của chúng tôi.
          </p>
        </div>
      </div>
    </div>
  );
}
