"use client";

import { Attachment } from "@/types/attachment";

// AuthContext から渡される User 型（完全な User 型とは異なる）
interface CurrentUser {
  id: number;
  name: string;
  email: string;
  role: string;
}

interface IncidentAttachmentsProps {
  attachments: Attachment[];
  loadingAttachments: boolean;
  uploadingFile: boolean;
  imageUrls: Record<number, string>;
  currentUser: CurrentUser | null;
  onFileUpload: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onFileDownload: (attachmentId: number, fileName: string) => void;
  onFileDelete: (attachmentId: number) => void;
  onImageClick: (imageUrl: string) => void;
  isImageFile: (fileName: string) => boolean;
  formatFileSize: (bytes: number) => string;
}

export function IncidentAttachments({
  attachments,
  loadingAttachments,
  uploadingFile,
  imageUrls,
  currentUser,
  onFileUpload,
  onFileDownload,
  onFileDelete,
  onImageClick,
  isImageFile,
  formatFileSize,
}: IncidentAttachmentsProps) {
  return (
    <div className="animate-fadeIn">
      <div
        className="rounded-2xl p-6 border"
        style={{
          background: "var(--surface)",
          borderColor: "var(--border)",
          boxShadow: "var(--shadow-lg)",
        }}
      >
        <h2
          className="text-xl font-bold mb-6"
          style={{
            color: "var(--foreground)",
            fontFamily: "var(--font-display)",
          }}
        >
          添付ファイル
        </h2>

        {/* アップロードフォーム */}
        <div
          className="mb-6 p-5 rounded-xl border-2 border-dashed transition-all"
          style={{
            borderColor: "var(--border)",
            background: "var(--gray-50)",
          }}
        >
          <label
            className="block text-sm font-semibold mb-3"
            style={{
              color: "var(--foreground)",
              fontFamily: "var(--font-body)",
            }}
          >
            ファイルをアップロード
          </label>
          <input
            type="file"
            onChange={onFileUpload}
            disabled={uploadingFile}
            className="block w-full text-sm border-2 rounded-lg cursor-pointer focus:outline-none transition-all disabled:opacity-50 disabled:cursor-not-allowed file:mr-4 file:py-2.5 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-semibold file:transition-all"
            style={{
              background: "var(--surface)",
              borderColor: "var(--border)",
              color: "var(--foreground)",
            }}
          />
          <p
            className="mt-2 text-xs"
            style={{
              color: "var(--foreground-secondary)",
              fontFamily: "var(--font-body)",
            }}
          >
            対応ファイル: 画像 (jpg, png, gif), PDF, テキスト, アーカイブ (zip,
            tar, gz) - 最大 50MB
          </p>
        </div>

        {/* 添付ファイルリスト */}
        {loadingAttachments ? (
          <div
            className="text-center py-8"
            style={{
              color: "var(--foreground-secondary)",
              fontFamily: "var(--font-body)",
            }}
          >
            読み込み中...
          </div>
        ) : attachments.length === 0 ? (
          <div
            className="text-center py-12 rounded-xl"
            style={{ background: "var(--gray-50)" }}
          >
            <svg
              className="mx-auto h-12 w-12 mb-3"
              style={{ color: "var(--foreground-secondary)" }}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"
              />
            </svg>
            <p
              className="text-sm font-medium"
              style={{
                color: "var(--foreground-secondary)",
                fontFamily: "var(--font-body)",
              }}
            >
              添付ファイルはありません
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {attachments.map((attachment) => {
              const isImage = isImageFile(attachment.file_name);
              const imageUrl = imageUrls[attachment.id];

              return (
                <div
                  key={attachment.id}
                  className="border-2 rounded-2xl p-4 transition-all duration-200"
                  style={{
                    background: "var(--surface)",
                    borderColor: "var(--border)",
                    boxShadow: "var(--shadow-sm)",
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.transform = "translateY(-4px)";
                    e.currentTarget.style.boxShadow = "var(--shadow-xl)";
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.transform = "translateY(0)";
                    e.currentTarget.style.boxShadow = "var(--shadow-sm)";
                  }}
                >
                  {isImage && imageUrl ? (
                    <div
                      className="mb-3 cursor-pointer group relative"
                      onClick={() => onImageClick(imageUrl)}
                    >
                      <img
                        src={imageUrl}
                        alt={attachment.file_name}
                        className="w-full h-48 object-cover rounded-md group-hover:opacity-90 transition-opacity"
                        onError={(e) => {
                          console.error(
                            "Image load error:",
                            attachment.file_name,
                          );
                          e.currentTarget.style.display = "none";
                        }}
                      />
                      <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity bg-black bg-opacity-30 rounded-md">
                        <svg
                          className="w-12 h-12 text-white"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM10 7v3m0 0v3m0-3h3m-3 0H7"
                          />
                        </svg>
                      </div>
                    </div>
                  ) : isImage ? (
                    <div className="mb-3 flex items-center justify-center h-48 bg-gray-100 rounded-md">
                      <div className="text-center">
                        <svg
                          className="w-12 h-12 text-gray-400 mx-auto mb-2"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                          />
                        </svg>
                        <p className="text-xs text-gray-500">読み込み中...</p>
                      </div>
                    </div>
                  ) : (
                    <div className="mb-3 flex items-center justify-center h-48 bg-gray-100 rounded-md">
                      <svg
                        className="w-16 h-16 text-gray-400"
                        fill="currentColor"
                        viewBox="0 0 20 20"
                      >
                        <path
                          fillRule="evenodd"
                          d="M8 4a3 3 0 00-3 3v4a5 5 0 0010 0V7a1 1 0 112 0v4a7 7 0 11-14 0V7a5 5 0 0110 0v4a3 3 0 11-6 0V7a1 1 0 012 0v4a1 1 0 102 0V7a3 3 0 00-3-3z"
                          clipRule="evenodd"
                        />
                      </svg>
                    </div>
                  )}
                  <div className="min-w-0">
                    <p
                      className="text-sm font-semibold truncate mb-2"
                      style={{
                        color: "var(--foreground)",
                        fontFamily: "var(--font-body)",
                      }}
                    >
                      {attachment.file_name}
                    </p>
                    <p
                      className="text-xs mb-3"
                      style={{
                        color: "var(--foreground-secondary)",
                        fontFamily: "var(--font-body)",
                      }}
                    >
                      {formatFileSize(attachment.file_size)} •{" "}
                      {attachment.user?.name}
                    </p>
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() =>
                          onFileDownload(attachment.id, attachment.file_name)
                        }
                        className="flex-1 px-3 py-2 text-xs font-semibold text-white rounded-lg transition-all duration-200"
                        style={{
                          background:
                            "linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%)",
                          fontFamily: "var(--font-body)",
                        }}
                        onMouseEnter={(e) =>
                          (e.currentTarget.style.transform = "scale(1.05)")
                        }
                        onMouseLeave={(e) =>
                          (e.currentTarget.style.transform = "scale(1)")
                        }
                      >
                        ダウンロード
                      </button>
                      {(currentUser?.role === "admin" ||
                        currentUser?.id === attachment.user_id) && (
                        <button
                          onClick={() => onFileDelete(attachment.id)}
                          className="px-3 py-2 text-xs font-semibold text-white rounded-lg transition-all duration-200"
                          style={{
                            background:
                              "linear-gradient(135deg, var(--error) 0%, var(--error-dark) 100%)",
                            fontFamily: "var(--font-body)",
                          }}
                          onMouseEnter={(e) =>
                            (e.currentTarget.style.transform = "scale(1.05)")
                          }
                          onMouseLeave={(e) =>
                            (e.currentTarget.style.transform = "scale(1)")
                          }
                        >
                          削除
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
