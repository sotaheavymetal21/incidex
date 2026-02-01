"use client";

import { useRouter } from "next/navigation";
import { Incident } from "@/types/incident";
import { getSeverityStyle, getStatusStyle } from "../utils";

interface IncidentHeaderProps {
  incident: Incident;
  canEdit: boolean;
  canDelete: boolean;
  deleting: boolean;
  exportingPDF: boolean;
  onDelete: () => void;
  onExportPDF: () => void;
}

export function IncidentHeader({
  incident,
  canEdit,
  canDelete,
  deleting,
  exportingPDF,
  onDelete,
  onExportPDF,
}: IncidentHeaderProps) {
  const router = useRouter();
  const severityStyle = getSeverityStyle(incident.severity);
  const statusStyle = getStatusStyle(incident.status);

  return (
    <div className="mb-6 animate-slideDown">
      <button
        onClick={() => router.push("/incidents")}
        className="mb-4 inline-flex items-center font-semibold transition-all"
        style={{ color: "var(--primary)", fontFamily: "var(--font-body)" }}
        onMouseEnter={(e) => {
          e.currentTarget.style.color = "var(--primary-hover)";
          e.currentTarget.style.transform = "translateX(-4px)";
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.color = "var(--primary)";
          e.currentTarget.style.transform = "translateX(0)";
        }}
      >
        ← インシデント一覧に戻る
      </button>

      <div className="flex flex-col lg:flex-row lg:justify-between lg:items-start gap-4">
        <div className="flex-1">
          <h1
            className="text-3xl md:text-4xl font-bold mb-3"
            style={{
              color: "var(--foreground)",
              fontFamily: "var(--font-display)",
            }}
          >
            {incident.title}
          </h1>
          <div className="flex flex-wrap gap-2">
            <span
              className="px-3 py-1.5 inline-flex text-sm font-bold rounded-full border-2"
              style={{
                background: severityStyle.background,
                color: severityStyle.color,
                borderColor: severityStyle.borderColor,
                fontFamily: "var(--font-body)",
              }}
            >
              {incident.severity.toUpperCase()}
            </span>
            <span
              className="px-3 py-1.5 inline-flex text-sm font-bold rounded-full border-2"
              style={{
                background: statusStyle.background,
                color: statusStyle.color,
                borderColor: statusStyle.borderColor,
                fontFamily: "var(--font-body)",
              }}
            >
              {incident.status.charAt(0).toUpperCase() +
                incident.status.slice(1)}
            </span>
          </div>
        </div>

        <div className="flex flex-wrap gap-2">
          <button
            onClick={onExportPDF}
            disabled={exportingPDF}
            className="px-4 py-2.5 text-white rounded-xl font-semibold transition-all duration-200 disabled:opacity-50"
            style={{
              background: "linear-gradient(135deg, #6366f1 0%, #4f46e5 100%)",
              fontFamily: "var(--font-body)",
              boxShadow: "0 4px 12px rgba(99, 102, 241, 0.3)",
            }}
            onMouseEnter={(e) => {
              if (!exportingPDF) {
                e.currentTarget.style.transform = "translateY(-2px)";
                e.currentTarget.style.boxShadow =
                  "0 8px 20px rgba(99, 102, 241, 0.4)";
              }
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = "translateY(0)";
              e.currentTarget.style.boxShadow =
                "0 4px 12px rgba(99, 102, 241, 0.3)";
            }}
          >
            {exportingPDF ? "PDF生成中..." : "PDF"}
          </button>
          <button
            onClick={() => router.push(`/incidents/${incident.id}/postmortem`)}
            className="px-4 py-2.5 text-white rounded-xl font-semibold transition-all duration-200"
            style={{
              background:
                "linear-gradient(135deg, var(--accent) 0%, var(--accent-hover) 100%)",
              fontFamily: "var(--font-body)",
              boxShadow: "0 4px 12px var(--accent-glow)",
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = "translateY(-2px)";
              e.currentTarget.style.boxShadow = "0 8px 20px var(--accent-glow)";
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = "translateY(0)";
              e.currentTarget.style.boxShadow = "0 4px 12px var(--accent-glow)";
            }}
          >
            Post-Mortem
          </button>
          {canEdit && (
            <button
              onClick={() => router.push(`/incidents/${incident.id}/edit`)}
              className="px-4 py-2.5 text-white rounded-xl font-semibold transition-all duration-200"
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
              編集
            </button>
          )}
          {canDelete && (
            <button
              onClick={onDelete}
              disabled={deleting}
              className="px-4 py-2.5 text-white rounded-xl font-semibold transition-all duration-200 disabled:opacity-50"
              style={{
                background:
                  "linear-gradient(135deg, var(--error) 0%, var(--error-dark) 100%)",
                fontFamily: "var(--font-body)",
                boxShadow: "0 4px 12px var(--error-glow)",
              }}
              onMouseEnter={(e) => {
                if (!deleting) {
                  e.currentTarget.style.transform = "translateY(-2px)";
                  e.currentTarget.style.boxShadow =
                    "0 8px 20px var(--error-glow)";
                }
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.transform = "translateY(0)";
                e.currentTarget.style.boxShadow =
                  "0 4px 12px var(--error-glow)";
              }}
            >
              {deleting ? "削除中..." : "削除"}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
