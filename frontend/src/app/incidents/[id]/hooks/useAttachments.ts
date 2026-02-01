import { useState, useEffect, useCallback, useRef } from "react";
import { attachmentApi } from "@/lib/api";
import { Attachment } from "@/types/attachment";

const IMAGE_EXTENSIONS = [
  ".jpg",
  ".jpeg",
  ".png",
  ".gif",
  ".webp",
  ".bmp",
  ".svg",
];

interface UseAttachmentsOptions {
  token: string | null;
  incidentId: string;
}

interface UseAttachmentsReturn {
  attachments: Attachment[];
  loadingAttachments: boolean;
  uploadingFile: boolean;
  imageUrls: Record<number, string>;
  lightboxImage: string | null;
  setLightboxImage: (url: string | null) => void;
  fetchAttachments: () => Promise<void>;
  handleFileUpload: (e: React.ChangeEvent<HTMLInputElement>) => Promise<void>;
  handleFileDownload: (attachmentId: number, fileName: string) => Promise<void>;
  handleFileDelete: (attachmentId: number) => Promise<void>;
  isImageFile: (fileName: string) => boolean;
  formatFileSize: (bytes: number) => string;
}

export function useAttachments({
  token,
  incidentId,
}: UseAttachmentsOptions): UseAttachmentsReturn {
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [loadingAttachments, setLoadingAttachments] = useState(true);
  const [uploadingFile, setUploadingFile] = useState(false);
  const [imageUrls, setImageUrls] = useState<Record<number, string>>({});
  const [lightboxImage, setLightboxImage] = useState<string | null>(null);
  const imageUrlsRef = useRef<Record<number, string>>({});

  const isImageFile = useCallback((fileName: string): boolean => {
    const extension = fileName
      .toLowerCase()
      .substring(fileName.lastIndexOf("."));
    return IMAGE_EXTENSIONS.includes(extension);
  }, []);

  const formatFileSize = useCallback((bytes: number): string => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
  }, []);

  const fetchAttachments = useCallback(async () => {
    if (!token) return;
    setLoadingAttachments(true);
    try {
      const data = await attachmentApi.getAttachments(
        token,
        parseInt(incidentId),
      );
      setAttachments(data);
    } catch (err: unknown) {
      console.error("Failed to fetch attachments:", err);
    } finally {
      setLoadingAttachments(false);
    }
  }, [token, incidentId]);

  useEffect(() => {
    if (token && incidentId) {
      fetchAttachments();
    }
  }, [token, incidentId, fetchAttachments]);

  // Fetch image blob URLs
  useEffect(() => {
    if (!token || attachments.length === 0) return;

    const fetchImageUrls = async () => {
      const urls: Record<number, string> = {};
      const apiUrl =
        process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";

      for (const attachment of attachments) {
        if (isImageFile(attachment.file_name)) {
          try {
            const response = await fetch(
              `${apiUrl}/incidents/${incidentId}/attachments/${attachment.id}/download`,
              {
                headers: {
                  Authorization: `Bearer ${token}`,
                },
              },
            );
            if (response.ok) {
              const blob = await response.blob();
              urls[attachment.id] = URL.createObjectURL(blob);
            }
          } catch (err) {
            console.error(`Failed to fetch image ${attachment.id}:`, err);
          }
        }
      }
      setImageUrls(urls);
      imageUrlsRef.current = urls;
    };

    fetchImageUrls();

    // Cleanup on unmount
    return () => {
      Object.values(imageUrlsRef.current).forEach((url) =>
        URL.revokeObjectURL(url),
      );
    };
  }, [token, attachments, incidentId, isImageFile]);

  const handleFileUpload = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (!file || !token) return;

      setUploadingFile(true);
      try {
        await attachmentApi.uploadAttachment(token, parseInt(incidentId), file);
        await fetchAttachments();
        e.target.value = "";
      } catch (err: unknown) {
        const message =
          err instanceof Error ? err.message : "Failed to upload file";
        alert(message);
      } finally {
        setUploadingFile(false);
      }
    },
    [token, incidentId, fetchAttachments],
  );

  const handleFileDownload = useCallback(
    async (attachmentId: number, fileName: string) => {
      if (!token) return;
      try {
        await attachmentApi.downloadAttachment(
          token,
          parseInt(incidentId),
          attachmentId,
          fileName,
        );
      } catch (err: unknown) {
        const message =
          err instanceof Error ? err.message : "Failed to download file";
        alert(message);
      }
    },
    [token, incidentId],
  );

  const handleFileDelete = useCallback(
    async (attachmentId: number) => {
      if (!token) return;
      if (!confirm("このファイルを削除しますか？")) {
        return;
      }

      try {
        await attachmentApi.deleteAttachment(
          token,
          parseInt(incidentId),
          attachmentId,
        );
        await fetchAttachments();
      } catch (err: unknown) {
        const message =
          err instanceof Error ? err.message : "Failed to delete file";
        alert(message);
      }
    },
    [token, incidentId, fetchAttachments],
  );

  return {
    attachments,
    loadingAttachments,
    uploadingFile,
    imageUrls,
    lightboxImage,
    setLightboxImage,
    fetchAttachments,
    handleFileUpload,
    handleFileDownload,
    handleFileDelete,
    isImageFile,
    formatFileSize,
  };
}
