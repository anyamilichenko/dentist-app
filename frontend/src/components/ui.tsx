import type { ReactNode } from "react";

export const cls = (...names: (string | false | null | undefined)[]) => names.filter(Boolean).join(" ");

export function Field({ label, children, className }: { label: string; children: ReactNode; className?: string }) {
  return (
    <label className={cls("field", className)}>
      <span className="field-label">{label}</span>
      {children}
    </label>
  );
}

export function ToothIcon() {
  return (
    <svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true">
      <path
        d="M8.3 3.5c-2.2 0-4 1.7-4 4.2 0 1.9.6 3.2 1.2 4.5.6 1.3 1 3 1.2 5 .3 2.2.9 3.8 2 3.8 1.3 0 1.6-1.7 2-3.5.3-1.3.8-2.3 2.3-2.3s2 1 2.3 2.3c.4 1.8.7 3.5 2 3.5 1.1 0 1.7-1.6 2-3.8.2-2 .6-3.7 1.2-5 .6-1.3 1.2-2.6 1.2-4.5 0-2.5-1.8-4.2-4-4.2-1.8 0-2.8.9-4.7.9s-2.9-.9-4.7-.9z"
        fill="currentColor"
      />
    </svg>
  );
}
