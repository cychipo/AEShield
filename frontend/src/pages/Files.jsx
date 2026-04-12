import { useState, useEffect, useRef, useCallback } from "react";
import { App as AntdApp, Modal } from "antd";
import { useNavigate } from "react-router-dom";
import {
  Shield,
  LayoutDashboard,
  FolderOpen,
  Settings,
  CloudUpload,
  Download,
  Eye,
  FileCode2,
  Pencil,
  Trash2,
} from "lucide-react";
import AppHeader from "../components/AppHeader";
import {
  decryptEncryptedFile,
  detectPreviewType,
  inspectEncryptedFile,
} from "../utils/encryptedPreview";

const API_BASE_URL =
  import.meta.env.VITE_API_URL ||
  (import.meta.env.DEV ? "http://localhost:6888/api/v1" : "/api/v1");

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
  const [ownedFiles, setOwnedFiles] = useState([]);
  const [sharedFiles, setSharedFiles] = useState([]);
  const [filesLoading, setFilesLoading] = useState(false);
  const [filesError, setFilesError] = useState("");
  const [downloadingFileId, setDownloadingFileId] = useState("");
  const [deletingFileId, setDeletingFileId] = useState("");
  const [editingFileId, setEditingFileId] = useState("");
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

  const [showEncryptedData, setShowEncryptedData] = useState(false);
  const [encryptedDataFile, setEncryptedDataFile] = useState(null);
  const [encryptedDataLoading, setEncryptedDataLoading] = useState(false);
  const [encryptedDataError, setEncryptedDataError] = useState("");
  const [encryptedDataStage, setEncryptedDataStage] = useState("");
  const [encryptedDataInfo, setEncryptedDataInfo] = useState(null);
  const [encryptedDataCopyMessage, setEncryptedDataCopyMessage] = useState("");

  const [showEditFile, setShowEditFile] = useState(false);
  const [editFile, setEditFile] = useState(null);
  const [editFilename, setEditFilename] = useState("");
  const [editAccessMode, setEditAccessMode] = useState("private");
  const [shareEmailInput, setShareEmailInput] = useState("");
  const [shareLookupLoading, setShareLookupLoading] = useState(false);
  const [shareResolveLoading, setShareResolveLoading] = useState(false);
  const [shareLookupError, setShareLookupError] = useState("");
  const [selectedWhitelistUsers, setSelectedWhitelistUsers] = useState([]);
  const [editLoading, setEditLoading] = useState(false);
  const [editError, setEditError] = useState("");
  const [selectedFileIds, setSelectedFileIds] = useState(new Set());
  const [batchDeleting, setBatchDeleting] = useState(false);
  const [batchDeleteErrors, setBatchDeleteErrors] = useState([]);
  const [storageUsage, setStorageUsage] = useState({
    used_bytes: 0,
    quota_bytes: 10 * 1024 * 1024 * 1024,
    percent_used: 0,
  });
  const fileInputRef = useRef(null);
  const previewObjectUrlRef = useRef("");
  const { message } = AntdApp.useApp();
  const navigate = useNavigate();

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

        const storageResponse = await fetch(`${API_BASE_URL}/storage/me`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (storageResponse.ok) {
          const storageData = await storageResponse.json();
          setStorageUsage(storageData);
        } else if (storageResponse.status === 401) {
          localStorage.removeItem("aeshield_token");
          localStorage.removeItem("aeshield_user");
          navigate("/", { replace: true });
          return;
        }

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
      setOwnedFiles(Array.isArray(filesPayload?.owned_files) ? filesPayload.owned_files : []);
      setSharedFiles(Array.isArray(filesPayload?.shared_with_me) ? filesPayload.shared_with_me : []);
      setSelectedFileIds((prev) => {
        const currentIds = new Set(filesPayload?.owned_files?.map((f) => f.id) || []);
        return new Set([...prev].filter((id) => currentIds.has(id)));
      });
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

  const deleteSingleFile = async (fileId, token) => {
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

    return { response, payload };
  };

  const handleDelete = async (fileId, filename) => {
    const token = localStorage.getItem("aeshield_token");
    if (!token) {
      navigate("/", { replace: true });
      return;
    }

    Modal.confirm({
      title: "Xác nhận xóa tệp",
      content: `Bạn có chắc chắn muốn xóa tệp "${filename}" không?`,
      okText: "Xóa",
      cancelText: "Hủy",
      okButtonProps: { danger: true },
      onOk: async () => {
        setFilesError("");
        setDeletingFileId(fileId);

        try {
          const { response, payload } = await deleteSingleFile(fileId, token);

          if (!response.ok) {
            if (response.status === 401) {
              localStorage.removeItem("aeshield_token");
              localStorage.removeItem("aeshield_user");
              navigate("/", { replace: true });
              return;
            }

            const errorMessage = payload?.error || "Không thể xóa tệp tin.";
            setFilesError(errorMessage);
            message.error(errorMessage);
            return;
          }

          setOwnedFiles((prevFiles) => prevFiles.filter((file) => file.id !== fileId));
          message.success(`Đã xóa tệp "${filename}".`);
          await fetchFiles(token);
        } catch (error) {
          console.error("Error deleting file:", error);
          const errorMessage = "Có lỗi xảy ra khi xóa tệp tin.";
          setFilesError(errorMessage);
          message.error(errorMessage);
        } finally {
          setDeletingFileId("");
        }
      },
    });
  };

  const handleBatchDelete = async () => {
    const token = localStorage.getItem("aeshield_token");
    if (!token) {
      navigate("/", { replace: true });
      return;
    }

    const ids = Array.from(selectedFileIds);
    if (ids.length === 0) return;

    Modal.confirm({
      title: "Xác nhận xóa nhiều tệp",
      content: `Bạn có chắc chắn muốn xóa ${ids.length} tệp tin đã chọn?`,
      okText: "Xóa tất cả",
      cancelText: "Hủy",
      okButtonProps: { danger: true },
      onOk: async () => {
        setOwnedFiles((prev) => prev.filter((f) => !selectedFileIds.has(f.id)));
        setSelectedFileIds(new Set());
        setBatchDeleting(true);
        setBatchDeleteErrors([]);
        setFilesError("");

        try {
          const results = await Promise.allSettled(
            ids.map(async (fileId) => {
              const { response, payload } = await deleteSingleFile(fileId, token);
              if (!response.ok) {
                const error = new Error(payload?.error || "Không thể xóa tệp tin.");
                error.status = response.status;
                throw error;
              }
              return fileId;
            })
          );

          const errors = [];
          let unauthorized = false;

          results.forEach((result, i) => {
            if (result.status === "rejected") {
              if (result.reason?.status === 401) {
                unauthorized = true;
              }
              errors.push({ id: ids[i], name: `File #${i + 1}` });
            }
          });

          if (unauthorized) {
            localStorage.removeItem("aeshield_token");
            localStorage.removeItem("aeshield_user");
            navigate("/", { replace: true });
            return;
          }

          if (errors.length > 0) {
            setBatchDeleteErrors(errors);
            const errorMessage = `Đã xóa ${ids.length - errors.length}/${ids.length} tệp. ${errors.length} tệp thất bại.`;
            setFilesError(errorMessage);
            message.warning(errorMessage);
          } else {
            message.success(`Đã xóa ${ids.length} tệp tin.`);
          }

          await fetchFiles(token);
        } catch (error) {
          console.error("Error deleting multiple files:", error);
          const errorMessage = "Có lỗi xảy ra khi xóa nhiều tệp tin.";
          setFilesError(errorMessage);
          message.error(errorMessage);
        } finally {
          setBatchDeleting(false);
        }
      },
    });
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

  const closeEncryptedData = () => {
    if (encryptedDataLoading) {
      return;
    }

    setShowEncryptedData(false);
    setEncryptedDataFile(null);
    setEncryptedDataLoading(false);
    setEncryptedDataError("");
    setEncryptedDataStage("");
    setEncryptedDataInfo(null);
    setEncryptedDataCopyMessage("");
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

  const openEncryptedData = (file) => {
    const rowFileId = file?.id;
    if (!rowFileId) {
      return;
    }

    setShowEncryptedData(true);
    setEncryptedDataFile(file);
    setEncryptedDataLoading(false);
    setEncryptedDataError("");
    setEncryptedDataStage("");
    setEncryptedDataInfo(null);
    setEncryptedDataCopyMessage("");
  };

  const hydrateWhitelistUsers = useCallback(async (whitelistIds) => {
    const token = localStorage.getItem("aeshield_token");
    if (!token || !Array.isArray(whitelistIds) || whitelistIds.length === 0) {
      return;
    }

    setShareResolveLoading(true);
    try {
      const response = await fetch(`${API_BASE_URL}/users/resolve`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ ids: whitelistIds }),
      });

      if (!response.ok) {
        return;
      }

      const payload = await response.json();
      if (!Array.isArray(payload)) {
        return;
      }

      const resolvedMap = new Map(payload.map((item) => [item.id, item]));
      setSelectedWhitelistUsers((current) =>
        current.map((item) => resolvedMap.get(item.id) || item)
      );
    } catch (error) {
      console.error("Error resolving whitelist users:", error);
    } finally {
      setShareResolveLoading(false);
    }
  }, [navigate]);

  const openEditFile = (file) => {
    const rowFileId = file?.id;
    if (!rowFileId) {
      return;
    }

    const initialWhitelist = Array.isArray(file?.whitelist)
      ? file.whitelist.map((id) => ({
          id,
          email: "",
          name: "",
          avatar: "",
        }))
      : [];

    setShowEditFile(true);
    setEditFile(file);
    setEditFilename(file?.filename || "");
    setEditAccessMode(file?.access_mode || "private");
    setShareEmailInput("");
    setShareLookupLoading(false);
    setShareResolveLoading(false);
    setShareLookupError("");
    setSelectedWhitelistUsers(initialWhitelist);
    setEditLoading(false);
    setEditError("");
    setEditingFileId(rowFileId);

    if (initialWhitelist.length > 0) {
      hydrateWhitelistUsers(initialWhitelist.map((item) => item.id));
    }
  };

  const closeEditFile = () => {
    if (editLoading || shareLookupLoading || shareResolveLoading) {
      return;
    }

    setShowEditFile(false);
    setEditFile(null);
    setEditFilename("");
    setEditAccessMode("private");
    setShareEmailInput("");
    setShareLookupLoading(false);
    setShareResolveLoading(false);
    setShareLookupError("");
    setSelectedWhitelistUsers([]);
    setEditLoading(false);
    setEditError("");
    setEditingFileId("");
  };

  const handleAddWhitelistUser = async () => {
    const token = localStorage.getItem("aeshield_token");
    if (!token) {
      navigate("/", { replace: true });
      return;
    }

    const email = shareEmailInput.trim();
    if (!email) {
      setShareLookupError("Vui lòng nhập email người dùng.");
      return;
    }

    setShareLookupLoading(true);
    setShareLookupError("");

    try {
      const response = await fetch(
        `${API_BASE_URL}/users/lookup?email=${encodeURIComponent(email)}`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

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

        setShareLookupError(payload?.error || "Không tìm thấy người dùng.");
        return;
      }

      if (!payload?.id) {
        setShareLookupError("Không nhận được thông tin người dùng hợp lệ.");
        return;
      }

      setSelectedWhitelistUsers((current) => {
        if (current.some((item) => item.id === payload.id)) {
          return current;
        }
        return [
          ...current,
          {
            id: payload.id,
            email: payload.email || email,
            name: payload.name || "",
            avatar: payload.avatar || "",
          },
        ];
      });
      setShareEmailInput("");
    } catch (error) {
      console.error("Error looking up user:", error);
      setShareLookupError("Có lỗi xảy ra khi tìm người dùng theo email.");
    } finally {
      setShareLookupLoading(false);
    }
  };

  const handleRemoveWhitelistUser = (userId) => {
    setSelectedWhitelistUsers((current) =>
      current.filter((item) => item.id !== userId)
    );
    setShareLookupError("");
  };

  const handleEditFile = async (event) => {
    event.preventDefault();

    const token = localStorage.getItem("aeshield_token");
    if (!token) {
      navigate("/", { replace: true });
      return;
    }

    const fileId = editFile?.id;
    if (!fileId) {
      setEditError("Không tìm thấy tệp tin để chỉnh sửa.");
      return;
    }

    const nextFilename = editFilename.trim();
    if (!nextFilename) {
      setEditError("Tên tệp không được để trống.");
      return;
    }

    const nextWhitelist =
      editAccessMode === "whitelist"
        ? selectedWhitelistUsers.map((item) => item.id).filter(Boolean)
        : [];

    setEditLoading(true);
    setEditError("");

    try {
      const response = await fetch(`${API_BASE_URL}/files/${fileId}`, {
        method: "PATCH",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          filename: nextFilename,
          access_mode: editAccessMode,
          whitelist: nextWhitelist,
        }),
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

        setEditError(payload?.error || "Không thể cập nhật tệp tin.");
        return;
      }

      setOwnedFiles((prevFiles) =>
        prevFiles.map((file) => (file.id === fileId ? payload : file))
      );

      closeEditFile();
      await fetchFiles(token);
    } catch (error) {
      console.error("Error editing file:", error);
      setEditError("Có lỗi xảy ra khi cập nhật tệp tin.");
    } finally {
      setEditLoading(false);
    }
  };

  const handleInspectEncryptedData = async () => {
    const fileId = encryptedDataFile?.id;
    if (!fileId) {
      setEncryptedDataError("Không tìm thấy tệp tin để xem dữ liệu mã hóa.");
      return;
    }

    const token = localStorage.getItem("aeshield_token");
    if (!token) {
      navigate("/", { replace: true });
      return;
    }

    setEncryptedDataLoading(true);
    setEncryptedDataError("");
    setEncryptedDataInfo(null);
    setEncryptedDataCopyMessage("");
    setEncryptedDataStage("Đang lấy liên kết tải...");

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

        setEncryptedDataError(
          payload?.error || "Không thể lấy tệp tin để xem dữ liệu mã hóa."
        );
        return;
      }

      if (!payload?.url) {
        setEncryptedDataError("Không nhận được đường dẫn tải tệp tin.");
        return;
      }

      setEncryptedDataStage("Đang tải dữ liệu mã hóa...");
      const encryptedResponse = await fetch(payload.url);
      if (!encryptedResponse.ok) {
        setEncryptedDataError("Không thể tải dữ liệu tệp tin.");
        return;
      }

      const encryptedBuffer = await encryptedResponse.arrayBuffer();

      setEncryptedDataStage("Đang phân tích dữ liệu mã hóa...");
      const inspected = inspectEncryptedFile(new Uint8Array(encryptedBuffer), {
        maxPreviewBytes: 512,
      });

      setEncryptedDataInfo(inspected);
      setEncryptedDataError("");
    } catch (error) {
      setEncryptedDataInfo(null);
      setEncryptedDataError(
        error instanceof Error && error.message
          ? error.message
          : "Có lỗi xảy ra khi phân tích dữ liệu mã hóa."
      );
    } finally {
      setEncryptedDataLoading(false);
      setEncryptedDataStage("");
    }
  };

  const handleCopyEncryptedData = async (kind) => {
    const value =
      kind === "base64"
        ? encryptedDataInfo?.previewBase64
        : encryptedDataInfo?.previewHex;

    if (!value) {
      setEncryptedDataCopyMessage("Không có dữ liệu để sao chép.");
      return;
    }

    if (!navigator?.clipboard?.writeText) {
      setEncryptedDataCopyMessage("Trình duyệt không hỗ trợ sao chép tự động.");
      return;
    }

    try {
      await navigator.clipboard.writeText(value);
      setEncryptedDataCopyMessage(
        kind === "base64" ? "Đã copy Base64." : "Đã copy Hex dump."
      );
    } catch {
      setEncryptedDataCopyMessage("Sao chép thất bại. Vui lòng thử lại.");
    }
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
      setSelectedFileIds(new Set());
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

  const handleToggleSelect = useCallback((fileId) => {
    setSelectedFileIds((prev) => {
      const next = new Set(prev);
      if (next.has(fileId)) {
        next.delete(fileId);
      } else {
        next.add(fileId);
      }
      return next;
    });
  }, []);

  const handleToggleSelectAll = useCallback((fileList) => {
    const allSelected = fileList.every((f) => selectedFileIds.has(f.id));
    setSelectedFileIds((prev) => {
      const next = new Set(prev);
      if (allSelected) {
        fileList.forEach((f) => next.delete(f.id));
      } else {
        fileList.forEach((f) => next.add(f.id));
      }
      return next;
    });
  }, [selectedFileIds]);

  const renderFileTable = (fileList, options = {}) => {
    const {
      allowManage = false,
      emptyTitle,
      emptyDescription,
      showSharedBadge = false,
    } = options;

    const tableSelectedCount = fileList.filter((f) => selectedFileIds.has(f.id)).length;

    if (fileList.length === 0) {
      return (
        <div className="bg-white dark:bg-slate-900 rounded-xl border border-primary/10 shadow-sm p-6 text-center">
          <FolderOpen size={40} className="text-slate-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-slate-700 dark:text-slate-300 mb-2">
            {emptyTitle}
          </h3>
          <p className="text-slate-500 dark:text-slate-400">{emptyDescription}</p>
        </div>
      );
    }

    return (
      <div className="bg-white dark:bg-slate-900 rounded-xl border border-primary/10 shadow-sm overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-slate-50 dark:bg-slate-800/50">
              <tr>
                {allowManage && (
                  <th className="w-10 px-4 py-3">
                    <input
                      checked={tableSelectedCount > 0 && tableSelectedCount === fileList.length}
                      className="w-4 h-4 rounded border-primary/30 text-primary focus:ring-primary cursor-pointer"
                      disabled={batchDeleting}
                      onChange={() => handleToggleSelectAll(fileList)}
                      title={tableSelectedCount === fileList.length ? "Bỏ chọn tất cả" : "Chọn tất cả"}
                      type="checkbox"
                    />
                  </th>
                )}
                <th className="text-left font-medium px-4 py-3">Tên tệp</th>
                <th className="text-left font-medium px-4 py-3">Kích thước</th>
                <th className="text-left font-medium px-4 py-3">Mã hóa</th>
                <th className="text-left font-medium px-4 py-3">Truy cập</th>
                <th className="text-left font-medium px-4 py-3">Cập nhật</th>
                <th className="text-left font-medium px-4 py-3">Thao tác</th>
              </tr>
            </thead>
            <tbody>
              {fileList.map((file) => {
                const rowFileId = file.id;
                const isDownloading = downloadingFileId === rowFileId;
                const isDeleting = deletingFileId === rowFileId;
                const isEditing = showEditFile && editingFileId === rowFileId;
                const isPreviewing = showPreview && previewingFileId === rowFileId;
                const isInspectingEncryptedData =
                  showEncryptedData && encryptedDataFile?.id === rowFileId;
                const isSelected = selectedFileIds.has(rowFileId);

                return (
                  <tr
                    className={`border-t border-primary/10 ${isSelected ? "bg-primary/5" : ""}`}
                    key={file.id || file.storage_path || file.filename}
                  >
                    {allowManage && (
                      <td className="px-4 py-3">
                        <input
                          checked={isSelected}
                          className="w-4 h-4 rounded border-primary/30 text-primary focus:ring-primary cursor-pointer"
                          disabled={batchDeleting || isDeleting}
                          onChange={() => handleToggleSelect(rowFileId)}
                          type="checkbox"
                        />
                      </td>
                    )}
                    <td className="px-4 py-3 font-medium">
                      <div className="flex items-center gap-2">
                        <span>{file.filename}</span>
                        {showSharedBadge && (
                          <span className="inline-flex items-center rounded-full bg-primary/10 px-2 py-0.5 text-[11px] font-medium text-primary">
                            Được chia sẻ
                          </span>
                        )}
                      </div>
                    </td>
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
                          className="inline-flex items-center justify-center p-1.5 rounded-lg border border-primary/20 text-xs font-medium hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed"
                          disabled={
                            !rowFileId ||
                            isPreviewing ||
                            isDownloading ||
                            isDeleting ||
                            isInspectingEncryptedData ||
                            isEditing
                          }
                          onClick={() => openPreview(file)}
                          title={isPreviewing ? "Đang xem trước" : "Xem trước"}
                          type="button"
                        >
                          <Eye size={14} />
                        </button>
                        <button
                          className="inline-flex items-center justify-center p-1.5 rounded-lg border border-primary/20 text-xs font-medium hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed"
                          disabled={
                            !rowFileId ||
                            isDownloading ||
                            isDeleting ||
                            isPreviewing ||
                            isInspectingEncryptedData ||
                            isEditing
                          }
                          onClick={() => openEncryptedData(file)}
                          title={
                            isInspectingEncryptedData
                              ? "Đang xem dữ liệu mã hóa"
                              : "Xem dữ liệu mã hóa"
                          }
                          type="button"
                        >
                          <FileCode2 size={14} />
                        </button>
                        {allowManage && (
                          <button
                            className="inline-flex items-center justify-center p-1.5 rounded-lg border border-primary/20 text-xs font-medium hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed"
                            disabled={
                              !rowFileId ||
                              isDownloading ||
                              isDeleting ||
                              isPreviewing ||
                              isInspectingEncryptedData ||
                              isEditing
                            }
                            onClick={() => openEditFile(file)}
                            title={isEditing ? "Đang chỉnh sửa" : "Chỉnh sửa"}
                            type="button"
                          >
                            <Pencil size={14} />
                          </button>
                        )}
                        <button
                          className="inline-flex items-center justify-center p-1.5 rounded-lg border border-primary/20 text-xs font-medium hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed"
                          disabled={
                            !rowFileId ||
                            isDownloading ||
                            isDeleting ||
                            isPreviewing ||
                            isInspectingEncryptedData ||
                            isEditing
                          }
                          onClick={() => handleDownload(rowFileId)}
                          title={isDownloading ? "Đang lấy link tải" : "Tải xuống"}
                          type="button"
                        >
                          <Download size={14} />
                        </button>
                        {allowManage && (
                          <button
                            className="inline-flex items-center justify-center p-1.5 rounded-lg border border-red-300 text-red-600 text-xs font-medium hover:bg-red-50 dark:border-red-700 dark:text-red-300 dark:hover:bg-red-950/30 disabled:opacity-50 disabled:cursor-not-allowed"
                            disabled={
                              !rowFileId ||
                              isDeleting ||
                              isDownloading ||
                              isPreviewing ||
                              isInspectingEncryptedData ||
                              isEditing
                            }
                            onClick={() =>
                              handleDelete(rowFileId, file.filename || "tệp tin này")
                            }
                            title={isDeleting ? "Đang xóa" : "Xóa"}
                            type="button"
                          >
                            <Trash2 size={14} />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    );
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
            className="flex items-center gap-3 px-3 py-2 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-primary/5 hover:text-primary dark:hover:bg-primary/10 dark:hover:text-primary transition-colors"
            href="/dashboard"
          >
            <LayoutDashboard size={20} />
            <span>Dashboard</span>
          </a>
          <a
            className="flex items-center gap-3 px-3 py-2 rounded-lg bg-primary/10 text-primary font-medium hover:bg-primary/20 hover:text-primary dark:hover:bg-primary/25 dark:hover:text-primary hover:shadow-sm transition-all"
            href="/files"
          >
            <FolderOpen size={20} />
            <span>Tệp tin</span>
          </a>
          <a
            className="flex items-center gap-3 px-3 py-2 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-primary/5 hover:text-primary dark:hover:bg-primary/10 dark:hover:text-primary transition-colors"
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
        <AppHeader user={user} searchPlaceholder="Tìm kiếm tệp tin..." />

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
            ) : (
              <div className="space-y-8">
                <section className="space-y-4">
                  <div className="flex items-center justify-between gap-3">
                    <div>
                      <h2 className="text-xl font-semibold">Tệp tin của bạn</h2>
                      <p className="text-sm text-slate-500 dark:text-slate-400">
                        Các tệp bạn đã tải lên và có toàn quyền quản lý.
                      </p>
                    </div>
                    {selectedFileIds.size > 0 && (
                      <div className="flex items-center gap-3">
                        <span className="text-sm text-slate-500 dark:text-slate-400">
                          {selectedFileIds.size} tệp đã chọn
                        </span>
                        <button
                          className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg border border-red-300 text-red-600 text-sm font-medium hover:bg-red-50 dark:border-red-700 dark:text-red-300 dark:hover:bg-red-950/30 disabled:opacity-50 disabled:cursor-not-allowed"
                          disabled={batchDeleting}
                          onClick={handleBatchDelete}
                          type="button"
                        >
                          <Trash2 size={14} />
                          {batchDeleting ? "Đang xóa..." : "Xóa đã chọn"}
                        </button>
                        <button
                          className="px-3 py-1.5 rounded-lg border border-slate-300 text-slate-600 text-sm font-medium hover:bg-slate-50 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-800"
                          onClick={() => setSelectedFileIds(new Set())}
                          type="button"
                        >
                          Bỏ chọn
                        </button>
                      </div>
                    )}
                  </div>

                  {ownedFiles.length === 0 ? (
                    <div className="bg-white dark:bg-slate-900 rounded-xl border border-primary/10 shadow-sm p-6 text-center">
                      <FolderOpen size={48} className="text-slate-400 mx-auto mb-4" />
                      <h3 className="text-lg font-medium text-slate-700 dark:text-slate-300 mb-2">
                        Chưa có tệp tin nào
                      </h3>
                      <p className="text-slate-500 dark:text-slate-400 mb-4">
                        Tải lên tệp tin đầu tiên của bạn để bắt đầu mã hóa và bảo vệ dữ liệu.
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
                    renderFileTable(ownedFiles, { allowManage: true })
                  )}
                </section>

                <section className="space-y-4">
                  <div>
                    <h2 className="text-xl font-semibold">Tệp tin được chia sẻ với bạn</h2>
                    <p className="text-sm text-slate-500 dark:text-slate-400">
                      Các tệp mà người khác đã cấp quyền cho bạn qua whitelist.
                    </p>
                  </div>

                  {renderFileTable(sharedFiles, {
                    allowManage: false,
                    showSharedBadge: true,
                    emptyTitle: "Chưa có tệp nào được chia sẻ",
                    emptyDescription: "Khi người khác thêm quyền cho bạn, tệp sẽ xuất hiện tại đây.",
                  })}
                </section>
              </div>
            )}
          </div>
        </div>
      </main>

      {showEditFile && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-2xl max-h-[90vh] overflow-hidden rounded-xl border border-primary/20 bg-white dark:bg-slate-900 shadow-xl flex flex-col">
            <div className="px-5 py-4 border-b border-primary/10 flex items-center justify-between gap-4">
              <div className="min-w-0">
                <h2 className="text-lg font-semibold">Chỉnh sửa tệp tin</h2>
                <p className="text-sm text-slate-500 dark:text-slate-400 truncate">
                  {editFile?.filename || "-"}
                </p>
              </div>
              <button
                className="px-3 py-1.5 rounded-lg border border-primary/20 text-sm hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed"
                disabled={editLoading}
                onClick={closeEditFile}
                type="button"
              >
                Đóng
              </button>
            </div>

            <div className="p-5 overflow-y-auto space-y-4">
              <form className="space-y-4" onSubmit={handleEditFile}>
                <div>
                  <label className="block text-sm font-medium mb-2" htmlFor="edit-filename">
                    Tên tệp
                  </label>
                  <input
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-800 border border-primary/20 rounded-lg text-sm"
                    disabled={editLoading}
                    id="edit-filename"
                    onChange={(event) => setEditFilename(event.target.value)}
                    placeholder="Nhập tên tệp"
                    type="text"
                    value={editFilename}
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2" htmlFor="edit-access-mode">
                    Chế độ truy cập
                  </label>
                  <select
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-800 border border-primary/20 rounded-lg text-sm"
                    disabled={editLoading}
                    id="edit-access-mode"
                    onChange={(event) => setEditAccessMode(event.target.value)}
                    value={editAccessMode}
                  >
                    <option value="private">private</option>
                    <option value="public">public</option>
                    <option value="whitelist">whitelist</option>
                  </select>
                </div>

                {editAccessMode === "whitelist" && (
                  <div className="space-y-3">
                    <div>
                      <label className="block text-sm font-medium mb-2" htmlFor="share-email-input">
                        Cấp quyền bằng email
                      </label>
                      <div className="flex gap-2">
                        <input
                          className="flex-1 px-3 py-2 bg-slate-50 dark:bg-slate-800 border border-primary/20 rounded-lg text-sm"
                          disabled={editLoading || shareLookupLoading}
                          id="share-email-input"
                          onChange={(event) => setShareEmailInput(event.target.value)}
                          placeholder="user@example.com"
                          type="email"
                          value={shareEmailInput}
                        />
                        <button
                          className="px-4 py-2 rounded-lg border border-primary/20 text-sm font-medium hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed"
                          disabled={editLoading || shareLookupLoading || shareResolveLoading}
                          onClick={handleAddWhitelistUser}
                          type="button"
                        >
                          {shareLookupLoading ? "Đang tìm..." : "Thêm quyền"}
                        </button>
                      </div>
                      <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
                        Nhập đúng email của người dùng đã có tài khoản trong hệ thống.
                      </p>
                    </div>

                    {shareResolveLoading && (
                      <p className="text-xs text-slate-500 dark:text-slate-400">
                        Đang tải thông tin whitelist hiện có...
                      </p>
                    )}

                    {selectedWhitelistUsers.length > 0 && (
                      <div className="space-y-2">
                        <p className="text-sm font-medium">Người dùng được cấp quyền</p>
                        <div className="space-y-2">
                          {selectedWhitelistUsers.map((shareUser) => (
                            <div
                              className="flex items-center justify-between gap-3 rounded-lg border border-primary/10 bg-slate-50 dark:bg-slate-800 px-3 py-2"
                              key={shareUser.id}
                            >
                              <div className="flex items-center gap-3 min-w-0">
                                <div
                                  className="w-10 h-10 rounded-full bg-slate-200 dark:bg-slate-700 border border-primary/10 bg-cover bg-center shrink-0"
                                  style={{
                                    backgroundImage: shareUser.avatar
                                      ? `url(${shareUser.avatar})`
                                      : "none",
                                  }}
                                ></div>
                                <div className="min-w-0">
                                  <p className="text-sm font-medium truncate">
                                    {shareUser.name || shareUser.email || shareUser.id}
                                  </p>
                                  <p className="text-xs text-slate-500 dark:text-slate-400 truncate">
                                    {shareUser.email || shareUser.id}
                                  </p>
                                </div>
                              </div>
                              <button
                                className="px-3 py-1.5 rounded-lg border border-red-300 text-red-600 text-xs font-medium hover:bg-red-50 dark:border-red-700 dark:text-red-300 dark:hover:bg-red-950/30 disabled:opacity-50 disabled:cursor-not-allowed"
                                disabled={editLoading || shareLookupLoading}
                                onClick={() => handleRemoveWhitelistUser(shareUser.id)}
                                type="button"
                              >
                                Gỡ
                              </button>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {shareLookupError && (
                      <div className="rounded-lg border border-red-200 bg-red-50 text-red-700 px-3 py-2 text-sm dark:bg-red-950/30 dark:border-red-800 dark:text-red-300">
                        {shareLookupError}
                      </div>
                    )}
                  </div>
                )}

                {editError && (
                  <div className="rounded-lg border border-red-200 bg-red-50 text-red-700 px-3 py-2 text-sm dark:bg-red-950/30 dark:border-red-800 dark:text-red-300">
                    {editError}
                  </div>
                )}

                <div className="flex items-center gap-3">
                  <button
                    className="bg-primary text-white px-4 py-2 rounded-lg font-medium hover:bg-orange-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                    disabled={editLoading || !editFile?.id}
                    type="submit"
                  >
                    {editLoading ? "Đang lưu..." : "Lưu thay đổi"}
                  </button>
                  <button
                    className="px-4 py-2 rounded-lg font-medium border border-primary/20 hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed"
                    disabled={editLoading}
                    onClick={closeEditFile}
                    type="button"
                  >
                    Hủy
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {showEncryptedData && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-4xl max-h-[90vh] overflow-hidden rounded-xl border border-primary/20 bg-white dark:bg-slate-900 shadow-xl flex flex-col">
            <div className="px-5 py-4 border-b border-primary/10 flex items-center justify-between gap-4">
              <div className="min-w-0">
                <h2 className="text-lg font-semibold">Dữ liệu mã hóa</h2>
                <p className="text-sm text-slate-500 dark:text-slate-400 truncate">
                  {encryptedDataFile?.filename || "-"}
                </p>
              </div>
              <button
                className="px-3 py-1.5 rounded-lg border border-primary/20 text-sm hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed"
                disabled={encryptedDataLoading}
                onClick={closeEncryptedData}
                type="button"
              >
                Đóng
              </button>
            </div>

            <div className="p-5 overflow-y-auto space-y-4">
              <div className="flex items-center gap-3">
                <button
                  className="bg-primary text-white px-4 py-2 rounded-lg font-medium hover:bg-orange-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  disabled={encryptedDataLoading || !encryptedDataFile?.id}
                  onClick={handleInspectEncryptedData}
                  type="button"
                >
                  {encryptedDataLoading ? "Đang phân tích..." : "Tải & phân tích dữ liệu mã hóa"}
                </button>
                <span className="text-xs text-slate-500 dark:text-slate-400">
                  Hiển thị header, chunk stats, hex/base64 bytes đầu
                </span>
              </div>

              {encryptedDataLoading && encryptedDataStage && (
                <div className="rounded-lg border border-primary/20 bg-primary/5 text-primary px-3 py-2 text-sm">
                  {encryptedDataStage}
                </div>
              )}

              {encryptedDataError && (
                <div className="rounded-lg border border-red-200 bg-red-50 text-red-700 px-3 py-2 text-sm dark:bg-red-950/30 dark:border-red-800 dark:text-red-300">
                  {encryptedDataError}
                </div>
              )}

              {!encryptedDataLoading && !encryptedDataError && !encryptedDataInfo && (
                <div className="rounded-lg border border-primary/20 bg-slate-50 dark:bg-slate-800/60 text-slate-600 dark:text-slate-300 px-3 py-2 text-sm">
                  Bấm nút để tải bytes mã hóa và xem cấu trúc dữ liệu ciphertext.
                </div>
              )}

              {!encryptedDataLoading && !encryptedDataError && encryptedDataInfo && (
                <div className="space-y-3">
                  <div className="flex flex-wrap items-center gap-2">
                    <button
                      className="px-3 py-1.5 rounded-lg border border-primary/20 text-xs font-medium hover:bg-slate-50 dark:hover:bg-slate-800"
                      onClick={() => handleCopyEncryptedData("hex")}
                      type="button"
                    >
                      Copy Hex
                    </button>
                    <button
                      className="px-3 py-1.5 rounded-lg border border-primary/20 text-xs font-medium hover:bg-slate-50 dark:hover:bg-slate-800"
                      onClick={() => handleCopyEncryptedData("base64")}
                      type="button"
                    >
                      Copy Base64
                    </button>
                    {encryptedDataCopyMessage && (
                      <span className="text-xs text-slate-500 dark:text-slate-400">
                        {encryptedDataCopyMessage}
                      </span>
                    )}
                  </div>
                  <div className="rounded-lg border border-primary/10 bg-slate-50 dark:bg-slate-800/50 p-4 text-sm grid grid-cols-1 md:grid-cols-2 gap-3">
                    <p>
                      <span className="font-semibold">Loại mã hóa:</span>{" "}
                      {encryptedDataFile?.encryption_type || "-"}
                    </p>
                    <p>
                      <span className="font-semibold">Tổng bytes mã hóa:</span>{" "}
                      {encryptedDataInfo.totalEncryptedBytes}
                    </p>
                    <p>
                      <span className="font-semibold">Header size:</span>{" "}
                      {encryptedDataInfo.headerSize}
                    </p>
                    <p>
                      <span className="font-semibold">Key byte:</span>{" "}
                      {encryptedDataInfo.keyByte}
                    </p>
                    <p>
                      <span className="font-semibold">Key length:</span>{" "}
                      {encryptedDataInfo.keyLength} bytes
                    </p>
                    <p className="md:col-span-2 break-all">
                      <span className="font-semibold">Salt (hex):</span>{" "}
                      {encryptedDataInfo.saltHex}
                    </p>
                    <p className="md:col-span-2 break-all">
                      <span className="font-semibold">Base nonce (hex):</span>{" "}
                      {encryptedDataInfo.baseNonceHex}
                    </p>
                    <p>
                      <span className="font-semibold">Số chunk:</span>{" "}
                      {encryptedDataInfo.chunkCount}
                    </p>
                    <p>
                      <span className="font-semibold">Cipher payload bytes:</span>{" "}
                      {encryptedDataInfo.totalCipherPayloadBytes}
                    </p>
                    <p className="md:col-span-2 text-xs text-slate-500 dark:text-slate-400">
                      {encryptedDataInfo.truncated}
                    </p>
                  </div>

                  <div className="space-y-2">
                    <h3 className="text-sm font-semibold">Hex dump (bytes đầu)</h3>
                    <pre className="rounded-lg border border-primary/10 bg-slate-50 dark:bg-slate-800/50 p-4 text-xs whitespace-pre-wrap break-all max-h-[28vh] overflow-auto">
                      {encryptedDataInfo.previewHex}
                    </pre>
                  </div>

                  <div className="space-y-2">
                    <h3 className="text-sm font-semibold">Base64 (bytes đầu)</h3>
                    <pre className="rounded-lg border border-primary/10 bg-slate-50 dark:bg-slate-800/50 p-4 text-xs whitespace-pre-wrap break-all max-h-[22vh] overflow-auto">
                      {encryptedDataInfo.previewBase64}
                    </pre>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

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
