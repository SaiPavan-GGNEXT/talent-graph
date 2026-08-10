import { ReactNode } from "react";
import { ApiError } from "../lib/api";

// StateGate renders the loading / error / offline / empty states so every
// page gets consistent treatment for free.
export function StateGate({
  loading,
  error,
  empty,
  emptyTitle = "Nothing here yet",
  emptyHint,
  onRetry,
  children,
}: {
  loading: boolean;
  error: unknown;
  empty?: boolean;
  emptyTitle?: string;
  emptyHint?: string;
  onRetry?: () => void;
  children: ReactNode;
}) {
  if (loading) {
    return (
      <div className="state" role="status" aria-live="polite">
        <div className="spinner" aria-hidden="true" />
        Loading…
      </div>
    );
  }
  if (error) {
    const offline = error instanceof ApiError && error.offline;
    return (
      <div className="state">
        <div className={offline ? "offline-icon" : "error-icon"} aria-hidden="true">
          {offline ? "🔌" : "⚠️"}
        </div>
        <div className="title">
          {offline ? "The graph database is unreachable" : "Something went wrong"}
        </div>
        <p>
          {error instanceof Error ? error.message : "Unexpected error."}
        </p>
        {onRetry && (
          <button className="btn ghost" onClick={onRetry}>
            Try again
          </button>
        )}
      </div>
    );
  }
  if (empty) {
    return (
      <div className="state">
        <div className="title">{emptyTitle}</div>
        {emptyHint && <p>{emptyHint}</p>}
      </div>
    );
  }
  return <>{children}</>;
}
