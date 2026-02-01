"use client";

import { useEffect, useState } from "react";
import { useRouter, useParams, useSearchParams } from "next/navigation";
import { useAuth } from "@/context/AuthContext";
import { usePermissions } from "@/hooks/usePermissions";
import Tabs from "@/components/Tabs";
import {
  IncidentHeader,
  IncidentOverview,
  IncidentAttachments,
  IncidentTimeline,
  ImageLightbox,
} from "./components";
import { useIncidentDetail, useActivities, useAttachments } from "./hooks";

const VALID_TABS = ["overview", "timeline", "attachments"] as const;
type TabId = (typeof VALID_TABS)[number];

export default function IncidentDetailPage() {
  const { token, user, loading: authLoading } = useAuth();
  const permissions = usePermissions();
  const router = useRouter();
  const params = useParams();
  const searchParams = useSearchParams();
  const id = params.id as string;

  // URL parameter for tab
  const tabFromUrl = searchParams.get("tab") as TabId | null;
  const [currentTab, setCurrentTab] = useState<TabId>(
    tabFromUrl && VALID_TABS.includes(tabFromUrl) ? tabFromUrl : "overview",
  );

  // Custom hooks
  const {
    incident,
    loading,
    error,
    users,
    assigningUser,
    deleting,
    exportingPDF,
    fetchIncident,
    handleAssignIncident,
    handleDelete,
    handleExportPDF,
  } = useIncidentDetail({ token, id, canEdit: permissions.canEdit });

  const activitiesHook = useActivities({ token, incidentId: id });
  const attachmentsHook = useAttachments({ token, incidentId: id });

  // Tab change handler
  const handleTabChange = (tabId: string) => {
    const newTab = tabId as TabId;
    setCurrentTab(newTab);
    const newUrl = new URL(window.location.href);
    newUrl.searchParams.set("tab", newTab);
    router.replace(newUrl.pathname + newUrl.search, { scroll: false });
  };

  // Auth redirect
  useEffect(() => {
    if (!authLoading && !token) {
      router.push("/login");
    }
  }, [token, authLoading, router]);

  // Scroll to top on id change
  useEffect(() => {
    window.scrollTo(0, 0);
  }, [id]);

  // Permission checks
  const canEdit = () => {
    if (!user || !incident) return false;
    return (
      user.role === "admin" ||
      (user.role === "editor" && incident.creator_id === user.id)
    );
  };

  const canDelete = () => user?.role === "admin";

  // Loading states
  if (authLoading || !token) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-600">Loading...</div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-600">Loading incident...</div>
      </div>
    );
  }

  if (error || !incident) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-red-600">{error || "Incident not found"}</div>
      </div>
    );
  }

  // Form toggle handler
  const handleToggleForm = () => {
    activitiesHook.setShowTimelineEventForm(
      !activitiesHook.showTimelineEventForm,
    );
    if (!activitiesHook.showTimelineEventForm) {
      activitiesHook.initializeForm();
    }
  };

  // Submit handler for timeline/comment
  const handleFormSubmit = (e: React.FormEvent) => {
    if (activitiesHook.entryType === "comment") {
      activitiesHook.handleAddComment(e);
    } else {
      activitiesHook.handleAddTimelineEvent(e);
    }
    fetchIncident();
  };

  return (
    <div
      className="min-h-screen py-8 px-4 sm:px-6 lg:px-8"
      style={{ background: "var(--background)" }}
    >
      <div className="max-w-6xl mx-auto">
        <IncidentHeader
          incident={incident}
          canEdit={canEdit()}
          canDelete={canDelete()}
          deleting={deleting}
          exportingPDF={exportingPDF}
          onDelete={handleDelete}
          onExportPDF={handleExportPDF}
        />

        <div
          className="rounded-2xl p-6 border animate-fadeIn"
          style={{
            background: "var(--surface)",
            borderColor: "var(--border)",
            boxShadow: "var(--shadow-lg)",
          }}
        >
          <Tabs
            activeTab={currentTab}
            onChange={handleTabChange}
            tabs={[
              {
                id: "overview",
                label: "概要",
                icon: (
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
                      d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                    />
                  </svg>
                ),
                content: (
                  <IncidentOverview
                    incident={incident}
                    users={users}
                    canEdit={permissions.canEdit}
                    assigningUser={assigningUser}
                    onAssignIncident={handleAssignIncident}
                  />
                ),
              },
              {
                id: "timeline",
                label: "タイムライン",
                icon: (
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
                      d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                ),
                content: (
                  <IncidentTimeline
                    activities={activitiesHook.activities}
                    loadingActivities={activitiesHook.loadingActivities}
                    canEdit={canEdit()}
                    showTimelineEventForm={activitiesHook.showTimelineEventForm}
                    entryType={activitiesHook.entryType}
                    newComment={activitiesHook.newComment}
                    timelineEventType={activitiesHook.timelineEventType}
                    timelineEventTime={activitiesHook.timelineEventTime}
                    timelineEventDescription={
                      activitiesHook.timelineEventDescription
                    }
                    submittingComment={activitiesHook.submittingComment}
                    submittingTimelineEvent={
                      activitiesHook.submittingTimelineEvent
                    }
                    onToggleForm={handleToggleForm}
                    onEntryTypeChange={activitiesHook.setEntryType}
                    onCommentChange={activitiesHook.setNewComment}
                    onEventTypeChange={activitiesHook.setTimelineEventType}
                    onEventTimeChange={activitiesHook.setTimelineEventTime}
                    onEventDescriptionChange={
                      activitiesHook.setTimelineEventDescription
                    }
                    onSubmit={handleFormSubmit}
                  />
                ),
              },
              {
                id: "attachments",
                label: "添付ファイル",
                icon: (
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
                      d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"
                    />
                  </svg>
                ),
                content: (
                  <IncidentAttachments
                    attachments={attachmentsHook.attachments}
                    loadingAttachments={attachmentsHook.loadingAttachments}
                    uploadingFile={attachmentsHook.uploadingFile}
                    imageUrls={attachmentsHook.imageUrls}
                    currentUser={user}
                    onFileUpload={attachmentsHook.handleFileUpload}
                    onFileDownload={attachmentsHook.handleFileDownload}
                    onFileDelete={attachmentsHook.handleFileDelete}
                    onImageClick={attachmentsHook.setLightboxImage}
                    isImageFile={attachmentsHook.isImageFile}
                    formatFileSize={attachmentsHook.formatFileSize}
                  />
                ),
              },
            ]}
          />
        </div>
      </div>

      <ImageLightbox
        imageUrl={attachmentsHook.lightboxImage}
        onClose={() => attachmentsHook.setLightboxImage(null)}
      />
    </div>
  );
}
