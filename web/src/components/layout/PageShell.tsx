// 产品页面的宽度与纵向节奏容器。
import type { HTMLAttributes } from "react";

type PageShellProps = HTMLAttributes<HTMLElement> & {
  width?: "content" | "gallery";
};

export function PageShell({ width = "content", className = "", ...props }: PageShellProps) {
  return <main {...props} className={`page-shell page-shell--${width} ${className}`.trim()} />;
}
