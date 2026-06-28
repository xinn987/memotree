// 状态标签只承担简短状态，不作为可点击控件。
import type { HTMLAttributes } from "react";

type ChipTone = "default" | "accent" | "ok" | "warn" | "error" | "info";

type ChipProps = HTMLAttributes<HTMLSpanElement> & {
  tone?: ChipTone;
};

export function Chip({ tone = "default", className = "", ...props }: ChipProps) {
  const toneClass = tone === "default" ? "" : `chip--${tone === "error" ? "err" : tone}`;
  return <span {...props} className={["chip", toneClass, className].filter(Boolean).join(" ")} />;
}
