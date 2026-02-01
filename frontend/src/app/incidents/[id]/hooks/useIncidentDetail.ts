import { useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import { incidentApi, userApi, exportApi } from "@/lib/api";
import { Incident } from "@/types/incident";
import { User } from "@/types/user";

interface UseIncidentDetailOptions {
  token: string | null;
  id: string;
  canEdit: boolean;
}

interface UseIncidentDetailReturn {
  incident: Incident | null;
  loading: boolean;
  error: string;
  users: User[];
  assigningUser: boolean;
  deleting: boolean;
  exportingPDF: boolean;
  fetchIncident: () => Promise<void>;
  handleAssignIncident: (assigneeId: number | null) => Promise<void>;
  handleDelete: () => Promise<void>;
  handleExportPDF: () => Promise<void>;
}

export function useIncidentDetail({
  token,
  id,
  canEdit,
}: UseIncidentDetailOptions): UseIncidentDetailReturn {
  const router = useRouter();
  const [incident, setIncident] = useState<Incident | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [users, setUsers] = useState<User[]>([]);
  const [assigningUser, setAssigningUser] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [exportingPDF, setExportingPDF] = useState(false);

  const fetchIncident = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    setError("");
    try {
      const data = await incidentApi.getById(token, parseInt(id));
      setIncident(data);
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : "Failed to fetch incident";
      setError(message);
    } finally {
      setLoading(false);
    }
  }, [token, id]);

  const fetchUsers = useCallback(async () => {
    if (!token) return;
    try {
      const data = await userApi.getAll(token);
      setUsers(data);
    } catch (err: unknown) {
      console.error("Failed to fetch users:", err);
    }
  }, [token]);

  useEffect(() => {
    if (token && id) {
      fetchIncident();
    }
  }, [token, id, fetchIncident]);

  useEffect(() => {
    if (token && canEdit) {
      fetchUsers();
    }
  }, [token, canEdit, fetchUsers]);

  const handleAssignIncident = useCallback(
    async (assigneeId: number | null) => {
      if (!token) return;
      setAssigningUser(true);
      try {
        const updatedIncident = await incidentApi.assignIncident(
          token,
          parseInt(id),
          assigneeId,
        );
        setIncident(updatedIncident);
        await fetchIncident();
      } catch (err: unknown) {
        const message =
          err instanceof Error ? err.message : "担当者の変更に失敗しました";
        alert(message);
      } finally {
        setAssigningUser(false);
      }
    },
    [token, id, fetchIncident],
  );

  const handleDelete = useCallback(async () => {
    if (!token) return;
    if (!confirm("Are you sure you want to delete this incident?")) {
      return;
    }

    setDeleting(true);
    try {
      await incidentApi.delete(token, parseInt(id));
      router.push("/incidents");
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : "Failed to delete incident";
      alert(message);
    } finally {
      setDeleting(false);
    }
  }, [token, id, router]);

  const handleExportPDF = useCallback(async () => {
    if (!token) return;
    setExportingPDF(true);
    try {
      await exportApi.exportIncidentPDF(token, parseInt(id));
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : "PDFのエクスポートに失敗しました";
      alert(message);
    } finally {
      setExportingPDF(false);
    }
  }, [token, id]);

  return {
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
  };
}
