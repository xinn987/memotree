// 登录与注册页：布局严格对应 auth demo，业务只调用既有认证 API。
import { useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { AuthShell } from "../../components/layout/AuthShell";
import { Button } from "../../components/ui/Button";
import { Field } from "../../components/ui/Field";
import { InlineError } from "../../components/ui/Feedback";
import { useSession } from "../../app/SessionProvider";
import { login, register } from "./auth.api";

type AuthMode = "login" | "register";

export function AuthPage() {
  const { refresh } = useSession();
  const [mode, setMode] = useState<AuthMode>("login");
  const [displayName, setDisplayName] = useState("");
  const [loginName, setLoginName] = useState("");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    if (!loginName.trim() || password.length < 6 || (mode === "register" && !displayName.trim())) {
      setError(mode === "login" ? "请填写登录名和至少 6 位密码" : "请把称呼、登录名和至少 6 位密码填完整");
      return;
    }

    setBusy(true);
    setError("");
    try {
      if (mode === "login") {
        await login({ loginName: loginName.trim(), password });
      } else {
        await register({
          displayName: displayName.trim(),
          loginName: loginName.trim(),
          password,
        });
      }
      await refresh();
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "这次没有登录成功，请再试一次");
    } finally {
      setBusy(false);
    }
  }

  return (
    <AuthShell
      imageUrl="https://picsum.photos/seed/family-warm/1200/1600"
      aside={
        <>
          <h1>
            Memo<em>Tree</em>
          </h1>
          <p>家人每天打开看看宝宝近况。</p>
        </>
      }
    >
      <section className="auth-page">
        <header className="auth-page__heading">
          <span className="eyebrow-brand">{mode === "login" ? "欢迎回来" : "第一次来"}</span>
          <h1>{mode === "login" ? "回到家里的相册" : "创建你的账号"}</h1>
        </header>

        <form className="auth-page__form" onSubmit={handleSubmit}>
          {mode === "register" && (
            <Field label="家人怎么称呼你" hint="比如：妈妈、爸爸、外婆">
              <input
                value={displayName}
                onChange={(event) => setDisplayName(event.target.value)}
                autoComplete="name"
              />
            </Field>
          )}
          <Field label="登录名" hint="手机号或用户名">
            <input
              value={loginName}
              onChange={(event) => setLoginName(event.target.value)}
              autoComplete="username"
            />
          </Field>
          <Field label="密码" hint={mode === "register" ? "至少 6 位" : undefined}>
            <input
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              type="password"
              autoComplete={mode === "login" ? "current-password" : "new-password"}
            />
          </Field>
          {error && <InlineError>{error}</InlineError>}
          <Button type="submit" variant="primary" loading={busy} className="auth-page__submit">
            {mode === "login" ? "登录" : "注册并继续"}
          </Button>
        </form>

        <p className="auth-page__switch">
          {mode === "login" ? "还没账号？" : "已经有账号？"}
          <button type="button" onClick={() => setMode(mode === "login" ? "register" : "login")}>
            {mode === "login" ? "创建账号" : "回到登录"}
          </button>
        </p>

        <div className="auth-page__divider">或者</div>
        <div className="auth-page__invite-note">
          <strong>收到家人的邀请？</strong>
          <span>打开邀请链接后，填一下自己的信息就能加入。</span>
          <Link to="/join">使用邀请加入</Link>
        </div>
      </section>
    </AuthShell>
  );
}
