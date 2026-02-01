"use client";

import Timeline from "@/components/Timeline";
import { IncidentActivity } from "@/types/activity";

type TimelineEventType =
  | "detected"
  | "investigation_started"
  | "root_cause_identified"
  | "mitigation"
  | "timeline_resolved"
  | "other";

interface IncidentTimelineProps {
  activities: IncidentActivity[];
  loadingActivities: boolean;
  canEdit: boolean;
  showTimelineEventForm: boolean;
  entryType: "comment" | "event";
  newComment: string;
  timelineEventType: TimelineEventType;
  timelineEventTime: string;
  timelineEventDescription: string;
  submittingComment: boolean;
  submittingTimelineEvent: boolean;
  onToggleForm: () => void;
  onEntryTypeChange: (type: "comment" | "event") => void;
  onCommentChange: (comment: string) => void;
  onEventTypeChange: (type: TimelineEventType) => void;
  onEventTimeChange: (time: string) => void;
  onEventDescriptionChange: (description: string) => void;
  onSubmit: (e: React.FormEvent) => void;
}

export function IncidentTimeline({
  activities,
  loadingActivities,
  canEdit,
  showTimelineEventForm,
  entryType,
  newComment,
  timelineEventType,
  timelineEventTime,
  timelineEventDescription,
  submittingComment,
  submittingTimelineEvent,
  onToggleForm,
  onEntryTypeChange,
  onCommentChange,
  onEventTypeChange,
  onEventTimeChange,
  onEventDescriptionChange,
  onSubmit,
}: IncidentTimelineProps) {
  const isSubmitting = submittingComment || submittingTimelineEvent;
  const isValid =
    entryType === "comment"
      ? newComment.trim()
      : timelineEventDescription.trim();

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
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
          <h2
            className="text-xl font-bold"
            style={{
              color: "var(--foreground)",
              fontFamily: "var(--font-display)",
            }}
          >
            タイムライン & アクティビティ
          </h2>
          {canEdit && (
            <button
              onClick={onToggleForm}
              className="px-4 py-2.5 text-white rounded-xl text-sm font-semibold transition-all duration-200"
              style={{
                background: showTimelineEventForm
                  ? "linear-gradient(135deg, var(--secondary) 0%, var(--secondary-dark) 100%)"
                  : "linear-gradient(135deg, var(--success) 0%, var(--success-dark) 100%)",
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
              {showTimelineEventForm ? "キャンセル" : "+ 追加"}
            </button>
          )}
        </div>

        {/* 統合入力フォーム */}
        {showTimelineEventForm && canEdit && (
          <form
            onSubmit={onSubmit}
            className="mb-6 p-5 rounded-xl border-2 animate-slideDown"
            style={{
              background: "var(--gray-50)",
              borderColor: "var(--border)",
            }}
          >
            <div className="mb-4">
              <label
                htmlFor="entry_type"
                className="block text-sm font-semibold mb-2"
                style={{
                  color: "var(--foreground)",
                  fontFamily: "var(--font-body)",
                }}
              >
                タイプ
              </label>
              <select
                id="entry_type"
                value={entryType}
                onChange={(e) =>
                  onEntryTypeChange(e.target.value as "comment" | "event")
                }
                className="w-full px-3 py-2 border-2 rounded-lg focus:outline-none transition-all"
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
                <option value="comment">💬 コメント</option>
                <option value="event">⏱️ イベント</option>
              </select>
            </div>

            {entryType === "event" && (
              <>
                <div className="mb-4">
                  <label
                    htmlFor="event_type"
                    className="block text-sm font-semibold mb-2"
                    style={{
                      color: "var(--foreground)",
                      fontFamily: "var(--font-body)",
                    }}
                  >
                    イベントタイプ
                  </label>
                  <select
                    id="event_type"
                    value={timelineEventType}
                    onChange={(e) =>
                      onEventTypeChange(e.target.value as TimelineEventType)
                    }
                    className="w-full px-3 py-2 border-2 rounded-lg focus:outline-none transition-all"
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
                    <option value="detected">検知</option>
                    <option value="investigation_started">調査開始</option>
                    <option value="root_cause_identified">原因特定</option>
                    <option value="mitigation">緩和</option>
                    <option value="timeline_resolved">解決</option>
                    <option value="other">その他</option>
                  </select>
                </div>
                <div className="mb-4">
                  <label
                    htmlFor="event_time"
                    className="block text-sm font-semibold mb-2"
                    style={{
                      color: "var(--foreground)",
                      fontFamily: "var(--font-body)",
                    }}
                  >
                    イベント時刻
                  </label>
                  <input
                    type="datetime-local"
                    id="event_time"
                    value={timelineEventTime}
                    onChange={(e) => onEventTimeChange(e.target.value)}
                    className="w-full px-3 py-2 border-2 rounded-lg focus:outline-none transition-all"
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
                    required
                  />
                </div>
              </>
            )}

            <div className="mb-4">
              <label
                htmlFor={
                  entryType === "comment" ? "comment" : "event_description"
                }
                className="block text-sm font-semibold mb-2"
                style={{
                  color: "var(--foreground)",
                  fontFamily: "var(--font-body)",
                }}
              >
                {entryType === "comment" ? "コメント" : "説明"}
              </label>
              <textarea
                id={entryType === "comment" ? "comment" : "event_description"}
                rows={4}
                value={
                  entryType === "comment"
                    ? newComment
                    : timelineEventDescription
                }
                onChange={(e) =>
                  entryType === "comment"
                    ? onCommentChange(e.target.value)
                    : onEventDescriptionChange(e.target.value)
                }
                placeholder={
                  entryType === "comment"
                    ? "コメントを入力してください..."
                    : "イベントの説明を入力してください..."
                }
                className="w-full px-3 py-2 border-2 rounded-lg focus:outline-none transition-all"
                style={{
                  background: "var(--surface)",
                  borderColor: "var(--border)",
                  color: "var(--foreground)",
                  fontFamily: "var(--font-body)",
                }}
                disabled={isSubmitting}
                onFocus={(e) => {
                  e.currentTarget.style.borderColor = "var(--primary)";
                  e.currentTarget.style.boxShadow =
                    "0 0 0 3px var(--primary-light)";
                }}
                onBlur={(e) => {
                  e.currentTarget.style.borderColor = "var(--border)";
                  e.currentTarget.style.boxShadow = "none";
                }}
                required
              />
            </div>
            <div className="flex justify-end">
              <button
                type="submit"
                disabled={isSubmitting || !isValid}
                className="px-5 py-2.5 text-white rounded-xl font-semibold transition-all duration-200 disabled:opacity-50"
                style={{
                  background:
                    entryType === "comment"
                      ? "linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%)"
                      : "linear-gradient(135deg, var(--success) 0%, var(--success-dark) 100%)",
                  fontFamily: "var(--font-body)",
                  boxShadow: "0 4px 12px var(--primary-glow)",
                }}
                onMouseEnter={(e) => {
                  if (!isSubmitting && isValid) {
                    e.currentTarget.style.transform = "translateY(-2px)";
                    e.currentTarget.style.boxShadow =
                      "0 8px 20px var(--primary-glow)";
                  }
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.transform = "translateY(0)";
                  e.currentTarget.style.boxShadow =
                    "0 4px 12px var(--primary-glow)";
                }}
              >
                {isSubmitting
                  ? "送信中..."
                  : entryType === "comment"
                    ? "コメントを投稿"
                    : "イベントを登録"}
              </button>
            </div>
          </form>
        )}

        {/* タイムライン */}
        {loadingActivities ? (
          <div
            className="text-center py-8"
            style={{
              color: "var(--foreground-secondary)",
              fontFamily: "var(--font-body)",
            }}
          >
            読み込み中...
          </div>
        ) : (
          <Timeline activities={activities} />
        )}
      </div>
    </div>
  );
}
