// 当前上传任务：状态、进度和恢复操作严格映射 upload demo。
import { Image, RotateCw } from "lucide-react";
import type { UploadItem, UploadTask } from "../../api/contracts";
import { Button } from "../../components/ui/Button";
import { Chip } from "../../components/ui/Chip";
import { ProgressBar } from "../../components/ui/StatusViews";
import { formatBytes, uploadBatchStatusText, uploadItemStatusText } from "../../utils/format";

type UploadTaskViewProps = {
  task: UploadTask;
  progress: Record<number, number>;
  busy: boolean;
  onRetryUpload: (item: UploadItem) => void;
  onRetryProcessing: (item: UploadItem) => void;
  onStop: () => void;
};

export function UploadTaskView({
  task,
  progress,
  busy,
  onRetryUpload,
  onRetryProcessing,
  onStop,
}: UploadTaskViewProps) {
  if (!task.batch) {
    return null;
  }

  return (
    <section className="upload-task">
      <header className="upload-task__head">
        <h2>
          这批正在传
          <Chip tone="accent">{uploadBatchStatusText(task.batch.status)}</Chip>
        </h2>
        <div className="upload-task__stats">
          <span>
            <strong>{task.batch.readyCount} / {task.batch.totalCount}</strong> 已传完
          </span>
          <span>
            <strong>{task.batch.failedCount}</strong> 没传成
          </span>
        </div>
      </header>

      <div className="upload-file-list">
        {task.items.map((item) => {
          const value = itemProgress(item, progress[item.id]);
          return (
            <article className="upload-file" key={item.id}>
              <div className="upload-file__thumb">
                <Image aria-hidden="true" size={20} />
              </div>
              <div className="upload-file__main">
                <div className="upload-file__name">
                  <strong>{item.originalFilename}</strong>
                  <span>{formatBytes(item.byteSize)}</span>
                </div>
                <ProgressBar value={value} label={`${item.originalFilename} 上传进度`} />
                <div className="upload-file__state">
                  <Chip tone={statusTone(item.status)}>{uploadItemStatusText(item.status, value)}</Chip>
                  {item.status === "upload_failed" && (
                    <Button type="button" variant="text" disabled={busy} onClick={() => onRetryUpload(item)}>
                      重新上传
                    </Button>
                  )}
                  {item.status === "processing_failed" && (
                    <Button
                      type="button"
                      variant="text"
                      icon={<RotateCw aria-hidden="true" size={14} />}
                      disabled={busy}
                      onClick={() => onRetryProcessing(item)}
                    >
                      重新整理
                    </Button>
                  )}
                </div>
                {item.errorMessage && <p className="upload-file__error">{item.errorMessage}</p>}
              </div>
            </article>
          );
        })}
      </div>

      <footer className="upload-task__foot">
        <span>还没传完的，关掉页面可能会中断。整理这步后台会继续。</span>
        <Button type="button" variant="ghost" disabled={busy} onClick={onStop}>
          先不传了
        </Button>
      </footer>
    </section>
  );
}

function itemProgress(item: UploadItem, current = 0): number {
  if (["uploaded", "processing", "ready", "processing_failed"].includes(item.status)) return 100;
  if (item.status === "upload_failed") return current;
  return item.status === "uploading" ? current : 0;
}

function statusTone(status: UploadItem["status"]): "default" | "ok" | "warn" | "error" {
  if (status === "ready") return "ok";
  if (status === "processing") return "warn";
  if (status === "upload_failed" || status === "processing_failed") return "error";
  return "default";
}
