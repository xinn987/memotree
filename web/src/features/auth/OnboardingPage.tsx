// 已注册但尚未进入家庭的引导页：只保留创建家庭和使用邀请两个明确动作。
import { useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { Button } from "../../components/ui/Button";
import { Field } from "../../components/ui/Field";
import { InlineError } from "../../components/ui/Feedback";
import { createFamily, joinFamily } from "./auth.api";
import { useSession } from "../../app/SessionProvider";

export function OnboardingPage() {
  const { refresh } = useSession();
  const [familyName, setFamilyName] = useState("");
  const [inviteToken, setInviteToken] = useState("");
  const [busyAction, setBusyAction] = useState<"create" | "join" | null>(null);
  const [error, setError] = useState("");

  async function handleCreate(event: FormEvent) {
    event.preventDefault();
    const name = familyName.trim();
    if (!name) {
      setError("先给这个家起个名字");
      return;
    }
    setBusyAction("create");
    setError("");
    try {
      await createFamily(name);
      await refresh();
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "家庭还没建好，请再试一次");
    } finally {
      setBusyAction(null);
    }
  }

  async function handleJoin(event: FormEvent) {
    event.preventDefault();
    const token = inviteToken.trim();
    if (!token) {
      setError("请粘贴家人发来的邀请");
      return;
    }
    setBusyAction("join");
    setError("");
    try {
      await joinFamily(token);
      await refresh();
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "这条邀请暂时用不了");
    } finally {
      setBusyAction(null);
    }
  }

  return (
    <main className="onboarding">
      <div className="onboarding__brand">
        Memo<em>Tree</em>
      </div>
      <section className="onboarding__intro">
        <span className="eyebrow-brand">第一次来</span>
        <h1>先建一个家</h1>
        <p>以后家里人打开这里，就能看到你认真挑选的照片。</p>
      </section>
      {error && <InlineError>{error}</InlineError>}
      <div className="onboarding__actions">
        <form onSubmit={handleCreate}>
          <h2>创建新的家庭空间</h2>
          <Field label="家庭名称" hint="比如：小满的家">
            <input value={familyName} onChange={(event) => setFamilyName(event.target.value)} />
          </Field>
          <Button type="submit" variant="primary" loading={busyAction === "create"}>
            创建家庭
          </Button>
        </form>
        <form onSubmit={handleJoin}>
          <h2>已经收到邀请</h2>
          <Field label="邀请内容" hint="粘贴链接最后的一段字符">
            <input value={inviteToken} onChange={(event) => setInviteToken(event.target.value)} />
          </Field>
          <Button type="submit" loading={busyAction === "join"}>
            加入家人
          </Button>
          <Link to={`/join?invite=${encodeURIComponent(inviteToken)}`}>打开完整加入页面</Link>
        </form>
      </div>
    </main>
  );
}
