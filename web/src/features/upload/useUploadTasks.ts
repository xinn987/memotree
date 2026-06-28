// 上传编排 hook：把浏览器直传、Worker 轮询和失败恢复从页面结构中隔离。
import { useCallback, useEffect, useState } from "react";
import type { UploadBatch, UploadItem, UploadTask } from "../../api/contracts";
import {
  completeUpload,
  createUploadIntent,
  failUpload,
  getActiveUploadTask,
  isSupportedImageFile,
  listRecentUploadTasks,
  putOriginalFile,
  retryProcessing,
  retryUpload,
  stopUploadTask,
} from "./upload.api";

const uploadPollIntervalMs = 2000;

export function useUploadTasks(familyId: number) {
  const [task, setTask] = useState<UploadTask>({ batch: null, items: [] });
  const [recentTasks, setRecentTasks] = useState<UploadTask[]>([]);
  const [progress, setProgress] = useState<Record<number, number>>({});
  const [localFiles, setLocalFiles] = useState<Record<number, File>>({});
  const [uploadingIds, setUploadingIds] = useState<Set<number>>(() => new Set());
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const loadTasks = useCallback(async () => {
    setLoading(true);
    try {
      const [active, recent] = await Promise.all([
        getActiveUploadTask(familyId),
        listRecentUploadTasks(familyId),
      ]);
      setTask(normalizeTask(active));
      setRecentTasks((recent.tasks ?? []).map(normalizeTask));
      setError("");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "上传记录暂时没有打开");
    } finally {
      setLoading(false);
    }
  }, [familyId]);

  useEffect(() => {
    void loadTasks();
  }, [loadTasks]);

  useEffect(() => {
    if (!shouldPoll(task)) {
      return;
    }
    const timer = window.setInterval(() => void loadTasks(), uploadPollIntervalMs);
    return () => window.clearInterval(timer);
  }, [loadTasks, task]);

  useEffect(() => {
    if (uploadingIds.size === 0) {
      return;
    }
    const warnBeforeLeave = (event: BeforeUnloadEvent) => {
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", warnBeforeLeave);
    return () => window.removeEventListener("beforeunload", warnBeforeLeave);
  }, [uploadingIds.size]);

  const uploadOne = useCallback(
    async (batchId: number, item: UploadItem, file: File) => {
      setUploadingIds((current) => new Set(current).add(item.id));
      setProgress((current) => ({ ...current, [item.id]: 0 }));
      try {
        await putOriginalFile(item, file, (value) => setProgress((current) => ({ ...current, [item.id]: value })));
        const mutation = await completeUpload(familyId, batchId, item.id);
        updateItem(mutation.batch, mutation.item);
        setProgress((current) => ({ ...current, [item.id]: 100 }));
      } catch (requestError) {
        const message = requestError instanceof Error ? requestError.message : "这张照片没传上去";
        try {
          const mutation = await failUpload(familyId, batchId, item.id, message);
          updateItem(mutation.batch, mutation.item);
        } catch {
          setError(message);
        }
      } finally {
        setUploadingIds((current) => {
          const next = new Set(current);
          next.delete(item.id);
          return next;
        });
      }
    },
    [familyId],
  );

  async function selectFiles(files: File[]) {
    if (files.length === 0) {
      return;
    }
    const unsupported = files.find((file) => !isSupportedImageFile(file));
    if (unsupported) {
      setError(
        unsupported.type.startsWith("video/")
          ? `当前版本暂不支持上传视频：${unsupported.name}`
          : `当前版本只支持 JPG、PNG 和 GIF：${unsupported.name}`,
      );
      return;
    }

    setBusy(true);
    setError("");
    try {
      const intent = await createUploadIntent(familyId, files);
      const nextTask = normalizeTask(intent);
      setTask(nextTask);
      setRecentTasks((current) => upsertTask(current, nextTask));
      if (intent.activeExisting) {
        setError("家里已经有一批正在传，请先把这一批处理完");
        return;
      }
      if (!nextTask.batch) {
        setError("这批照片还没有准备好，请重新选择");
        return;
      }
      const fileMap = Object.fromEntries(nextTask.items.map((item, index) => [item.id, files[index]]).filter(([, file]) => file));
      setLocalFiles((current) => ({ ...current, ...fileMap }));
      for (const item of nextTask.items) {
        const file = fileMap[item.id] as File | undefined;
        if (file && item.uploadUrl) {
          await uploadOne(nextTask.batch.id, item, file);
        }
      }
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "这批照片还没有开始上传");
    } finally {
      setBusy(false);
    }
  }

  async function retryFailedUpload(item: UploadItem) {
    const batchId = task.batch?.id ?? item.uploadBatchId;
    const file = localFiles[item.id];
    if (!batchId || !file) {
      setError("当前浏览器没有这张原图，请重新选择后上传");
      return;
    }
    setBusy(true);
    setError("");
    try {
      const mutation = await retryUpload(familyId, batchId, item.id);
      updateItem(mutation.batch, mutation.item);
      await uploadOne(batchId, mutation.item, file);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "重新上传没有开始");
    } finally {
      setBusy(false);
    }
  }

  async function retryFailedProcessing(item: UploadItem) {
    const batchId = task.batch?.id ?? item.uploadBatchId;
    if (!batchId) {
      return;
    }
    setBusy(true);
    setError("");
    try {
      const mutation = await retryProcessing(familyId, batchId, item.id);
      updateItem(mutation.batch, mutation.item);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "重新整理没有开始");
    } finally {
      setBusy(false);
    }
  }

  async function stop() {
    if (!task.batch) {
      return;
    }
    setBusy(true);
    try {
      const stopped = normalizeTask(await stopUploadTask(familyId, task.batch.id));
      setTask(stopped);
      setRecentTasks((current) => upsertTask(current, stopped));
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "这批照片暂时没有停下");
    } finally {
      setBusy(false);
    }
  }

  function updateItem(batch: UploadBatch, item: UploadItem) {
    setTask((current) => upsertItem(current, batch, item));
    setRecentTasks((current) => current.map((entry) => upsertItem(entry, batch, item)));
  }

  return {
    task,
    recentTasks,
    progress,
    loading,
    busy,
    error,
    uploadingIds,
    selectFiles,
    retryFailedUpload,
    retryFailedProcessing,
    stop,
    refresh: loadTasks,
  };
}

function normalizeTask(task: UploadTask | null | undefined): UploadTask {
  return { batch: task?.batch ?? null, items: task?.items ?? [] };
}

function shouldPoll(task: UploadTask): boolean {
  return task.items.some((item) => ["waiting", "uploading", "uploaded", "processing"].includes(item.status));
}

function upsertItem(task: UploadTask, batch: UploadBatch, item: UploadItem): UploadTask {
  if (task.batch?.id !== batch.id) {
    return task;
  }
  return {
    batch,
    items: task.items.map((current) => (current.id === item.id ? item : current)),
  };
}

function upsertTask(tasks: UploadTask[], next: UploadTask): UploadTask[] {
  if (!next.batch) {
    return tasks;
  }
  return [next, ...tasks.filter((task) => task.batch?.id !== next.batch?.id)];
}
