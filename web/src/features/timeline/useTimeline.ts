// 时间线状态 hook：负责筛选、分页、请求取消和同日期分组合并。
import { useCallback, useEffect, useState } from "react";
import type { TimelineGroup, TimelineMediaTypeFilter } from "../../api/contracts";
import { listTimeline } from "./timeline.api";

export function useTimeline(familyId: number) {
  const [groups, setGroups] = useState<TimelineGroup[]>([]);
  const [mediaType, setMediaType] = useState<TimelineMediaTypeFilter>("");
  const [month, setMonth] = useState("");
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState("");

  const loadInitial = useCallback(
    async (signal?: AbortSignal) => {
      setLoading(true);
      setError("");
      try {
        const response = await listTimeline(familyId, { mediaType, month, limit: 30 }, signal);
        setGroups(response.groups ?? []);
        setNextCursor(response.nextCursor ?? null);
      } catch (requestError) {
        if (requestError instanceof DOMException && requestError.name === "AbortError") {
          return;
        }
        setError(requestError instanceof Error ? requestError.message : "时间线暂时没有打开");
      } finally {
        if (!signal?.aborted) {
          setLoading(false);
        }
      }
    },
    [familyId, mediaType, month],
  );

  useEffect(() => {
    const controller = new AbortController();
    void loadInitial(controller.signal);
    return () => controller.abort();
  }, [loadInitial]);

  const loadMore = useCallback(async () => {
    if (!nextCursor || loadingMore) {
      return;
    }
    setLoadingMore(true);
    setError("");
    try {
      const response = await listTimeline(familyId, {
        cursor: nextCursor,
        mediaType,
        month,
        limit: 30,
      });
      setGroups((current) => mergeTimelineGroups(current, response.groups ?? []));
      setNextCursor(response.nextCursor ?? null);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "更早的照片暂时没有打开");
    } finally {
      setLoadingMore(false);
    }
  }, [familyId, loadingMore, mediaType, month, nextCursor]);

  return {
    groups,
    mediaType,
    month,
    nextCursor,
    loading,
    loadingMore,
    error,
    setMediaType,
    setMonth,
    refresh: () => loadInitial(),
    loadMore,
  };
}

// 同一天可能跨分页出现；合并时去重，保证月度瀑布流稳定。
export function mergeTimelineGroups(current: TimelineGroup[], incoming: TimelineGroup[]): TimelineGroup[] {
  const merged = current.map((group) => ({ ...group, items: [...group.items] }));
  for (const nextGroup of incoming) {
    const existing = merged.find((group) => group.date === nextGroup.date);
    if (!existing) {
      merged.push({ ...nextGroup, items: [...nextGroup.items] });
      continue;
    }
    const knownIds = new Set(existing.items.map((item) => item.id));
    existing.items.push(...nextGroup.items.filter((item) => !knownIds.has(item.id)));
  }
  return merged;
}
