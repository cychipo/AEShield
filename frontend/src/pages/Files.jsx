import { useState, useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import {
  Shield,
  LayoutDashboard,
  FolderOpen,
  UserCheck,
  Settings,
  Search,
  Bell,
  CloudUpload,
  Download,
  Eye,
  Trash2,
} from "lucide-react";
import {
  decryptEncryptedFile,
  detectPreviewType,
} from "../utils/encryptedPreview";

const API_BASE_URL =
  import.meta.env.VITE_API_URL || "http://localhost:6888/api/v1";

export default function Files() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [showUploadForm, setShowUploadForm] = useState(false);
  const [selectedFile, setSelectedFile] = useState(null);
  const [password, setPassword] = useState("");
  const [encryptionType, setEncryptionType] = useState("AES-256");
  const [accessMode, setAccessMode] = useState("private");
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState("");
  const [uploadSuccess, setUploadSuccess] = useState("");
  const [lastUploadedFile, setLastUploadedFile] = useState(null);
  const [files, setFiles] = useState([]);
  const [filesLoading, setFilesLoading] = useState(false);
  const [filesError, setFilesError] = useState("");
  const [downloadingFileId, setDownloadingFileId] = useState("");
  const [deletingFileId, setDeletingFileId] = useState("");
  const [previewingFileId, setPreviewingFileId] = useState("");
  const [showPreview, setShowPreview] = useState(false);
  const [previewFile, setPreviewFile] = useState(null);
  const [previewPassword, setPreviewPassword] = useState("");
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewStage, setPreviewStage] = useState("");
  const [previewError, setPreviewError] = useState("");
  const [previewType, setPreviewType] = useState("unsupported");
  const [previewText, setPreviewText] = useState("");
  const [previewObjectUrl, setPreviewObjectUrl] = useState("");
  const fileInputRef = useRef(null);
  const previewObjectUrlRef = useRef("");
  const navigate = useNavigate();

  useEffect(() => {
    fetchUser();
  }, []);

  useEffect(() => {
    return () => {
      if (previewObjectUrlRef.current) {
        URL.revokeObjectURL(previewObjectUrlRef.current);
      }
    };
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
        await fetchFiles(token);
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

  const fetchFiles = async (token) => {
    setFilesLoading(true);
    setFilesError("");

    try {
      const response = await fetch(`${API_BASE_URL}/files`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        if (response.status === 401) {
          localStorage.removeItem("aeshield_token");
          localStorage.removeItem("aeshield_user");
          navigate("/", { replace: true });
          return;
        }

        let errorPayload = null;
        try {
          errorPayload = await response.json();
        } catch {
          errorPayload = null;
        }

        setFilesError(
          errorPayload?.error || "Không thể tải danh sách tệp tin."
        );
        return;
      }

      const filesPayload = await response.json();
      setFiles(Array.isArray(filesPayload) ? filesPayload : []);
    } catch (error) {
      console.error("Error fetching files:", error);
      setFilesError("Có lỗi xảy ra khi tải danh sách tệp tin.");
    } finally {
      setFilesLoading(false);
    }
  };

  const handleDownload = async (fileId) => {
    const token = localStorage.getItem("aeshield_token");
    if (!token) {
      navigate("/", { replace: true });
      return;
    }

    setFilesError("");
    setDownloadingFileId(fileId);

    try {
      const response = await fetch(`${API_BASE_URL}/files/${fileId}/download`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      let payload = null;
      try {
        payload = await response.json();
      } catch {
        payload = null;
      }

      if (!response.ok) {
        if (response.status === 401) {
          localStorage.removeItem("aeshield_token");
          localStorage.removeItem("aeshield_user");
          navigate("/", { replace: true });
          return;
        }

        setFilesError(payload?.error || "Không thể tải tệp tin.");
        return;
      }

      if (!payload?.url) {
        setFilesError("Không nhận được đường dẫn tải tệp tin.");
        return;
      }

      window.open(payload.url, "_blank", "noopener,noreferrer");
    } catch (error) {
      console.error("Error downloading file:", error);
      setFilesError("Có lỗi xảy ra khi tải tệp tin.");
    } finally {
      setDownloadingFileId("");
    }
  };

  const handleDelete = async (fileId, filename) => {
    const token = localStorage.getItem("aeshield_token");
    if (!token) {
      navigate("/", { replace: true });
      return;
    }

    const confirmed = window.confirm(
      `Bạn có chắc chắn muốn xóa tệp \"${filename}\" không?`
    );
    if (!confirmed) {
      return;
    }

    setFilesError("");
    setDeletingFileId(fileId);

    try {
      const response = await fetch(`${API_BASE_URL}/files/${fileId}`, {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      let payload = null;
      try {
        payload = await response.json();
      } catch {
        payload = null;
      }

      if (!response.ok) {
        if (response.status === 401) {
          localStorage.removeItem("aeshield_token");
          localStorage.removeItem("aeshield_user");
          navigate("/", { replace: true });
          return;
        }

        setFilesError(payload?.error || "Không thể xóa tệp tin.");
        return;
      }

      setFiles((prevFiles) => prevFiles.filter((file) => file.id !== fileId));
      await fetchFiles(token);
    } catch (error) {
      console.error("Error deleting file:", error);
      setFilesError("Có lỗi xảy ra khi xóa tệp tin.");
    } finally {
      setDeletingFileId("");
    }
  };

  const clearPreviewObjectUrl = () => {
    if (previewObjectUrlRef.current) {
      URL.revokeObjectURL(previewObjectUrlRef.current);
      previewObjectUrlRef.current = "";
    }

    setPreviewObjectUrl("");
  };

  const closePreview = () => {
    if (previewLoading) {
      return;
    }

    setShowPreview(false);
    setPreviewingFileId("");
    setPreviewFile(null);
    setPreviewPassword("");
    setPreviewLoading(false);
    setPreviewStage("");
    setPreviewError("");
    setPreviewType("unsupported");
    setPreviewText("");
    clearPreviewObjectUrl();
  };

  const openPreview = (file) => {
    const rowFileId = file?.id;
    if (!rowFileId) {
      return;
    }

    const detectedType = detectPreviewType(file?.filename || "");

    setShowPreview(true);
    setPreviewingFileId(rowFileId);
    setPreviewFile(file);
    setPreviewPassword("");
    setPreviewLoading(false);
    setPreviewStage("");
    setPreviewError("");
    setPreviewType(detectedType.type);
    setPreviewText("");
    clearPreviewObjectUrl();
  };

  const handleDecryptPreview = async (event) => {
    event.preventDefault();

    const fileId = previewFile?.id;
    if (!fileId) {
      setPreviewError("Không tìm thấy tệp tin để xem trước.");
      setPreviewPassword("");
      return;
    }

    const detectedType = detectPreviewType(previewFile?.filename || "");
    if (!detectedType.supported) {
      setPreviewError("Định dạng tệp tin này chưa hỗ trợ xem trước.");
      setPreviewPassword("");
      return;
    }

    const inputPassword = previewPassword;
    if (!inputPassword.trim()) {
      setPreviewError("Vui lòng nhập mật khẩu để giải mã xem trước.");
      return;
    }

    const token = localStorage.getItem("aeshield_token");
    if (!token) {
      setPreviewPassword("");
      navigate("/", { replace: true });
      return;
    }

    setPreviewLoading(true);
    setPreviewStage("Đang lấy liên kết tải...");
    setPreviewError("");
    setPreviewText("");
    clearPreviewObjectUrl();

    try {
      const response = await fetch(`${API_BASE_URL}/files/${fileId}/download`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      let payload = null;
      try {
        payload = await response.json();
      } catch {
        payload = null;
      }

      if (!response.ok) {
        if (response.status === 401) {
          localStorage.removeItem("aeshield_token");
          localStorage.removeItem("aeshield_user");
          navigate("/", { replace: true });
          return;
        }

        setPreviewError(payload?.error || "Không thể lấy tệp tin để xem trước.");
        return;
      }

      if (!payload?.url) {
        setPreviewError("Không nhận được đường dẫn tải tệp tin.");
        return;
      }

      setPreviewStage("Đang tải dữ liệu mã hóa...");
      const encryptedResponse = await fetch(payload.url);
      if (!encryptedResponse.ok) {
        setPreviewError("Không thể tải dữ liệu tệp tin.");
        return;
      }

      const encryptedBuffer = await encryptedResponse.arrayBuffer();

      setPreviewStage("Đang giải mã tệp tin...");
      const plaintextBytes = await decryptEncryptedFile(
        new Uint8Array(encryptedBuffer),
        inputPassword
      );

      setPreviewStage("Đang chuẩn bị hiển thị...");
      setPreviewType(detectedType.type);

      if (detectedType.type === "text") {
        const decodedText = new TextDecoder("utf-8").decode(plaintextBytes);
        setPreviewText(decodedText);
      } else {
        const blob = new Blob([plaintextBytes], { type: detectedType.mimeType });
        const nextObjectUrl = URL.createObjectURL(blob);
        previewObjectUrlRef.current = nextObjectUrl;
        setPreviewObjectUrl(nextObjectUrl);
      }

      setPreviewError("");
    } catch (error) {
      setPreviewText("");
      clearPreviewObjectUrl();
      setPreviewError(
        error instanceof Error && error.message
          ? error.message
          : "Có lỗi xảy ra khi giải mã xem trước."
      );
    } finally {
      setPreviewLoading(false);
      setPreviewStage("");
      setPreviewPassword("");
    }
  };

  const openUploadForm = () => {
    setShowUploadForm(true);
    setUploadError("");
    setUploadSuccess("");
  };

  const closeUploadForm = () => {
    if (uploading) {
      return;
    }

    setShowUploadForm(false);
    setSelectedFile(null);
    setPassword("");
    setEncryptionType("AES-256");
    setAccessMode("private");
    setUploadError("");
    setUploadSuccess("");

    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  const handleFileChange = (event) => {
    const file = event.target.files?.[0] || null;
    setSelectedFile(file);
    setUploadError("");
    setUploadSuccess("");
  };

  const handleUpload = async (event) => {
    event.preventDefault();
    setUploadError("");
    setUploadSuccess("");

    const token = localStorage.getItem("aeshield_token");
    if (!token) {
      setPassword("");
      navigate("/", { replace: true });
      return;
    }

    if (!selectedFile) {
      setUploadError("Vui lòng chọn tệp tin cần tải lên.");
      return;
    }

    if (!password.trim()) {
      setUploadError("Vui lòng nhập mật khẩu mã hóa.");
      return;
    }

    setUploading(true);

    try {
      const formData = new FormData();
      formData.append("file", selectedFile);
      formData.append("password", password);
      formData.append("encryption_type", encryptionType);
      formData.append("access_mode", accessMode);

      const response = await fetch(`${API_BASE_URL}/files/upload`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      });

      let payload = null;
      try {
        payload = await response.json();
      } catch {
        payload = null;
      }

      if (!response.ok) {
        if (response.status === 401) {
          localStorage.removeItem("aeshield_token");
          localStorage.removeItem("aeshield_user");
          navigate("/", { replace: true });
          return;
        }

        setUploadError(payload?.error || "Tải lên thất bại. Vui lòng thử lại.");
        return;
      }

      setLastUploadedFile(payload);
      setUploadSuccess(
        `Tải lên thành công: ${payload?.filename || selectedFile.name}`
      );
      setSelectedFile(null);
      await fetchFiles(token);

      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    } catch (error) {
      console.error("Error uploading file:", error);
      setUploadError("Có lỗi xảy ra khi tải lên tệp tin.");
    } finally {
      setUploading(false);
      setPassword("");
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
            className="flex items-center gap-3 px-3 py-2 rounded-lg bg-primary/10 text-primary font-medium"
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
            className="flex items-center gap-3 px-3 py-2 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
            href="/settings"
          >
            <Settings size={20} />
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
                placeholder="Tìm kiếm tệp tin..."
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
          <div className="max-w-6xl mx-auto">
            <div className="flex justify-between items-center mb-8">
              <h1 className="text-2xl font-bold">Tệp tin của bạn</h1>
              <button
                className="bg-primary text-white px-4 py-2 rounded-lg font-medium hover:bg-orange-600 transition-colors"
                onClick={openUploadForm}
                type="button"
              >
                Tải lên tệp mới
              </button>
            </div>

            {showUploadForm && (
              <div className="bg-white dark:bg-slate-900 rounded-xl border border-primary/10 shadow-sm p-6 mb-6">
                <div className="flex items-center gap-2 mb-4">
                  <CloudUpload size={18} className="text-primary" />
                  <h2 className="text-lg font-semibold">Tải lên tệp mới</h2>
                </div>

                <form className="space-y-4" onSubmit={handleUpload}>
                  <div>
                    <label className="block text-sm font-medium mb-2" htmlFor="upload-file">
                      Tệp tin
                    </label>
                    <input
                      className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-800 border border-primary/20 rounded-lg text-sm"
                      disabled={uploading}
                      id="upload-file"
                      onChange={handleFileChange}
                      ref={fileInputRef}
                      required
                      type="file"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-2" htmlFor="upload-password">
                      Mật khẩu mã hóa
                    </label>
                    <input
                      autoComplete="new-password"
                      className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-800 border border-primary/20 rounded-lg text-sm"
                      disabled={uploading}
                      id="upload-password"
                      onChange={(event) => setPassword(event.target.value)}
                      placeholder="Nhập mật khẩu để mã hóa"
                      required
                      type="password"
                      value={password}
                    />
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium mb-2" htmlFor="upload-encryption-type">
                        Loại mã hóa
                      </label>
                      <select
                        className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-800 border border-primary/20 rounded-lg text-sm"
                        disabled={uploading}
                        id="upload-encryption-type"
                        onChange={(event) => setEncryptionType(event.target.value)}
                        value={encryptionType}
                      >
                        <option value="AES-128">AES-128</option>
                        <option value="AES-192">AES-192</option>
                        <option value="AES-256">AES-256</option>
                      </select>
                    </div>

                    <div>
                      <label className="block text-sm font-medium mb-2" htmlFor="upload-access-mode">
                        Chế độ truy cập
                      </label>
                      <select
                        className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-800 border border-primary/20 rounded-lg text-sm"
                        disabled={uploading}
                        id="upload-access-mode"
                        onChange={(event) => setAccessMode(event.target.value)}
                        value={accessMode}
                      >
                        <option value="private">private</option>
                        <option value="public">public</option>
                        <option value="whitelist">whitelist</option>
                      </select>
                    </div>
                  </div>

                  {uploadError && (
                    <div className="rounded-lg border border-red-200 bg-red-50 text-red-700 px-3 py-2 text-sm dark:bg-red-950/30 dark:border-red-800 dark:text-red-300">
                      {uploadError}
                    </div>
                  )}

                  <div className="flex items-center gap-3">
                    <button
                      className="bg-primary text-white px-4 py-2 rounded-lg font-medium hover:bg-orange-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                      disabled={uploading}
                      type="submit"
                    >
                      {uploading ? "Đang tải lên..." : "Tải lên"}
                    </button>
                    <button
                      className="px-4 py-2 rounded-lg font-medium border border-primary/20 hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed"
                      disabled={uploading}
                      onClick={closeUploadForm}
                      type="button"
                    >
                      Hủy
                    </button>
                  </div>
                </form>
              </div>
            )}

            {uploadSuccess && (
              <div className="mb-6 rounded-xl border border-green-200 bg-green-50 text-green-700 px-4 py-3 text-sm dark:bg-green-950/30 dark:border-green-800 dark:text-green-300">
                {uploadSuccess}
              </div>
            )}

            {lastUploadedFile && (
              <div className="mb-6 bg-white dark:bg-slate-900 rounded-xl border border-primary/10 shadow-sm p-4">
                <p className="text-sm text-slate-600 dark:text-slate-300">
                  <span className="font-semibold">Tệp vừa tải lên:</span>{" "}
                  {lastUploadedFile.filename}
                </p>
                <p className="text-sm text-slate-500 dark:text-slate-400 mt-1">
                  Mã hóa: {lastUploadedFile.encryption_type} • Truy cập: {lastUploadedFile.access_mode}
                </p>
              </div>
            )}

            {filesError && (
              <div className="mb-6 rounded-xl border border-red-200 bg-red-50 text-red-700 px-4 py-3 text-sm dark:bg-red-950/30 dark:border-red-800 dark:text-red-300">
                {filesError}
              </div>
            )}

            {filesLoading ? (
              <div className="bg-white dark:bg-slate-900 rounded-xl border border-primary/10 shadow-sm p-6 text-center">
                <p className="text-slate-500 dark:text-slate-400">Đang tải danh sách tệp tin...</p>
              </div>
            ) : files.length === 0 ? (
              <div className="bg-white dark:bg-slate-900 rounded-xl border border-primary/10 shadow-sm p-6 text-center">
                <FolderOpen size={48} className="text-slate-400 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-slate-700 dark:text-slate-300 mb-2">
                  Chưa có tệp tin nào
                </h3>
                <p className="text-slate-500 dark:text-slate-400 mb-4">
                  Tải lên tệp tin đầu tiên của bạn để bắt đầu mã hóa và bảo vệ dữ
                  liệu.
                </p>
                <button
                  className="bg-primary text-white px-6 py-2 rounded-lg font-medium hover:bg-orange-600 transition-colors"
                  onClick={openUploadForm}
                  type="button"
                >
                  Tải lên ngay
                </button>
              </div>
            ) : (
              <div className="bg-white dark:bg-slate-900 rounded-xl border border-primary/10 shadow-sm overflow-hidden">
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead className="bg-slate-50 dark:bg-slate-800/50">
                      <tr>
                        <th className="text-left font-medium px-4 py-3">Tên tệp</th>
                        <th className="text-left font-medium px-4 py-3">Kích thước</th>
                        <th className="text-left font-medium px-4 py-3">Mã hóa</th>
                        <th className="text-left font-medium px-4 py-3">Truy cập</th>
                        <th className="text-left font-medium px-4 py-3">Cập nhật</th>
                        <th className="text-left font-medium px-4 py-3">Thao tác</th>
                      </tr>
                    </thead>
                    <tbody>
                      {files.map((file) => {
                        const rowFileId = file.id;
                        const isDownloading = downloadingFileId === rowFileId;
                        const isDeleting = deletingFileId === rowFileId;
                        const isPreviewing =
                          showPreview && previewingFileId === rowFileId;

                        return (
                          <tr
                            className="border-t border-primary/10"
                            key={file.id || file.storage_path || file.filename}
                          >
                            <td className="px-4 py-3 font-medium">{file.filename}</td>
                            <td className="px-4 py-3 text-slate-500 dark:text-slate-400">
                              {typeof file.size === "number"
                                ? `${(file.size / 1024 / 1024).toFixed(2)} MB`
                                : "-"}
                            </td>
                            <td className="px-4 py-3 text-slate-600 dark:text-slate-300">
                              {file.encryption_type || "-"}
                            </td>
                            <td className="px-4 py-3 text-slate-600 dark:text-slate-300">
                              {file.access_mode || "-"}
                            </td>
                            <td className="px-4 py-3 text-slate-500 dark:text-slate-400">
                              {file.updated_at
                                ? new Date(file.updated_at).toLocaleString("vi-VN")
                                : "-"}
                            </td>
                            <td className="px-4 py-3">
                              <div className="flex items-center gap-2">
                                <button
                                  className="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg border border-primary/20 text-xs font-medium hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed"
                                  disabled={!rowFileId || isPreviewing || isDownloading || isDeleting}
                                  onClick={() => openPreview(file)}
                                  type="button"
                                >
                                  <Eye size={14} />
                                  {isPreviewing ? "Đang xem..." : "Preview"}
                                </button>
                                <button
                                  className="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg border border-primary/20 text-xs font-medium hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed"
                                  disabled={!rowFileId || isDownloading || isDeleting || isPreviewing}
                                  onClick={() => handleDownload(rowFileId)}
                                  type="button"
                                >
                                  <Download size={14} />
                                  {isDownloading ? "Đang lấy link..." : "Tải xuống"}
                                </button>
                                <button
                                  className="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg border border-red-300 text-red-600 text-xs font-medium hover:bg-red-50 dark:border-red-700 dark:text-red-300 dark:hover:bg-red-950/30 disabled:opacity-50 disabled:cursor-not-allowed"
                                  disabled={!rowFileId || isDeleting || isDownloading || isPreviewing}
                                  onClick={() =>
                                    handleDelete(rowFileId, file.filename || "tệp tin này")
                                  }
                                  type="button"
                                >
                                  <Trash2 size={14} />
                                  {isDeleting ? "Đang xóa..." : "Xóa"}
                                </button>
                              </div>
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
        </div>
      </main>

      {showPreview && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-4xl max-h-[90vh] overflow-hidden rounded-xl border border-primary/20 bg-white dark:bg-slate-900 shadow-xl flex flex-col">
            <div className="px-5 py-4 border-b border-primary/10 flex items-center justify-between gap-4">
              <div className="min-w-0">
                <h2 className="text-lg font-semibold">Xem trước tệp mã hóa</h2>
                <p className="text-sm text-slate-500 dark:text-slate-400 truncate">
                  {previewFile?.filename || "-"}
                </p>
              </div>
              <button
                className="px-3 py-1.5 rounded-lg border border-primary/20 text-sm hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed"
                disabled={previewLoading}
                onClick={closePreview}
                type="button"
              >
                Đóng
              </button>
            </div>

            <div className="p-5 overflow-y-auto space-y-4">
              <form className="space-y-3" onSubmit={handleDecryptPreview}>
                <div>
                  <label className="block text-sm font-medium mb-2" htmlFor="preview-password">
                    Mật khẩu giải mã
                  </label>
                  <input
                    autoComplete="off"
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-800 border border-primary/20 rounded-lg text-sm"
                    disabled={previewLoading}
                    id="preview-password"
                    onChange={(event) => setPreviewPassword(event.target.value)}
                    placeholder="Nhập mật khẩu để xem trước"
                    type="password"
                    value={previewPassword}
                  />
                </div>

                <div className="flex items-center gap-3">
                  <button
                    className="bg-primary text-white px-4 py-2 rounded-lg font-medium hover:bg-orange-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                    disabled={previewLoading || !previewFile?.id}
                    type="submit"
                  >
                    {previewLoading ? "Đang giải mã..." : "Giải mã & xem trước"}
                  </button>
                  <span className="text-xs text-slate-500 dark:text-slate-400">
                    Hỗ trợ: text, image, pdf
                  </span>
                </div>
              </form>

              {previewLoading && previewStage && (
                <div className="rounded-lg border border-primary/20 bg-primary/5 text-primary px-3 py-2 text-sm">
                  {previewStage}
                </div>
              )}

              {previewError && (
                <div className="rounded-lg border border-red-200 bg-red-50 text-red-700 px-3 py-2 text-sm dark:bg-red-950/30 dark:border-red-800 dark:text-red-300">
                  {previewError}
                </div>
              )}

              {!previewLoading && !previewError && previewType === "unsupported" && (
                <div className="rounded-lg border border-yellow-200 bg-yellow-50 text-yellow-700 px-3 py-2 text-sm dark:bg-yellow-950/30 dark:border-yellow-800 dark:text-yellow-300">
                  Định dạng tệp tin này chưa hỗ trợ xem trước. Bạn vẫn có thể tải xuống
                  để giải mã cục bộ.
                </div>
              )}

              {!previewLoading &&
                !previewError &&
                previewType !== "unsupported" &&
                !previewText &&
                !previewObjectUrl && (
                  <div className="rounded-lg border border-primary/20 bg-slate-50 dark:bg-slate-800/60 text-slate-600 dark:text-slate-300 px-3 py-2 text-sm">
                    Nhập mật khẩu và bấm "Giải mã & xem trước" để hiển thị nội dung.
                  </div>
                )}

              {!previewLoading &&
                !previewError &&
                previewType === "text" &&
                previewText && (
                  <pre className="rounded-lg border border-primary/10 bg-slate-50 dark:bg-slate-800/50 p-4 text-sm whitespace-pre-wrap break-words max-h-[55vh] overflow-auto">
                    {previewText}
                  </pre>
                )}

              {!previewLoading &&
                !previewError &&
                previewType === "image" &&
                previewObjectUrl && (
                  <div className="rounded-lg border border-primary/10 p-3 bg-slate-50 dark:bg-slate-800/50 max-h-[60vh] overflow-auto">
                    <img
                      alt={previewFile?.filename || "preview-image"}
                      className="max-w-full h-auto mx-auto"
                      src={previewObjectUrl}
                    />
                  </div>
                )}

              {!previewLoading &&
                !previewError &&
                previewType === "pdf" &&
                previewObjectUrl && (
                  <div className="rounded-lg border border-primary/10 overflow-hidden bg-slate-50 dark:bg-slate-800/50">
                    <iframe
                      className="w-full h-[60vh]"
                      src={previewObjectUrl}
                      title={previewFile?.filename || "preview-pdf"}
                    />
                  </div>
                )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
