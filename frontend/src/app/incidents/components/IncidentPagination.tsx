"use client";

import { PaginationResult } from "@/types/incident";

interface IncidentPaginationProps {
  pagination: PaginationResult;
  setPagination: React.Dispatch<React.SetStateAction<PaginationResult>>;
}

export default function IncidentPagination({
  pagination,
  setPagination,
}: IncidentPaginationProps) {
  const goToPrevious = () => {
    setPagination((prev) => ({ ...prev, page: Math.max(prev.page - 1, 1) }));
  };

  const goToNext = () => {
    setPagination((prev) => ({
      ...prev,
      page: Math.min(prev.page + 1, prev.total_pages),
    }));
  };

  return (
    <div
      className="px-4 py-3.5 flex items-center justify-between border-t sm:px-6"
      style={{
        background: "var(--secondary-light)",
        borderColor: "var(--border)",
      }}
    >
      {/* モバイル表示 */}
      <div className="flex-1 flex justify-between sm:hidden">
        <button
          onClick={goToPrevious}
          disabled={pagination.page === 1}
          className="relative inline-flex items-center px-4 py-2 border-2 text-sm font-medium rounded-lg transition-all disabled:opacity-40"
          style={{
            background: "var(--surface)",
            borderColor: "var(--border)",
            color: "var(--foreground)",
          }}
          onMouseEnter={(e) =>
            !e.currentTarget.disabled &&
            (e.currentTarget.style.borderColor = "var(--primary)")
          }
          onMouseLeave={(e) =>
            (e.currentTarget.style.borderColor = "var(--border)")
          }
        >
          Previous
        </button>
        <button
          onClick={goToNext}
          disabled={pagination.page === pagination.total_pages}
          className="ml-3 relative inline-flex items-center px-4 py-2 border-2 text-sm font-medium rounded-lg transition-all disabled:opacity-40"
          style={{
            background: "var(--surface)",
            borderColor: "var(--border)",
            color: "var(--foreground)",
          }}
          onMouseEnter={(e) =>
            !e.currentTarget.disabled &&
            (e.currentTarget.style.borderColor = "var(--primary)")
          }
          onMouseLeave={(e) =>
            (e.currentTarget.style.borderColor = "var(--border)")
          }
        >
          Next
        </button>
      </div>

      {/* デスクトップ表示 */}
      <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
        <div>
          <p className="text-sm" style={{ color: "var(--foreground)" }}>
            Showing page{" "}
            <span className="font-semibold">{pagination.page}</span> of{" "}
            <span className="font-semibold">{pagination.total_pages}</span> (
            <span className="font-semibold">{pagination.total}</span> total)
          </p>
        </div>
        <div>
          <nav className="relative z-0 inline-flex rounded-lg shadow-sm gap-2">
            <button
              onClick={goToPrevious}
              disabled={pagination.page === 1}
              className="relative inline-flex items-center px-4 py-2 rounded-lg border-2 text-sm font-medium transition-all disabled:opacity-40"
              style={{
                background: "var(--surface)",
                borderColor: "var(--border)",
                color: "var(--foreground)",
              }}
              onMouseEnter={(e) =>
                !e.currentTarget.disabled &&
                (e.currentTarget.style.borderColor = "var(--primary)")
              }
              onMouseLeave={(e) =>
                (e.currentTarget.style.borderColor = "var(--border)")
              }
            >
              Previous
            </button>
            <button
              onClick={goToNext}
              disabled={pagination.page === pagination.total_pages}
              className="relative inline-flex items-center px-4 py-2 rounded-lg border-2 text-sm font-medium transition-all disabled:opacity-40"
              style={{
                background: "var(--surface)",
                borderColor: "var(--border)",
                color: "var(--foreground)",
              }}
              onMouseEnter={(e) =>
                !e.currentTarget.disabled &&
                (e.currentTarget.style.borderColor = "var(--primary)")
              }
              onMouseLeave={(e) =>
                (e.currentTarget.style.borderColor = "var(--border)")
              }
            >
              Next
            </button>
          </nav>
        </div>
      </div>
    </div>
  );
}
