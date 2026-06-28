// 登录与邀请加入共用双栏结构；移动端仍保留情绪图片 banner。
import type { ReactNode } from "react";

type AuthShellProps = {
  imageUrl: string;
  imageAlt?: string;
  aside: ReactNode;
  children: ReactNode;
};

export function AuthShell({ imageUrl, imageAlt = "", aside, children }: AuthShellProps) {
  return (
    <div className="auth-shell">
      <aside className="auth-shell__aside">
        <img src={imageUrl} alt={imageAlt} />
        <div className="auth-shell__overlay" />
        <div className="auth-shell__brand">{aside}</div>
      </aside>
      <main className="auth-shell__main">
        <div className="auth-shell__card">{children}</div>
      </main>
    </div>
  );
}
