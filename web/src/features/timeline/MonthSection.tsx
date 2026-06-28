// 月份区块把后端按日期分组的数据重新组织成 demo 的月度故事流。
import type { TimelineMedia } from "../../api/contracts";
import { PhotoTile } from "./PhotoTile";

type MonthSectionProps = {
  familyId: number;
  month: string;
  items: TimelineMedia[];
  allMediaIds: number[];
};

export function MonthSection({ familyId, month, items, allMediaIds }: MonthSectionProps) {
  const photoCount = items.filter((item) => item.mediaType !== "video").length;
  const videoCount = items.filter((item) => item.mediaType === "video").length;

  return (
    <section className="month">
      <div className="month__head">
        <h2 className="month__title">{formatMonthTitle(month)}</h2>
        <span className="month__count">
          {photoCount > 0 && `${photoCount} 张照片`}
          {photoCount > 0 && videoCount > 0 && " · "}
          {videoCount > 0 && `${videoCount} 段视频`}
        </span>
      </div>
      <p className="month__note">这个月的家庭手记稍后补上</p>
      <div className="story-grid">
        {items.map((media) => (
          <PhotoTile key={media.id} familyId={familyId} media={media} mediaIds={allMediaIds} />
        ))}
      </div>
    </section>
  );
}

function formatMonthTitle(month: string): string {
  const [year, monthNumber] = month.split("-").map(Number);
  if (!year || !monthNumber) {
    return month;
  }
  return new Intl.DateTimeFormat("zh-CN", { month: "long" }).format(new Date(year, monthNumber - 1, 1));
}
