import { useState, useEffect, useCallback } from "react";
import { activityApi } from "@/lib/api";
import { IncidentActivity } from "@/types/activity";

type TimelineEventType =
  | "detected"
  | "investigation_started"
  | "root_cause_identified"
  | "mitigation"
  | "timeline_resolved"
  | "other";

interface UseActivitiesOptions {
  token: string | null;
  incidentId: string;
}

interface UseActivitiesReturn {
  activities: IncidentActivity[];
  loadingActivities: boolean;
  newComment: string;
  setNewComment: (comment: string) => void;
  submittingComment: boolean;
  showTimelineEventForm: boolean;
  setShowTimelineEventForm: (show: boolean) => void;
  entryType: "comment" | "event";
  setEntryType: (type: "comment" | "event") => void;
  timelineEventType: TimelineEventType;
  setTimelineEventType: (type: TimelineEventType) => void;
  timelineEventTime: string;
  setTimelineEventTime: (time: string) => void;
  timelineEventDescription: string;
  setTimelineEventDescription: (description: string) => void;
  submittingTimelineEvent: boolean;
  fetchActivities: () => Promise<void>;
  handleAddComment: (e: React.FormEvent) => Promise<void>;
  handleAddTimelineEvent: (e: React.FormEvent) => Promise<void>;
  initializeForm: () => void;
}

export function useActivities({
  token,
  incidentId,
}: UseActivitiesOptions): UseActivitiesReturn {
  const [activities, setActivities] = useState<IncidentActivity[]>([]);
  const [loadingActivities, setLoadingActivities] = useState(true);
  const [newComment, setNewComment] = useState("");
  const [submittingComment, setSubmittingComment] = useState(false);
  const [showTimelineEventForm, setShowTimelineEventForm] = useState(false);
  const [entryType, setEntryType] = useState<"comment" | "event">("comment");
  const [timelineEventType, setTimelineEventType] =
    useState<TimelineEventType>("other");
  const [timelineEventTime, setTimelineEventTime] = useState("");
  const [timelineEventDescription, setTimelineEventDescription] = useState("");
  const [submittingTimelineEvent, setSubmittingTimelineEvent] = useState(false);

  const fetchActivities = useCallback(async () => {
    if (!token) return;
    setLoadingActivities(true);
    try {
      const data = await activityApi.getActivities(token, parseInt(incidentId));
      setActivities(data);
    } catch (err: unknown) {
      console.error("Failed to fetch activities:", err);
    } finally {
      setLoadingActivities(false);
    }
  }, [token, incidentId]);

  useEffect(() => {
    if (token && incidentId) {
      fetchActivities();
    }
  }, [token, incidentId, fetchActivities]);

  const initializeForm = useCallback(() => {
    const now = new Date();
    const localDateTime = new Date(
      now.getTime() - now.getTimezoneOffset() * 60000,
    )
      .toISOString()
      .slice(0, 16);
    setTimelineEventTime(localDateTime);
    setEntryType("comment");
    setNewComment("");
    setTimelineEventDescription("");
  }, []);

  const handleAddComment = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      if (!token || !newComment.trim()) return;

      setSubmittingComment(true);
      try {
        await activityApi.addComment(token, parseInt(incidentId), {
          comment: newComment,
        });
        setNewComment("");
        setShowTimelineEventForm(false);
        await fetchActivities();
      } catch (err: unknown) {
        const message =
          err instanceof Error ? err.message : "Failed to add comment";
        alert(message);
      } finally {
        setSubmittingComment(false);
      }
    },
    [token, incidentId, newComment, fetchActivities],
  );

  const handleAddTimelineEvent = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      if (!token || !timelineEventDescription.trim()) return;

      setSubmittingTimelineEvent(true);
      try {
        await activityApi.addTimelineEvent(token, parseInt(incidentId), {
          event_type: timelineEventType,
          event_time: timelineEventTime
            ? new Date(timelineEventTime).toISOString()
            : new Date().toISOString(),
          description: timelineEventDescription,
        });
        setTimelineEventType("other");
        setTimelineEventTime("");
        setTimelineEventDescription("");
        setShowTimelineEventForm(false);
        await fetchActivities();
      } catch (err: unknown) {
        const message =
          err instanceof Error
            ? err.message
            : "タイムラインイベントの追加に失敗しました";
        alert(message);
      } finally {
        setSubmittingTimelineEvent(false);
      }
    },
    [
      token,
      incidentId,
      timelineEventType,
      timelineEventTime,
      timelineEventDescription,
      fetchActivities,
    ],
  );

  return {
    activities,
    loadingActivities,
    newComment,
    setNewComment,
    submittingComment,
    showTimelineEventForm,
    setShowTimelineEventForm,
    entryType,
    setEntryType,
    timelineEventType,
    setTimelineEventType,
    timelineEventTime,
    setTimelineEventTime,
    timelineEventDescription,
    setTimelineEventDescription,
    submittingTimelineEvent,
    fetchActivities,
    handleAddComment,
    handleAddTimelineEvent,
    initializeForm,
  };
}
