"use client";

import { useRouter } from "next/navigation";
import { Incident, Severity, Status } from "@/types/incident";
import { SortField, SortOrder } from "../hooks/useIncidentFilters";

interface IncidentTableProps {
  incidents: Incident[];
  sortBy: SortField;
  sortOrder: SortOrder;
  onSort: (field: SortField) => void;
}

const getSeverityStyle = (severity: Severity) => {
  switch (severity) {
    case "critical":
      return {
        background: "var(--critical-light)",
        color: "var(--critical)",
        borderColor: "var(--critical)",
      };
    case "high":
      return {
        background: "var(--high-light)",
        color: "var(--high)",
        borderColor: "var(--high)",
      };
    case "medium":
      return {
        background: "var(--medium-light)",
        color: "var(--medium)",
        borderColor: "var(--medium)",
      };
    case "low":
      return {
        background: "var(--low-light)",
        color: "var(--low)",
        borderColor: "var(--low)",
      };
    default:
      return {
        background: "var(--gray-100)",
        color: "var(--gray-700)",
        borderColor: "var(--gray-300)",
      };
  }
};

const getStatusStyle = (status: Status) => {
  switch (status) {
    case "open":
      return {
        background: "var(--status-open-light)",
        color: "var(--status-open)",
        borderColor: "var(--status-open)",
      };
    case "investigating":
      return {
        background: "var(--status-investigating-light)",
        color: "var(--status-investigating)",
        borderColor: "var(--status-investigating)",
      };
    case "resolved":
      return {
        background: "var(--status-resolved-light)",
        color: "var(--status-resolved)",
        borderColor: "var(--status-resolved)",
      };
    case "closed":
      return {
        background: "var(--status-closed-light)",
        color: "var(--status-closed)",
        borderColor: "var(--status-closed)",
      };
    default:
      return {
        background: "var(--gray-100)",
        color: "var(--gray-700)",
        borderColor: "var(--gray-300)",
      };
  }
};

// ソート関数
const sortIncidents = (
  incidentsToSort: Incident[],
  sortBy: SortField,
  sortOrder: SortOrder,
): Incident[] => {
  const sorted = [...incidentsToSort].sort((a, b) => {
    let comparison = 0;

    switch (sortBy) {
      case "detected_at":
        comparison =
          new Date(a.detected_at).getTime() - new Date(b.detected_at).getTime();
        break;
      case "resolved_at":
        if (!a.resolved_at && !b.resolved_at) comparison = 0;
        else if (!a.resolved_at) comparison = 1;
        else if (!b.resolved_at) comparison = -1;
        else
          comparison =
            new Date(a.resolved_at).getTime() -
            new Date(b.resolved_at).getTime();
        break;
      case "severity":
        const severityOrder = { critical: 0, high: 1, medium: 2, low: 3 };
        comparison = severityOrder[a.severity] - severityOrder[b.severity];
        break;
      case "status":
        const statusOrder = {
          open: 0,
          investigating: 1,
          resolved: 2,
          closed: 3,
        };
        comparison = statusOrder[a.status] - statusOrder[b.status];
        break;
      case "title":
        comparison = a.title.localeCompare(b.title);
        break;
    }

    return sortOrder === "asc" ? comparison : -comparison;
  });

  return sorted;
};

export default function IncidentTable({
  incidents,
  sortBy,
  sortOrder,
  onSort,
}: IncidentTableProps) {
  const router = useRouter();
  const sortedIncidents = sortIncidents(incidents, sortBy, sortOrder);

  const SortButton = ({
    field,
    label,
  }: {
    field: SortField;
    label: string;
  }) => (
    <button
      onClick={() => onSort(field)}
      className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider transition-colors"
      style={{
        color: sortBy === field ? "var(--primary)" : "var(--foreground)",
      }}
      onMouseEnter={(e) => (e.currentTarget.style.color = "var(--primary)")}
      onMouseLeave={(e) =>
        (e.currentTarget.style.color =
          sortBy === field ? "var(--primary)" : "var(--foreground)")
      }
    >
      {label}
      {sortBy === field && (
        <span className="text-sm">{sortOrder === "asc" ? "↑" : "↓"}</span>
      )}
    </button>
  );

  return (
    <div
      className="rounded-xl shadow-lg overflow-hidden border"
      style={{ background: "var(--surface)", borderColor: "var(--border)" }}
    >
      <table className="min-w-full">
        <thead style={{ background: "var(--secondary-light)" }}>
          <tr>
            <th className="px-6 py-3.5 text-left">
              <SortButton field="title" label="Title" />
            </th>
            <th className="px-6 py-3.5 text-left">
              <SortButton field="severity" label="Severity" />
            </th>
            <th className="px-6 py-3.5 text-left">
              <SortButton field="status" label="Status" />
            </th>
            <th className="px-6 py-3.5 text-left">
              <SortButton field="detected_at" label="Detected At" />
            </th>
            <th
              className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wider"
              style={{ color: "var(--foreground)" }}
            >
              Assignee
            </th>
            <th
              className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wider"
              style={{ color: "var(--foreground)" }}
            >
              Tags
            </th>
          </tr>
        </thead>
        <tbody>
          {sortedIncidents.map((incident, index) => (
            <tr
              key={incident.id}
              onClick={() => router.push(`/incidents/${incident.id}`)}
              className="cursor-pointer transition-all"
              style={{
                borderTop: index > 0 ? `1px solid var(--border)` : "none",
              }}
              onMouseEnter={(e) =>
                (e.currentTarget.style.background = "var(--secondary-light)")
              }
              onMouseLeave={(e) =>
                (e.currentTarget.style.background = "transparent")
              }
            >
              <td className="px-6 py-4">
                <div
                  className="text-sm font-medium"
                  style={{ color: "var(--foreground)" }}
                >
                  {incident.title}
                </div>
              </td>
              <td className="px-6 py-4 whitespace-nowrap">
                <span
                  className="px-2.5 py-1 inline-flex text-xs leading-5 font-semibold rounded-full border-2"
                  style={getSeverityStyle(incident.severity)}
                >
                  {incident.severity.toUpperCase()}
                </span>
              </td>
              <td className="px-6 py-4 whitespace-nowrap">
                <span
                  className="px-2.5 py-1 inline-flex text-xs leading-5 font-semibold rounded-full border-2"
                  style={getStatusStyle(incident.status)}
                >
                  {incident.status.charAt(0).toUpperCase() +
                    incident.status.slice(1)}
                </span>
              </td>
              <td
                className="px-6 py-4 whitespace-nowrap text-sm"
                style={{ color: "var(--secondary)" }}
              >
                {new Date(incident.detected_at).toLocaleString()}
              </td>
              <td
                className="px-6 py-4 whitespace-nowrap text-sm"
                style={{ color: "var(--secondary)" }}
              >
                {incident.assignee?.name || "-"}
              </td>
              <td className="px-6 py-4">
                <div className="flex flex-wrap gap-1.5 max-w-xs">
                  {incident.tags?.map((tag) => (
                    <span
                      key={tag.id}
                      className="px-2.5 py-1 text-xs rounded-full text-white whitespace-nowrap shadow-sm"
                      style={{ backgroundColor: tag.color }}
                    >
                      {tag.name}
                    </span>
                  )) || "-"}
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
