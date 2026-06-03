import { Upload } from "lucide-react";

type TimelineItem = {
  id: string;
  kind: "photo" | "video";
  title: string;
  dateLabel: string;
};

const sampleTimeline: TimelineItem[] = [
  { id: "1", kind: "photo", title: "早晨晒太阳", dateLabel: "今天" },
  { id: "2", kind: "video", title: "第一次自己翻身", dateLabel: "今天" },
  { id: "3", kind: "photo", title: "和爷爷散步", dateLabel: "昨天" },
];

export function App() {
  return (
    <main className="shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">家庭相册</p>
          <h1>MemoTree</h1>
        </div>
        <button className="icon-button" type="button" aria-label="上传照片或视频">
          <Upload aria-hidden="true" size={22} />
        </button>
      </header>

      <section className="timeline" aria-label="最近照片和视频">
        <h2>最近</h2>
        <div className="media-grid">
          {sampleTimeline.map((item) => (
            <article className="media-tile" key={item.id}>
              {/* 后续这里替换为后端返回的缩略图或视频封面，避免加载原文件。 */}
              <div className="media-preview" data-kind={item.kind}>
                <span>{item.kind === "video" ? "视频" : "照片"}</span>
              </div>
              <div className="media-meta">
                <strong>{item.title}</strong>
                <span>{item.dateLabel}</span>
              </div>
            </article>
          ))}
        </div>
      </section>
    </main>
  );
}
