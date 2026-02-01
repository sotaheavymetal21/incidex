import { useState, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { Severity, Status, PaginationResult } from "@/types/incident";

export type SortField =
  | "detected_at"
  | "resolved_at"
  | "severity"
  | "status"
  | "title";
export type SortOrder = "asc" | "desc";

export interface IncidentFiltersState {
  search: string;
  severity: Severity | "";
  status: Status | "";
  selectedTagIds: number[];
  sortBy: SortField;
  sortOrder: SortOrder;
  pagination: PaginationResult;
}

export interface IncidentFiltersActions {
  setSearch: (value: string) => void;
  setSeverity: (value: Severity | "") => void;
  setStatus: (value: Status | "") => void;
  setSelectedTagIds: React.Dispatch<React.SetStateAction<number[]>>;
  setSortBy: (value: SortField) => void;
  setSortOrder: (value: SortOrder) => void;
  setPagination: React.Dispatch<React.SetStateAction<PaginationResult>>;
  handleSearchChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  handleTagToggle: (tagId: number) => void;
  handleSort: (field: SortField) => void;
  clearFilters: () => void;
  clearFilter: (
    filterType: "search" | "severity" | "status" | "tag",
    value?: number,
  ) => void;
  applyPreset: (preset: "unresolved" | "my-assigned" | "critical") => void;
  hasActiveFilters: boolean;
}

export function useIncidentFilters(): IncidentFiltersState &
  IncidentFiltersActions {
  const searchParams = useSearchParams();

  const [search, setSearch] = useState("");
  const [severity, setSeverity] = useState<Severity | "">("");
  const [status, setStatus] = useState<Status | "">("");
  const [selectedTagIds, setSelectedTagIds] = useState<number[]>([]);
  const [sortBy, setSortBy] = useState<SortField>("detected_at");
  const [sortOrder, setSortOrder] = useState<SortOrder>("desc");
  const [pagination, setPagination] = useState<PaginationResult>({
    page: 1,
    limit: 20,
    total: 0,
    total_pages: 0,
  });

  // URL パラメータからフィルターを読み込み
  useEffect(() => {
    const paramSeverity = searchParams.get("severity") as Severity | null;
    const paramStatus = searchParams.get("status") as Status | null;
    const paramTags = searchParams.get("tags");
    const paramSearch = searchParams.get("search");

    if (paramSeverity) setSeverity(paramSeverity);
    if (paramStatus) setStatus(paramStatus);
    if (paramTags) setSelectedTagIds(paramTags.split(",").map(Number));
    if (paramSearch) setSearch(paramSearch);
  }, [searchParams]);

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearch(e.target.value);
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const handleTagToggle = (tagId: number) => {
    setSelectedTagIds((prev) =>
      prev.includes(tagId)
        ? prev.filter((id) => id !== tagId)
        : [...prev, tagId],
    );
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const handleSort = (field: SortField) => {
    if (sortBy === field) {
      setSortOrder(sortOrder === "asc" ? "desc" : "asc");
    } else {
      setSortBy(field);
      setSortOrder("desc");
    }
  };

  const clearFilters = () => {
    setSearch("");
    setSeverity("");
    setStatus("");
    setSelectedTagIds([]);
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const clearFilter = (
    filterType: "search" | "severity" | "status" | "tag",
    value?: number,
  ) => {
    switch (filterType) {
      case "search":
        setSearch("");
        break;
      case "severity":
        setSeverity("");
        break;
      case "status":
        setStatus("");
        break;
      case "tag":
        if (value !== undefined) {
          setSelectedTagIds((prev) => prev.filter((id) => id !== value));
        }
        break;
    }
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const applyPreset = (preset: "unresolved" | "my-assigned" | "critical") => {
    clearFilters();
    switch (preset) {
      case "unresolved":
        setStatus("open");
        break;
      case "my-assigned":
        // 担当者フィルターのAPIサポートが必要
        break;
      case "critical":
        setSeverity("critical");
        break;
    }
  };

  const hasActiveFilters =
    search !== "" ||
    severity !== "" ||
    status !== "" ||
    selectedTagIds.length > 0;

  return {
    search,
    severity,
    status,
    selectedTagIds,
    sortBy,
    sortOrder,
    pagination,
    setSearch,
    setSeverity,
    setStatus,
    setSelectedTagIds,
    setSortBy,
    setSortOrder,
    setPagination,
    handleSearchChange,
    handleTagToggle,
    handleSort,
    clearFilters,
    clearFilter,
    applyPreset,
    hasActiveFilters,
  };
}
