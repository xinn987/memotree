// 邀请加入页：在不新增后端接口的前提下，组合注册、加入和会话刷新。
import { useState, type FormEvent } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { AuthShell } from "../../components/layout/AuthShell";
import { Button } from "../../components/ui/Button";
import { Field } from "../../components/ui/Field";
import { InlineError } from "../../components/ui/Feedback";
import { useSession } from "../../app/SessionProvider";
import { joinFamily, register } from "./auth.api";

export function JoinPage() {
  const [searchParams] = useSearchParams();
  const inviteToken = searchParams.get("invite")?.trim() ?? "";
  const { user, refresh } = useSession();
  const [displayName, setDisplayName] = useState("");
  const [loginName, setLoginName] = useState("");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    if (!inviteToken) {
      setError("这条邀请里没有可用的信息，请让家人重新发一条");
      return;
    }
    if (!user && (!displayName.trim() || !loginName.trim() || password.length < 6)) {
      setError("请把称呼、登录名和至少 6 位密码填完整");
      return;
    }

    setBusy(true);
    setError("");
    try {
      if (!user) {
        await register({
          displayName: displayName.trim(),
          loginName: loginName.trim(),
          password,
        });
      }
      await joinFamily(inviteToken);
      await refresh();
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "这条邀请暂时用不了，请让家人重新发一条");
    } finally {
      setBusy(false);
    }
  }

  return (
    <AuthShell
      imageUrl="https://picsum.photos/seed/family-warm2/1200/1600"
      aside={
        <>
          <h1>
            Memo<em>Tree</em>
          </h1>
          <p>你被邀请加入家人的相册。</p>
          <div className="join-page__family-card">
            <div className="join-page__avatars" aria-hidden="true">
              <span>妈</span>
              <span>爸</span>
              <span>家</span>
            </div>
            <span>家人的相册</span>
          </div>
        </>
      }
    >
      <section className="join-page">
        <header className="auth-page__heading">
          <span className="eyebrow-brand">欢迎加入</span>
          <h1>填一下你的信息</h1>
          <p>{user ? "确认后就能加入这个家庭相册。" : "填完就能看到家里人的照片了。"}</p>
        </header>

        <form className="auth-page__form" onSubmit={handleSubmit}>
          {!user && (
            <>
              <Field label="家人怎么称呼你" hint="邀请人给你的称呼也可以之后再改">
                <input value={displayName} onChange={(event) => setDisplayName(event.target.value)} />
              </Field>
              <Field label="登录名" hint="手机号或用户名，用来以后登录">
                <input value={loginName} onChange={(event) => setLoginName(event.target.value)} autoComplete="username" />
              </Field>
              <Field label="设一个密码" hint="至少 6 位">
                <input
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  type="password"
                  autoComplete="new-password"
                />
              </Field>
            </>
          )}
          {error && <InlineError>{error}</InlineError>}
          <Button type="submit" variant="primary" loading={busy} className="auth-page__submit">
            加入家人的相册
          </Button>
        </form>

        <p className="join-page__hint">
          加入后你能看到时间线里的照片，也能上传新的。这里是私密家庭空间，不会生成公开分享链接。
        </p>
        {!user && (
          <p className="auth-page__switch">
            已经有账号？<Link to="/login">先登录</Link>
          </p>
        )}
      </section>
    </AuthShell>
  );
}
