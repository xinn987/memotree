// 家庭首页：以照片和月份为主角，上传与管理入口保持可见但次要。
import { Filter, Upload } from "lucide-react";
import { Link } from "react-router-dom";
import type { TimelineMedia, TimelineMediaTypeFilter } from "../../api/contracts";
import { useFamilyRoute } from "../../app/ProtectedFamilyRoute";
import { Button } from "../../components/ui/Button";
import { InlineError } from "../../components/ui/Feedback";
import { EmptyState, Skeleton } from "../../components/ui/StatusViews";
import { MonthSection } from "./MonthSection";
import { useTimeline } from "./useTimeline";

const mediaFilters: Array<{ value: TimelineMediaTypeFilter; label: string }> = [
  { value: "", label: "全部" },
  { value: "photo", label: "照片" },
  { value: "video", label: "视频" },
  { value: "live_photo", label: "实况" },
];

export function TimelinePage() {
  const { family } = useFamilyRoute();
  const timeline = useTimeline(family.id);
  const months = groupByMonth(timeline.groups.flatMap((group) => group.items.map((item) => ({ ...item, month: group.month }))));
  const allMediaIds = months.flatMap((group) => group.items.map((item) => item.id));

  return (
    <>
      <main className="page-shell page-shell--gallery timeline-page">
        <section className="timeline-hero">
          <span className="eyebrow-brand">{new Date().getFullYear()} 年 · 家庭相册</span>
          <h1>
            {family.displayName.replace(/的家$/, "") || "家里"}最近<em>挺好的</em>
          </h1>
          <p>这是家里人最近拍下的瞬间。慢慢看。</p>
        </section>

        <div className="toolbar" role="toolbar" aria-label="时间线筛选">
          <div className="toolbar__group" role="tablist" aria-label="媒体类型">
            {mediaFilters.map((filter) => (
              <button
                key={filter.value || "all"}
                className={`seg${timeline.mediaType === filter.value ? " is-active" : ""}`}
                type="button"
                role="tab"
                aria-selected={timeline.mediaType === filter.value}
                onClick={() => timeline.setMediaType(filter.value)}
              >
                {filter.label}
              </button>
            ))}
          </div>
          <Button type="button" variant="text" loading={timeline.loading} onClick={() => void timeline.refresh()}>
            刷新
          </Button>
        </div>

        <div className="mobile-filter-bar" role="toolbar" aria-label="移动端时间线筛选">
          <div className="mobile-filter-bar__types" role="tablist" aria-label="媒体类型">
            {mediaFilters.map((filter) => (
              <button
                key={filter.value || "all"}
                className={`m-filter${timeline.mediaType === filter.value ? " is-active" : ""}`}
                type="button"
                role="tab"
                aria-selected={timeline.mediaType === filter.value}
                onClick={() => timeline.setMediaType(filter.value)}
              >
                {filter.label}
              </button>
            ))}
          </div>
          <button className="m-filter-icon" type="button" aria-label="月份筛选暂未开放">
            <Filter aria-hidden="true" size={17} />
          </button>
        </div>

        {timeline.error && (
          <div className="timeline-feedback">
            <InlineError>{timeline.error}</InlineError>
            <Button type="button" variant="text" onClick={() => void timeline.refresh()}>
              再试一次
            </Button>
          </div>
        )}

        {timeline.loading ? (
          <div className="timeline-skeletons">
            <Skeleton label="正在读取照片" lines={7} />
          </div>
        ) : months.length === 0 ? (
          <EmptyState
            title="这里还没有照片"
            action={
              <Link className="btn btn--primary" to={`/families/${family.id}/upload`}>
                选几张给家里看
              </Link>
            }
          >
            从相册里挑想让家人看到的那些，传完就会出现在这里。
          </EmptyState>
        ) : (
          months.map((group) => (
            <MonthSection
              key={group.month}
              familyId={family.id}
              month={group.month}
              items={group.items}
              allMediaIds={allMediaIds}
            />
          ))
        )}

        {timeline.nextCursor && (
          <div className="timeline-more">
            <Button type="button" loading={timeline.loadingMore} onClick={() => void timeline.loadMore()}>
              再看更早的
            </Button>
          </div>
        )}
      </main>

      <Link className="fab-upload" to={`/families/${family.id}/upload`}>
        <Upload aria-hidden="true" size={18} />
        <span>上传新的</span>
      </Link>
    </>
  );
}

type MonthGroup = {
  month: string;
  items: TimelineMedia[];
};

function groupByMonth(items: Array<TimelineMedia & { month: string }>): MonthGroup[] {
  const groups = new Map<string, TimelineMedia[]>();
  for (const item of items) {
    const group = groups.get(item.month) ?? [];
    group.push(item);
    groups.set(item.month, group);
  }
  return Array.from(groups, ([month, monthItems]) => ({ month, items: monthItems }));
}
