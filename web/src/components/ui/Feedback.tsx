// 页面级轻反馈：真实成功、错误和暂未开放都从同一可访问出口呈现。
import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import { capabilityMessage, type FrontendCapability } from "../../app/capabilities";
import { Button, type ButtonVariant } from "./Button";

type FeedbackTone = "info" | "success" | "error";

type FeedbackState = {
  message: string;
  tone: FeedbackTone;
};

type FeedbackContextValue = {
  notify: (message: string, tone?: FeedbackTone) => void;
  clear: () => void;
};

const FeedbackContext = createContext<FeedbackContextValue | null>(null);

export function FeedbackProvider({ children }: { children: ReactNode }) {
  const [feedback, setFeedback] = useState<FeedbackState | null>(null);
  const notify = useCallback((message: string, tone: FeedbackTone = "info") => setFeedback({ message, tone }), []);
  const clear = useCallback(() => setFeedback(null), []);
  const value = useMemo(() => ({ notify, clear }), [clear, notify]);

  return (
    <FeedbackContext.Provider value={value}>
      {children}
      {feedback && (
        <div className={`toast toast--${feedback.tone}`} role="status" onClick={clear}>
          {feedback.message}
        </div>
      )}
    </FeedbackContext.Provider>
  );
}

export function useFeedback(): FeedbackContextValue {
  const context = useContext(FeedbackContext);
  if (!context) {
    throw new Error("useFeedback 必须在 FeedbackProvider 内使用");
  }
  return context;
}

type PlaceholderButtonProps = {
  capability: FrontendCapability;
  children: ReactNode;
  variant?: ButtonVariant;
  className?: string;
};

export function PlaceholderButton({
  capability,
  children,
  variant = "ghost",
  className,
}: PlaceholderButtonProps) {
  const { notify } = useFeedback();

  return (
    <Button
      type="button"
      variant={variant}
      className={className}
      onClick={() => notify(capabilityMessage(capability))}
    >
      {children}
    </Button>
  );
}

export function InlineError({ children }: { children: ReactNode }) {
  return (
    <p className="inline-error" role="alert">
      {children}
    </p>
  );
}
