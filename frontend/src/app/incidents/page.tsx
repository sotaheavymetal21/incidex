"use client";

import { useEffect, useState, useCallback, Suspense } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/AuthContext";
import { usePermissions } from "@/hooks/usePermissions";
import { incidentApi, tagApi, exportApi } from "@/lib/api";
import { Incident, Severity, Status } from "@/types/incident";
import { Tag } from "@/types/tag";
import SeverityGuide from "@/components/SeverityGuide";
import {
  IncidentFilters,
  IncidentTable,
  IncidentPagination,
} from "./components";
import { useIncidentFilters } from "./hooks/useIncidentFilters";
import { getSeverityStyle, getStatusStyle } from "./[id]/utils/styles";

const SEVERITY_LABELS: Record<Severity, string> = {
  critical: "Critical",
  high: "High",
  medium: "Medium",
  low: "Low",
};

const STATUS_LABELS: Record<Status, string> = {
  open: "Open",
  investigating: "Investigating",
  resolved: "Resolved",
  closed: "Closed",
};

function IncidentsPageContent() {
  const { token, loading: authLoading } = useAuth();
  const permissions = usePermissions();
  const router = useRouter();
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [showSidebar, setShowSidebar] = useState(true);

  const filters = useIncidentFilters();

  const fetchTags = useCallback(async () => {
    if (!token) return;
    try {
      const fetchedTags = await tagApi.getAll(token);
      setTags(fetchedTags);
    } catch (err) {
      console.error("Failed to fetch tags:", err);
    }
  }, [token]);

  const fetchIncidents = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    setError("");
    try {
      const response = await incidentApi.getAll(token, {
        page: filters.pagination.page,
        limit: filters.pagination.limit,
        search: filters.search || undefined,
        severity: filters.severity || undefined,
        status: filters.status || undefined,
        tag_ids:
          filters.selectedTagIds.length > 0
            ? filters.selectedTagIds.join(",")
            : undefined,
      });
      setIncidents(response.incidents);
      filters.setPagination(response.pagination);
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to fetch incidents";
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  }, [
    token,
    filters.pagination.page,
    filters.pagination.limit,
    filters.search,
    filters.severity,
    filters.status,
    filters.selectedTagIds,
    filters.setPagination,
  ]);

  useEffect(() => {
    if (authLoading) return;
    if (!token) {
      router.push("/login");
    }
  }, [token, authLoading, router]);

  useEffect(() => {
    if (authLoading || !token) return;
    fetchTags();
  }, [token, authLoading, fetchTags]);

  useEffect(() => {
    if (authLoading || !token) return;
    fetchIncidents();
  }, [token, authLoading, fetchIncidents]);

  const handleExportCSV = useCallback(async () => {
    if (!token) return;
    try {
      const blob = await exportApi.exportIncidentsCSV(token, {
        search: filters.search || undefined,
        severity: filters.severity || undefined,
        status: filters.status || undefined,
        tag_ids:
          filters.selectedTagIds.length > 0
            ? filters.selectedTagIds.join(",")
            : undefined,
      });

      const url = window.URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `incidents_${new Date().toISOString().split("T")[0]}.csv`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (err) {
      console.error("Export failed:", err);
      setError("Failed to export incidents");
    }
  }, [
    token,
    filters.search,
    filters.severity,
    filters.status,
    filters.selectedTagIds,
  ]);

  if (authLoading || !token) {
    return (
      <div
        className="min-h-screen flex items-center justify-center"
        style={{ background: "var(--background)" }}
      >
        <div style={{ color: "var(--secondary)" }}>Loading...</div>
      </div>
    );
  }

  return (
    <div
      className="min-h-screen py-8 px-4 sm:px-6 lg:px-8"
      style={{ background: "var(--background)" }}
    >
      <div className="max-w-7xl mx-auto">
        {/* ヘッダー */}
        <div className="mb-8 flex justify-between items-center animate-slideDown">
          <div>
            <h1
              className="text-4xl font-bold mb-2"
              style={{
                color: "var(--foreground)",
                fontFamily: "var(--font-display)",
                background:
                  "linear-gradient(135deg, var(--foreground) 0%, var(--primary) 100%)",
                WebkitBackgroundClip: "text",
                WebkitTextFillColor: "transparent",
              }}
            >
              インシデント一覧
            </h1>
            <p
              className="text-base font-medium"
              style={{
                color: "var(--foreground-secondary)",
                fontFamily: "var(--font-body)",
              }}
            >
              インシデントの管理と追跡
            </p>
          </div>
          <div className="flex space-x-3">
            <button
              onClick={handleExportCSV}
              className="px-4 py-2.5 text-white rounded-xl flex items-center font-semibold transition-all duration-200"
              style={{
                background:
                  "linear-gradient(135deg, var(--accent) 0%, var(--accent-hover) 100%)",
                fontFamily: "var(--font-body)",
                boxShadow: "0 4px 12px var(--accent-glow)",
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.transform = "translateY(-2px)";
                e.currentTarget.style.boxShadow =
                  "0 8px 20px var(--accent-glow)";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.transform = "translateY(0)";
                e.currentTarget.style.boxShadow =
                  "0 4px 12px var(--accent-glow)";
              }}
            >
              <svg
                className="w-5 h-5 mr-2"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                />
              </svg>
              CSVエクスポート
            </button>
            {permissions.canCreateIncidents && (
              <button
                onClick={() => router.push("/incidents/create")}
                className="px-5 py-2.5 text-white rounded-xl font-bold transition-all duration-200"
                style={{
                  background:
                    "linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%)",
                  fontFamily: "var(--font-body)",
                  boxShadow: "0 4px 12px var(--primary-glow)",
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.transform = "translateY(-2px)";
                  e.currentTarget.style.boxShadow =
                    "0 8px 20px var(--primary-glow)";
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.transform = "translateY(0)";
                  e.currentTarget.style.boxShadow =
                    "0 4px 12px var(--primary-glow)";
                }}
              >
                新規作成
              </button>
            )}
          </div>
        </div>

        {/* 重要度ガイド */}
        <SeverityGuide />

        <div className="flex gap-6">
          {/* サイドバーフィルターパネル */}
          <IncidentFilters
            search={filters.search}
            severity={filters.severity}
            status={filters.status}
            selectedTagIds={filters.selectedTagIds}
            tags={tags}
            pagination={filters.pagination}
            showSidebar={showSidebar}
            onSearchChange={filters.handleSearchChange}
            onSeverityChange={filters.setSeverity}
            onStatusChange={filters.setStatus}
            onTagToggle={filters.handleTagToggle}
            onClearFilters={filters.clearFilters}
            onApplyPreset={filters.applyPreset}
            onCloseSidebar={() => setShowSidebar(false)}
            setPagination={filters.setPagination}
          />

          {/* メインコンテンツ */}
          <div className="flex-1 min-w-0">
            {/* サイドバー切り替えボタン */}
            {!showSidebar && (
              <button
                onClick={() => setShowSidebar(true)}
                className="mb-4 px-4 py-2 rounded-lg flex items-center border transition-all"
                style={{
                  background: "var(--surface)",
                  borderColor: "var(--border)",
                  color: "var(--foreground)",
                }}
                onMouseEnter={(e) =>
                  (e.currentTarget.style.background = "var(--secondary-light)")
                }
                onMouseLeave={(e) =>
                  (e.currentTarget.style.background = "var(--surface)")
                }
              >
                <svg
                  className="w-5 h-5 mr-2"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
                  />
                </svg>
                フィルターを表示
              </button>
            )}

            {/* フィルターチップ */}
            {filters.hasActiveFilters && (
              <div
                className="mb-4 p-4 rounded-xl shadow-lg border"
                style={{
                  background: "var(--surface)",
                  borderColor: "var(--border)",
                }}
              >
                <div className="flex items-center justify-between mb-2">
                  <h3
                    className="text-sm font-medium"
                    style={{ color: "var(--foreground)" }}
                  >
                    適用中のフィルター:
                  </h3>
                  <button
                    onClick={filters.clearFilters}
                    className="text-sm transition-colors"
                    style={{ color: "var(--primary)" }}
                    onMouseEnter={(e) =>
                      (e.currentTarget.style.color = "var(--primary-hover)")
                    }
                    onMouseLeave={(e) =>
                      (e.currentTarget.style.color = "var(--primary)")
                    }
                  >
                    すべてクリア
                  </button>
                </div>
                <div className="flex flex-wrap gap-2">
                  {filters.search && (
                    <span
                      className="inline-flex items-center px-3 py-1.5 rounded-full text-sm border-2"
                      style={{
                        background: "var(--info-light)",
                        color: "var(--info)",
                        borderColor: "var(--info)",
                      }}
                    >
                      検索: {filters.search}
                      <button
                        onClick={() => filters.clearFilter("search")}
                        className="ml-2 transition-opacity"
                        onMouseEnter={(e) =>
                          (e.currentTarget.style.opacity = "0.7")
                        }
                        onMouseLeave={(e) =>
                          (e.currentTarget.style.opacity = "1")
                        }
                      >
                        <svg
                          className="w-4 h-4"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M6 18L18 6M6 6l12 12"
                          />
                        </svg>
                      </button>
                    </span>
                  )}
                  {filters.severity && (
                    <span
                      className="inline-flex items-center px-3 py-1.5 rounded-full text-sm border-2"
                      style={getSeverityStyle(filters.severity)}
                    >
                      深刻度: {SEVERITY_LABELS[filters.severity]}
                      <button
                        onClick={() => filters.clearFilter("severity")}
                        className="ml-2 transition-opacity"
                        onMouseEnter={(e) =>
                          (e.currentTarget.style.opacity = "0.7")
                        }
                        onMouseLeave={(e) =>
                          (e.currentTarget.style.opacity = "1")
                        }
                      >
                        <svg
                          className="w-4 h-4"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M6 18L18 6M6 6l12 12"
                          />
                        </svg>
                      </button>
                    </span>
                  )}
                  {filters.status && (
                    <span
                      className="inline-flex items-center px-3 py-1.5 rounded-full text-sm border-2"
                      style={getStatusStyle(filters.status)}
                    >
                      ステータス: {STATUS_LABELS[filters.status]}
                      <button
                        onClick={() => filters.clearFilter("status")}
                        className="ml-2 transition-opacity"
                        onMouseEnter={(e) =>
                          (e.currentTarget.style.opacity = "0.7")
                        }
                        onMouseLeave={(e) =>
                          (e.currentTarget.style.opacity = "1")
                        }
                      >
                        <svg
                          className="w-4 h-4"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M6 18L18 6M6 6l12 12"
                          />
                        </svg>
                      </button>
                    </span>
                  )}
                  {filters.selectedTagIds.map((tagId) => {
                    const tag = tags.find((t) => t.id === tagId);
                    return tag ? (
                      <span
                        key={tagId}
                        className="inline-flex items-center px-3 py-1.5 rounded-full text-sm text-white shadow-sm"
                        style={{ backgroundColor: tag.color }}
                      >
                        {tag.name}
                        <button
                          onClick={() => filters.clearFilter("tag", tagId)}
                          className="ml-2 text-white transition-opacity"
                          onMouseEnter={(e) =>
                            (e.currentTarget.style.opacity = "0.7")
                          }
                          onMouseLeave={(e) =>
                            (e.currentTarget.style.opacity = "1")
                          }
                        >
                          <svg
                            className="w-4 h-4"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M6 18L18 6M6 6l12 12"
                            />
                          </svg>
                        </button>
                      </span>
                    ) : null;
                  })}
                </div>
              </div>
            )}

            {/* エラーメッセージ */}
            {error && (
              <div
                className="px-4 py-3 rounded-xl mb-4 border-2"
                style={{
                  background: "var(--error-light)",
                  borderColor: "var(--error)",
                  color: "var(--error)",
                }}
              >
                {error}
              </div>
            )}

            {/* インシデントテーブル */}
            {loading ? (
              <div
                className="text-center py-8"
                style={{ color: "var(--secondary)" }}
              >
                Loading incidents...
              </div>
            ) : incidents.length === 0 ? (
              <div
                className="text-center py-12 px-4 rounded-xl shadow-lg border"
                style={{
                  background: "var(--surface)",
                  borderColor: "var(--border)",
                  color: "var(--secondary)",
                }}
              >
                <svg
                  className="mx-auto h-12 w-12 mb-4"
                  style={{ color: "var(--secondary)" }}
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                  />
                </svg>
                <p
                  className="text-base font-medium"
                  style={{ color: "var(--foreground)" }}
                >
                  インシデントが見つかりませんでした
                </p>
                <p className="text-sm mt-1">
                  フィルターを変更するか、新しいインシデントを作成してください。
                </p>
              </div>
            ) : (
              <>
                <IncidentTable
                  incidents={incidents}
                  sortBy={filters.sortBy}
                  sortOrder={filters.sortOrder}
                  onSort={filters.handleSort}
                />
                <IncidentPagination
                  pagination={filters.pagination}
                  setPagination={filters.setPagination}
                />
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

export default function IncidentsPage() {
  return (
    <Suspense
      fallback={
        <div
          className="min-h-screen flex items-center justify-center"
          style={{ background: "var(--background)" }}
        >
          <div style={{ color: "var(--secondary)" }}>Loading...</div>
        </div>
      }
    >
      <IncidentsPageContent />
    </Suspense>
  );
}
