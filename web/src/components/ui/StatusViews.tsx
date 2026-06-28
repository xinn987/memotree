// 加载、空状态和进度是正式页面状态，不用零散文字或无语义 div 代替。
import type { ReactNode } from "react";

export function ProgressBar({ value, label }: { value: number; label: string }) {
  const normalized = Math.max(0, Math.min(100, Math.round(value)));
  return (
    <div
      className="progress"
      role="progressbar"
      aria-label={label}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-valuenow={normalized}
    >
      <span className="progress__bar" style={{ width: `${normalized}%` }} />
    </div>
  );
}

export function Skeleton({ label, lines = 3 }: { label: string; lines?: number }) {
  return (
    <div className="skeleton" aria-label={label} aria-busy="true">
      {Array.from({ length: lines }, (_, index) => (
        <span key={index} style={{ width: `${100 - index * 12}%` }} />
      ))}
    </div>
  );
}

export function EmptyState({
  title,
  children,
  action,
}: {
  title: string;
  children: ReactNode;
  action?: ReactNode;
}) {
  return (
    <section className="empty-state">
      <div className="empty-state__mark" aria-hidden="true">
        M
      </div>
      <h2>{title}</h2>
      <p>{children}</p>
      {action}
    </section>
  );
}
