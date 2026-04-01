import { Search } from "lucide-react";
import NotificationsBell from "./NotificationsBell";

export default function AppHeader({ user, searchPlaceholder = "Tìm kiếm..." }) {
  return (
    <header className="h-16 border-b border-primary/10 bg-white dark:bg-slate-900 flex items-center justify-between px-8">
      <div className="flex items-center gap-4 flex-1 max-w-xl">
        <div className="relative w-full">
          <Search
            size={18}
            className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
          />
          <input
            className="w-full pl-10 pr-4 py-2 bg-slate-50 dark:bg-slate-800 border-none rounded-lg focus:ring-1 focus:ring-primary text-sm"
            placeholder={searchPlaceholder}
            type="text"
          />
        </div>
      </div>
      <div className="flex items-center gap-4">
        <NotificationsBell />
        <div className="h-8 w-px bg-primary/10 mx-2"></div>
        <div className="flex items-center gap-3">
          <div className="text-right hidden sm:block">
            <p className="text-sm font-semibold leading-none">{user?.name || "User"}</p>
            <p className="text-xs text-slate-500">{user?.email || ""}</p>
          </div>
          <div
            className="w-10 h-10 rounded-full bg-slate-200 dark:bg-slate-700 border-2 border-primary/20 bg-cover bg-center"
            style={{
              backgroundImage: user?.avatar ? `url(${user.avatar})` : "none",
            }}
          ></div>
        </div>
      </div>
    </header>
  );
}
