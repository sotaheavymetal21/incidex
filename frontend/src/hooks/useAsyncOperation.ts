import { useState, useCallback } from "react";

type AsyncState<T> =
  | { status: "idle"; data: null; error: null }
  | { status: "loading"; data: null; error: null }
  | { status: "success"; data: T; error: null }
  | { status: "error"; data: null; error: Error };

interface UseAsyncOperationReturn<T> {
  status: "idle" | "loading" | "success" | "error";
  data: T | null;
  error: Error | null;
  isIdle: boolean;
  isLoading: boolean;
  isSuccess: boolean;
  isError: boolean;
  execute: (operation: () => Promise<T>) => Promise<T>;
  reset: () => void;
}

/**
 * 非同期操作の状態管理を統一するカスタムフック
 *
 * @example
 * const { execute, isLoading, data, error } = useAsyncOperation<User>();
 *
 * const handleSubmit = async () => {
 *   try {
 *     const user = await execute(() => userApi.create(formData));
 *     console.log('Created:', user);
 *   } catch (err) {
 *     console.error('Failed:', err);
 *   }
 * };
 */
export function useAsyncOperation<T>(): UseAsyncOperationReturn<T> {
  const [state, setState] = useState<AsyncState<T>>({
    status: "idle",
    data: null,
    error: null,
  });

  const execute = useCallback(
    async (operation: () => Promise<T>): Promise<T> => {
      setState({ status: "loading", data: null, error: null });
      try {
        const data = await operation();
        setState({ status: "success", data, error: null });
        return data;
      } catch (err) {
        const error = err instanceof Error ? err : new Error(String(err));
        setState({ status: "error", data: null, error });
        throw error;
      }
    },
    [],
  );

  const reset = useCallback(() => {
    setState({ status: "idle", data: null, error: null });
  }, []);

  return {
    status: state.status,
    data: state.data,
    error: state.error,
    isIdle: state.status === "idle",
    isLoading: state.status === "loading",
    isSuccess: state.status === "success",
    isError: state.status === "error",
    execute,
    reset,
  };
}

/**
 * 複数の非同期操作を管理するフック
 * 各操作に名前をつけて個別に状態を追跡できる
 *
 * @example
 * const { getOperation, execute } = useMultipleAsyncOperations();
 *
 * await execute('createUser', () => userApi.create(data));
 * await execute('sendEmail', () => emailApi.send(email));
 *
 * if (getOperation('createUser').isLoading) {
 *   // Creating user...
 * }
 */
export function useMultipleAsyncOperations() {
  const [operations, setOperations] = useState<
    Record<string, AsyncState<unknown>>
  >({});

  const getOperation = useCallback(
    (name: string) => {
      const op = operations[name] || {
        status: "idle",
        data: null,
        error: null,
      };
      return {
        ...op,
        isIdle: op.status === "idle",
        isLoading: op.status === "loading",
        isSuccess: op.status === "success",
        isError: op.status === "error",
      };
    },
    [operations],
  );

  const execute = useCallback(
    async <T>(name: string, operation: () => Promise<T>): Promise<T> => {
      setOperations((prev) => ({
        ...prev,
        [name]: { status: "loading", data: null, error: null },
      }));

      try {
        const data = await operation();
        setOperations((prev) => ({
          ...prev,
          [name]: { status: "success", data, error: null },
        }));
        return data;
      } catch (err) {
        const error = err instanceof Error ? err : new Error(String(err));
        setOperations((prev) => ({
          ...prev,
          [name]: { status: "error", data: null, error },
        }));
        throw error;
      }
    },
    [],
  );

  const reset = useCallback((name?: string) => {
    if (name) {
      setOperations((prev) => ({
        ...prev,
        [name]: { status: "idle", data: null, error: null },
      }));
    } else {
      setOperations({});
    }
  }, []);

  return {
    operations,
    getOperation,
    execute,
    reset,
  };
}
