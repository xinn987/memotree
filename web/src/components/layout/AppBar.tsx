// 家庭页面顶栏：品牌、家庭上下文和高频导航保持克制，不与照片争夺注意力。
import { Home, LogOut, Settings, Upload } from "lucide-react";
import { Link, useNavigate } from "react-router-dom";
import type { Family, User } from "../../api/contracts";
import { Button } from "../ui/Button";

type AppBarProps = {
  family: Family;
  families: Family[];
  user: User;
  onLogout: () => void;
};

export function AppBar({ family, families, user, onLogout }: AppBarProps) {
  const navigate = useNavigate();
  const basePath = `/families/${family.id}`;
  const avatarText = (family.memberDisplayName || user.displayName || "家").slice(0, 1);

  return (
    <header className="app-bar">
      <Link className="app-bar__brand" to={`${basePath}/timeline`}>
        Memo<em>Tree</em>
      </Link>
      {families.length > 1 ? (
        <label className="app-bar__family app-bar__family-picker">
          <span className="sr-only">切换家庭</span>
          <select
            aria-label="切换家庭"
            value={family.id}
            onChange={(event) => navigate(`/families/${event.target.value}/timeline`)}
          >
            {families.map((item) => (
              <option key={item.id} value={item.id}>
                {item.displayName}
              </option>
            ))}
          </select>
        </label>
      ) : (
        <div className="app-bar__family">
          <span>{family.displayName}</span>
          <span className="faint">· {family.role === "admin" ? "管理员" : "家人"}</span>
        </div>
      )}
      <nav className="app-bar__actions" aria-label="家庭导航">
        <Link className="btn btn--icon btn--ghost app-bar__desktop-action" to={`${basePath}/timeline`} aria-label="时间线">
          <Home aria-hidden="true" size={18} />
        </Link>
        <Link className="btn btn--icon btn--ghost app-bar__desktop-action" to={`${basePath}/upload`} aria-label="上传">
          <Upload aria-hidden="true" size={18} />
        </Link>
        {family.role === "admin" && (
          <Link className="btn btn--icon btn--ghost app-bar__desktop-action" to={`${basePath}/invites`} aria-label="家人管理">
            <Settings aria-hidden="true" size={18} />
          </Link>
        )}
        <button className="avatar" type="button" aria-label="退出登录" title="退出登录" onClick={onLogout}>
          <span>{avatarText}</span>
          <LogOut className="avatar__logout" aria-hidden="true" size={14} />
        </button>
      </nav>
    </header>
  );
}
