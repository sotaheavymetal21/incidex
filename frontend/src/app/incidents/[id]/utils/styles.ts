import { Severity, Status } from "@/types/incident";

interface StyleConfig {
  background: string;
  color: string;
  borderColor: string;
}

export function getSeverityStyle(severity: Severity): StyleConfig {
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
}

export function getStatusStyle(status: Status): StyleConfig {
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
}

export function getSeverityColor(severity: Severity): string {
  switch (severity) {
    case "critical":
      return "bg-red-100 text-red-800 border-red-300";
    case "high":
      return "bg-orange-100 text-orange-800 border-orange-300";
    case "medium":
      return "bg-yellow-100 text-yellow-800 border-yellow-300";
    case "low":
      return "bg-green-100 text-green-800 border-green-300";
    default:
      return "bg-gray-100 text-gray-800 border-gray-300";
  }
}

export function getStatusColor(status: Status): string {
  switch (status) {
    case "open":
      return "bg-gray-100 text-gray-800 border-gray-300";
    case "investigating":
      return "bg-blue-100 text-blue-800 border-blue-300";
    case "resolved":
      return "bg-green-100 text-green-800 border-green-300";
    case "closed":
      return "bg-purple-100 text-purple-800 border-purple-300";
    default:
      return "bg-gray-100 text-gray-800 border-gray-300";
  }
}
