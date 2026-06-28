// 共享按钮：统一 demo 的尺寸、语义变体、加载和禁用行为。
import { forwardRef, type ButtonHTMLAttributes, type ReactNode } from "react";

export type ButtonVariant = "primary" | "ghost" | "text" | "danger" | "danger-solid" | "icon";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
  loading?: boolean;
  icon?: ReactNode;
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { variant = "ghost", loading = false, disabled, icon, children, className = "", ...props },
  ref,
) {
  const classes = ["btn", `btn--${variant}`, className].filter(Boolean).join(" ");

  return (
    <button
      {...props}
      ref={ref}
      className={classes}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
    >
      {loading ? <span className="btn__spinner" aria-hidden="true" /> : icon}
      <span className="btn__label">{children}</span>
    </button>
  );
});
