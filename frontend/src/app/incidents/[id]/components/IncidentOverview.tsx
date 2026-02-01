"use client";

import { Incident } from "@/types/incident";
import { User } from "@/types/user";

interface IncidentOverviewProps {
  incident: Incident;
  users: User[];
  canEdit: boolean;
  assigningUser: boolean;
  onAssignIncident: (assigneeId: number | null) => void;
}

export function IncidentOverview({
  incident,
  users,
  canEdit,
  assigningUser,
  onAssignIncident,
}: IncidentOverviewProps) {
  return (
    <div className="space-y-6">
      {/* メタデータセクション */}
      <div
        className="rounded-2xl p-6 border"
        style={{
          background: "var(--gray-50)",
          borderColor: "var(--border)",
        }}
      >
        <h3
          className="text-lg font-bold mb-5"
          style={{
            color: "var(--foreground)",
            fontFamily: "var(--font-display)",
          }}
        >
          メタデータ
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          <div>
            <p
              className="text-sm font-semibold mb-1.5"
              style={{
                color: "var(--foreground-secondary)",
                fontFamily: "var(--font-body)",
              }}
            >
              検出日時
            </p>
            <p
              className="text-sm font-medium"
              style={{
                color: "var(--foreground)",
                fontFamily: "var(--font-mono)",
              }}
            >
              {new Date(incident.detected_at).toLocaleString("ja-JP")}
            </p>
          </div>
          <div>
            <p
              className="text-sm font-semibold mb-1.5"
              style={{
                color: "var(--foreground-secondary)",
                fontFamily: "var(--font-body)",
              }}
            >
              解決日時
            </p>
            <p
              className="text-sm font-medium"
              style={{
                color: "var(--foreground)",
                fontFamily: "var(--font-mono)",
              }}
            >
              {incident.resolved_at
                ? new Date(incident.resolved_at).toLocaleString("ja-JP")
                : "未解決"}
            </p>
          </div>
          <div>
            <p
              className="text-sm font-semibold mb-1.5"
              style={{
                color: "var(--foreground-secondary)",
                fontFamily: "var(--font-body)",
              }}
            >
              担当者
            </p>
            {canEdit ? (
              <select
                value={incident.assignee_id || ""}
                onChange={(e) => {
                  const value = e.target.value;
                  onAssignIncident(value === "" ? null : parseInt(value));
                }}
                disabled={assigningUser}
                className="text-sm font-medium border-2 rounded-lg px-3 py-2 focus:outline-none transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                style={{
                  background: "var(--surface)",
                  borderColor: "var(--border)",
                  color: "var(--foreground)",
                  fontFamily: "var(--font-body)",
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
              >
                <option value="">未割り当て</option>
                {users.map((user) => (
                  <option key={user.id} value={user.id}>
                    {user.name} ({user.email})
                  </option>
                ))}
              </select>
            ) : (
              <p
                className="text-sm font-medium"
                style={{
                  color: "var(--foreground)",
                  fontFamily: "var(--font-body)",
                }}
              >
                {incident.assignee
                  ? `${incident.assignee.name} (${incident.assignee.email})`
                  : "未割り当て"}
              </p>
            )}
          </div>
          <div>
            <p
              className="text-sm font-semibold mb-1.5"
              style={{
                color: "var(--foreground-secondary)",
                fontFamily: "var(--font-body)",
              }}
            >
              作成者
            </p>
            <p
              className="text-sm font-medium"
              style={{
                color: "var(--foreground)",
                fontFamily: "var(--font-body)",
              }}
            >
              {incident.creator.name} ({incident.creator.email})
            </p>
          </div>
          <div>
            <p
              className="text-sm font-semibold mb-1.5"
              style={{
                color: "var(--foreground-secondary)",
                fontFamily: "var(--font-body)",
              }}
            >
              作成日時
            </p>
            <p
              className="text-sm font-medium"
              style={{
                color: "var(--foreground)",
                fontFamily: "var(--font-mono)",
              }}
            >
              {new Date(incident.created_at).toLocaleString("ja-JP")}
            </p>
          </div>
        </div>
      </div>

      {/* 説明セクション */}
      <div
        className="rounded-2xl p-6 border animate-scaleIn"
        style={{
          background: "var(--surface)",
          borderColor: "var(--border)",
          boxShadow: "var(--shadow-lg)",
          animationDelay: "0.1s",
        }}
      >
        <h2
          className="text-xl font-bold mb-4"
          style={{
            color: "var(--foreground)",
            fontFamily: "var(--font-display)",
          }}
        >
          説明
        </h2>
        <p
          className="text-sm leading-relaxed whitespace-pre-wrap"
          style={{
            color: "var(--foreground)",
            fontFamily: "var(--font-body)",
          }}
        >
          {incident.description}
        </p>
      </div>

      {/* 影響範囲セクション */}
      {incident.impact_scope && (
        <div
          className="rounded-2xl p-6 border animate-scaleIn"
          style={{
            background: "var(--surface)",
            borderColor: "var(--border)",
            boxShadow: "var(--shadow-lg)",
            animationDelay: "0.15s",
          }}
        >
          <h2
            className="text-xl font-bold mb-4"
            style={{
              color: "var(--foreground)",
              fontFamily: "var(--font-display)",
            }}
          >
            影響範囲
          </h2>
          <p
            className="text-sm leading-relaxed whitespace-pre-wrap"
            style={{
              color: "var(--foreground)",
              fontFamily: "var(--font-body)",
            }}
          >
            {incident.impact_scope}
          </p>
        </div>
      )}

      {/* タグセクション */}
      <div
        className="rounded-2xl p-6 border animate-scaleIn"
        style={{
          background: "var(--surface)",
          borderColor: "var(--border)",
          boxShadow: "var(--shadow-lg)",
          animationDelay: "0.2s",
        }}
      >
        <h2
          className="text-xl font-bold mb-4"
          style={{
            color: "var(--foreground)",
            fontFamily: "var(--font-display)",
          }}
        >
          タグ
        </h2>
        {incident.tags && incident.tags.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {incident.tags.map((tag) => (
              <span
                key={tag.id}
                className="px-4 py-2 rounded-full text-white text-sm font-semibold shadow-sm"
                style={{
                  backgroundColor: tag.color,
                  fontFamily: "var(--font-body)",
                }}
              >
                {tag.name}
              </span>
            ))}
          </div>
        ) : (
          <p
            className="text-sm"
            style={{
              color: "var(--foreground-secondary)",
              fontFamily: "var(--font-body)",
            }}
          >
            タグが設定されていません
          </p>
        )}
      </div>
    </div>
  );
}
