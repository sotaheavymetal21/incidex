import { useState, useMemo, useCallback } from "react";

interface UsePaginationOptions {
  totalItems: number;
  pageSize: number;
  initialPage?: number;
}

interface UsePaginationReturn {
  currentPage: number;
  totalPages: number;
  pageSize: number;
  startIndex: number;
  endIndex: number;
  canGoNext: boolean;
  canGoPrev: boolean;
  goToPage: (page: number) => void;
  nextPage: () => void;
  prevPage: () => void;
  firstPage: () => void;
  lastPage: () => void;
  setPageSize: (size: number) => void;
  paginatedItems: <T>(items: T[]) => T[];
  pageNumbers: number[];
}

/**
 * ページネーション管理のカスタムフック
 *
 * @example
 * const {
 *   currentPage,
 *   totalPages,
 *   paginatedItems,
 *   nextPage,
 *   prevPage,
 *   goToPage,
 * } = usePagination({ totalItems: incidents.length, pageSize: 10 });
 *
 * const displayedIncidents = paginatedItems(incidents);
 */
export function usePagination({
  totalItems,
  pageSize: initialPageSize,
  initialPage = 1,
}: UsePaginationOptions): UsePaginationReturn {
  const [currentPage, setCurrentPage] = useState(initialPage);
  const [pageSize, setPageSizeState] = useState(initialPageSize);

  const totalPages = useMemo(
    () => Math.max(1, Math.ceil(totalItems / pageSize)),
    [totalItems, pageSize],
  );

  // Ensure current page is within valid range
  const validatedPage = useMemo(() => {
    if (currentPage < 1) return 1;
    if (currentPage > totalPages) return totalPages;
    return currentPage;
  }, [currentPage, totalPages]);

  // Update page if it's out of range
  if (validatedPage !== currentPage) {
    setCurrentPage(validatedPage);
  }

  const startIndex = useMemo(
    () => (validatedPage - 1) * pageSize,
    [validatedPage, pageSize],
  );

  const endIndex = useMemo(
    () => Math.min(startIndex + pageSize, totalItems),
    [startIndex, pageSize, totalItems],
  );

  const canGoNext = validatedPage < totalPages;
  const canGoPrev = validatedPage > 1;

  const goToPage = useCallback(
    (page: number) => {
      const validPage = Math.max(1, Math.min(page, totalPages));
      setCurrentPage(validPage);
    },
    [totalPages],
  );

  const nextPage = useCallback(() => {
    if (canGoNext) {
      setCurrentPage((prev) => prev + 1);
    }
  }, [canGoNext]);

  const prevPage = useCallback(() => {
    if (canGoPrev) {
      setCurrentPage((prev) => prev - 1);
    }
  }, [canGoPrev]);

  const firstPage = useCallback(() => {
    setCurrentPage(1);
  }, []);

  const lastPage = useCallback(() => {
    setCurrentPage(totalPages);
  }, [totalPages]);

  const setPageSize = useCallback((size: number) => {
    setPageSizeState(size);
    setCurrentPage(1); // Reset to first page when page size changes
  }, []);

  const paginatedItems = useCallback(
    <T>(items: T[]): T[] => {
      return items.slice(startIndex, endIndex);
    },
    [startIndex, endIndex],
  );

  // Generate page numbers for pagination UI
  const pageNumbers = useMemo(() => {
    const pages: number[] = [];
    const maxVisiblePages = 5;

    if (totalPages <= maxVisiblePages) {
      // Show all pages
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      // Show pages around current page
      let startPage = Math.max(1, validatedPage - 2);
      let endPage = Math.min(totalPages, validatedPage + 2);

      // Adjust if at the beginning
      if (validatedPage <= 3) {
        endPage = maxVisiblePages;
      }

      // Adjust if at the end
      if (validatedPage >= totalPages - 2) {
        startPage = totalPages - maxVisiblePages + 1;
      }

      for (let i = startPage; i <= endPage; i++) {
        pages.push(i);
      }
    }

    return pages;
  }, [totalPages, validatedPage]);

  return {
    currentPage: validatedPage,
    totalPages,
    pageSize,
    startIndex,
    endIndex,
    canGoNext,
    canGoPrev,
    goToPage,
    nextPage,
    prevPage,
    firstPage,
    lastPage,
    setPageSize,
    paginatedItems,
    pageNumbers,
  };
}

/**
 * サーバーサイドページネーション用のフック
 * ページ変更時にコールバックを呼び出す
 *
 * @example
 * const pagination = useServerPagination({
 *   totalItems: data.total,
 *   pageSize: 20,
 *   onPageChange: (page) => fetchData({ page }),
 * });
 */
interface UseServerPaginationOptions extends UsePaginationOptions {
  onPageChange?: (page: number) => void;
  onPageSizeChange?: (pageSize: number) => void;
}

export function useServerPagination({
  totalItems,
  pageSize: initialPageSize,
  initialPage = 1,
  onPageChange,
  onPageSizeChange,
}: UseServerPaginationOptions) {
  const pagination = usePagination({
    totalItems,
    pageSize: initialPageSize,
    initialPage,
  });

  const goToPage = useCallback(
    (page: number) => {
      pagination.goToPage(page);
      onPageChange?.(page);
    },
    [pagination, onPageChange],
  );

  const nextPage = useCallback(() => {
    if (pagination.canGoNext) {
      const newPage = pagination.currentPage + 1;
      pagination.nextPage();
      onPageChange?.(newPage);
    }
  }, [pagination, onPageChange]);

  const prevPage = useCallback(() => {
    if (pagination.canGoPrev) {
      const newPage = pagination.currentPage - 1;
      pagination.prevPage();
      onPageChange?.(newPage);
    }
  }, [pagination, onPageChange]);

  const setPageSize = useCallback(
    (size: number) => {
      pagination.setPageSize(size);
      onPageSizeChange?.(size);
      onPageChange?.(1); // Reset to first page
    },
    [pagination, onPageSizeChange, onPageChange],
  );

  return {
    ...pagination,
    goToPage,
    nextPage,
    prevPage,
    setPageSize,
  };
}
