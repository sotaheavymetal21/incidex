"use client";

import { Severity, Status, PaginationResult } from "@/types/incident";
import { Tag } from "@/types/tag";

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

interface IncidentFiltersProps {
  search: string;
  severity: Severity | "";
  status: Status | "";
  selectedTagIds: number[];
  tags: Tag[];
  pagination: PaginationResult;
  showSidebar: boolean;
  onSearchChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onSeverityChange: (value: Severity | "") => void;
  onStatusChange: (value: Status | "") => void;
  onTagToggle: (tagId: number) => void;
  onClearFilters: () => void;
  onApplyPreset: (preset: "unresolved" | "my-assigned" | "critical") => void;
  onCloseSidebar: () => void;
  setPagination: React.Dispatch<React.SetStateAction<PaginationResult>>;
}

export default function IncidentFilters({
  search,
  severity,
  status,
  selectedTagIds,
  tags,
  pagination,
  showSidebar,
  onSearchChange,
  onSeverityChange,
  onStatusChange,
  onTagToggle,
  onClearFilters,
  onApplyPreset,
  onCloseSidebar,
  setPagination,
}: IncidentFiltersProps) {
  if (!showSidebar) return null;

  return (
    <div className="w-64 flex-shrink-0 animate-slideUp">
      <div
        className="rounded-2xl p-6 sticky top-8 border card-green-accent transition-all duration-200"
        style={{
          background: "var(--surface)",
          borderColor: "var(--border)",
          boxShadow: "var(--shadow-lg)",
        }}
      >
        <div className="flex items-center justify-between mb-6">
          <h2
            className="text-xl font-bold"
            style={{
              color: "var(--foreground)",
              fontFamily: "var(--font-display)",
            }}
          >
            フィルター
          </h2>
          <button
            onClick={onCloseSidebar}
            className="lg:hidden transition-all duration-200 p-1 rounded-lg"
            style={{ color: "var(--secondary)" }}
            onMouseEnter={(e) => {
              e.currentTarget.style.color = "var(--error)";
              e.currentTarget.style.background = "var(--error-light)";
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.color = "var(--secondary)";
              e.currentTarget.style.background = "transparent";
            }}
          >
            <svg
              className="w-5 h-5"
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
        </div>

        {/* フィルタープリセット */}
        <div className="mb-6">
          <h3
            className="text-sm font-semibold mb-3"
            style={{
              color: "var(--foreground)",
              fontFamily: "var(--font-body)",
            }}
          >
            クイックフィルター
          </h3>
          <div className="space-y-2">
            <button
              onClick={() => onApplyPreset("unresolved")}
              className="w-full px-4 py-2.5 text-sm font-semibold text-left rounded-xl transition-all duration-200"
              style={{
                background: "var(--gray-100)",
                color: "var(--foreground)",
                fontFamily: "var(--font-body)",
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = "var(--primary-light)";
                e.currentTarget.style.transform = "translateX(4px)";
                e.currentTarget.style.boxShadow =
                  "0 4px 8px var(--primary-glow)";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = "var(--gray-100)";
                e.currentTarget.style.transform = "translateX(0)";
                e.currentTarget.style.boxShadow = "none";
              }}
            >
              未解決のインシデント
            </button>
            <button
              onClick={() => onApplyPreset("critical")}
              className="w-full px-4 py-2.5 text-sm font-semibold text-left rounded-xl transition-all duration-200"
              style={{
                background: "var(--gray-100)",
                color: "var(--foreground)",
                fontFamily: "var(--font-body)",
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = "var(--primary-light)";
                e.currentTarget.style.transform = "translateX(4px)";
                e.currentTarget.style.boxShadow =
                  "0 4px 8px var(--primary-glow)";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = "var(--gray-100)";
                e.currentTarget.style.transform = "translateX(0)";
                e.currentTarget.style.boxShadow = "none";
              }}
            >
              Critical のみ
            </button>
            <button
              onClick={onClearFilters}
              className="w-full px-4 py-2.5 text-sm font-bold text-left rounded-xl border-2 transition-all duration-200"
              style={{
                borderColor: "var(--border)",
                color: "var(--foreground-secondary)",
                fontFamily: "var(--font-body)",
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = "var(--error-light)";
                e.currentTarget.style.borderColor = "var(--error)";
                e.currentTarget.style.color = "var(--error)";
                e.currentTarget.style.transform = "translateX(4px)";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = "transparent";
                e.currentTarget.style.borderColor = "var(--border)";
                e.currentTarget.style.color = "var(--foreground-secondary)";
                e.currentTarget.style.transform = "translateX(0)";
              }}
            >
              すべてクリア
            </button>
          </div>
        </div>

        {/* 検索 */}
        <div className="mb-6">
          <label
            className="block text-sm font-medium mb-2"
            style={{ color: "var(--foreground)" }}
          >
            検索
          </label>
          <input
            type="text"
            value={search}
            onChange={onSearchChange}
            placeholder="タイトル、説明を検索..."
            className="w-full px-3 py-2.5 border-2 rounded-lg focus:outline-none focus:ring-2 transition-all"
            style={{
              background: "var(--surface)",
              borderColor: "var(--border)",
              color: "var(--foreground)",
            }}
            onFocus={(e) => {
              e.currentTarget.style.borderColor = "var(--primary)";
              e.currentTarget.style.boxShadow =
                "0 0 0 3px var(--primary-light)";
            }}
            onBlur={(e) => {
              e.currentTarget.style.borderColor = "var(--border)";
              e.currentTarget.style.boxShadow = "none";
            }}
          />
        </div>

        {/* 深刻度フィルター */}
        <div className="mb-6">
          <label
            className="block text-sm font-medium mb-2"
            style={{ color: "var(--foreground)" }}
          >
            深刻度
          </label>
          <div className="space-y-2">
            {(Object.keys(SEVERITY_LABELS) as Severity[]).map((sev) => (
              <label key={sev} className="flex items-center cursor-pointer">
                <input
                  type="radio"
                  name="severity"
                  checked={severity === sev}
                  onChange={() => {
                    onSeverityChange(sev);
                    setPagination((prev) => ({ ...prev, page: 1 }));
                  }}
                  className="h-4 w-4 accent-[--primary] border-2"
                  style={{ borderColor: "var(--border)" }}
                />
                <span
                  className="ml-2 text-sm"
                  style={{ color: "var(--foreground)" }}
                >
                  {SEVERITY_LABELS[sev]}
                </span>
              </label>
            ))}
            <label className="flex items-center cursor-pointer">
              <input
                type="radio"
                name="severity"
                checked={severity === ""}
                onChange={() => {
                  onSeverityChange("");
                  setPagination((prev) => ({ ...prev, page: 1 }));
                }}
                className="h-4 w-4 accent-[--primary] border-2"
                style={{ borderColor: "var(--border)" }}
              />
              <span
                className="ml-2 text-sm"
                style={{ color: "var(--foreground)" }}
              >
                すべて
              </span>
            </label>
          </div>
        </div>

        {/* ステータスフィルター */}
        <div className="mb-6">
          <label
            className="block text-sm font-medium mb-2"
            style={{ color: "var(--foreground)" }}
          >
            ステータス
          </label>
          <div className="space-y-2">
            {(Object.keys(STATUS_LABELS) as Status[]).map((st) => (
              <label key={st} className="flex items-center cursor-pointer">
                <input
                  type="radio"
                  name="status"
                  checked={status === st}
                  onChange={() => {
                    onStatusChange(st);
                    setPagination((prev) => ({ ...prev, page: 1 }));
                  }}
                  className="h-4 w-4 accent-[--primary] border-2"
                  style={{ borderColor: "var(--border)" }}
                />
                <span
                  className="ml-2 text-sm"
                  style={{ color: "var(--foreground)" }}
                >
                  {STATUS_LABELS[st]}
                </span>
              </label>
            ))}
            <label className="flex items-center cursor-pointer">
              <input
                type="radio"
                name="status"
                checked={status === ""}
                onChange={() => {
                  onStatusChange("");
                  setPagination((prev) => ({ ...prev, page: 1 }));
                }}
                className="h-4 w-4 accent-[--primary] border-2"
                style={{ borderColor: "var(--border)" }}
              />
              <span
                className="ml-2 text-sm"
                style={{ color: "var(--foreground)" }}
              >
                すべて
              </span>
            </label>
          </div>
        </div>

        {/* タグフィルター */}
        <div>
          <label
            className="block text-sm font-medium mb-2"
            style={{ color: "var(--foreground)" }}
          >
            タグ
          </label>
          <div className="space-y-2">
            {tags.map((tag) => (
              <label key={tag.id} className="flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={selectedTagIds.includes(tag.id)}
                  onChange={() => onTagToggle(tag.id)}
                  className="h-4 w-4 rounded accent-[--primary] border-2"
                  style={{ borderColor: "var(--border)" }}
                />
                <span
                  className="ml-2 text-sm flex-1"
                  style={{ color: "var(--foreground)" }}
                >
                  {tag.name}
                </span>
                <span
                  className="w-3 h-3 rounded-full shadow-sm"
                  style={{ backgroundColor: tag.color }}
                ></span>
              </label>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
