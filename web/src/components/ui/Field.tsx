// Field 通过原生 label/id 关系组织输入框，避免页面手工拼接可访问属性。
import { Children, cloneElement, isValidElement, useId, type ReactElement, type ReactNode } from "react";

type FieldProps = {
  label: string;
  hint?: string;
  error?: string;
  children: ReactElement<Record<string, unknown>>;
};

export function Field({ label, hint, error, children }: FieldProps) {
  const generatedId = useId();
  const child = Children.only(children);
  const inputId = String(child.props.id ?? generatedId);
  const descriptionId = hint || error ? `${inputId}-description` : undefined;

  if (!isValidElement(child)) {
    return null;
  }

  return (
    <div className={`field${error ? " field--error" : ""}`}>
      <label className="field__label" htmlFor={inputId}>
        {label}
      </label>
      {cloneElement(child, {
        id: inputId,
        className: ["input", child.props.className].filter(Boolean).join(" "),
        "aria-describedby": descriptionId,
        "aria-invalid": error ? true : undefined,
      })}
      {(error || hint) && (
        <span id={descriptionId} className={error ? "field__error" : "field__hint"}>
          {error || hint}
        </span>
      )}
    </div>
  );
}

export function FieldGroup({ children }: { children: ReactNode }) {
  return <div className="field-group">{children}</div>;
}
