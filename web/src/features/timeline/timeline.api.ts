// 时间线 API 适配器：查询参数严格对应现有 Go API。
import type { TimelineMediaTypeFilter, TimelineResponse } from "../../api/contracts";
import { requestJSON } from "../../api/client";

export type TimelineQuery = {
  cursor?: string;
  mediaType?: TimelineMediaTypeFilter;
  month?: string;
  limit?: number;
};

export function listTimeline(familyId: number, query: TimelineQuery = {}, signal?: AbortSignal) {
  const params = new URLSearchParams({ limit: String(query.limit ?? 30) });
  if (query.cursor) {
    params.set("cursor", query.cursor);
  }
  if (query.mediaType) {
    params.set("mediaType", query.mediaType);
  }
  if (query.month) {
    params.set("month", query.month);
  }
  return requestJSON<TimelineResponse>(`/families/${familyId}/timeline?${params.toString()}`, { signal });
}
