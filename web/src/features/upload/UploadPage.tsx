// 上传页面：选择入口次于时间线，但任务反馈必须清楚可靠。
import { Upload } from "lucide-react";
import { useFamilyRoute } from "../../app/ProtectedFamilyRoute";
import { Button } from "../../components/ui/Button";
import { InlineError } from "../../components/ui/Feedback";
import { Skeleton } from "../../components/ui/StatusViews";
import { formatDateTime, uploadBatchStatusText } from "../../utils/format";
import { UploadTaskView } from "./UploadTaskView";
import { useUploadTasks } from "./useUploadTasks";

export function UploadPage() {
  const { family } = useFamilyRoute();
  const uploads = useUploadTasks(family.id);

  return (
    <main className="page-shell upload-page">
      <section className="upload-hero">
        <span className="eyebrow-brand">上传</span>
        <h1>给家里加几张新的</h1>
        <p>挑你想让家人看到的那些。传完会自动整理好，出现在时间线上。</p>
      </section>

      <label className="upload-drop">
        <input
          type="file"
          multiple
          accept="image/*,video/*"
          aria-label="选择照片或视频"
          onChange={(event) => void uploads.selectFiles(Array.from(event.target.files ?? []))}
        />
        <span className="upload-drop__icon">
          <Upload aria-hidden="true" size={26} />
        </span>
        <strong>选择照片或视频</strong>
        <span>可以一次选很多张。手机上也能直接从相册里挑。</span>
        <small>当前实际支持 JPG、PNG、GIF；视频入口先按 demo 保留。</small>
      </label>

      <label className="mobile-upload-action">
        <Upload aria-hidden="true" size={19} />
        从相册选照片
        <input
          type="file"
          multiple
          accept="image/*,video/*"
          aria-label="从相册选择照片或视频"
          onChange={(event) => void uploads.selectFiles(Array.from(event.target.files ?? []))}
        />
      </label>

      {uploads.error && <InlineError>{uploads.error}</InlineError>}
      {uploads.loading ? (
        <section className="upload-loading">
          <Skeleton label="正在读取上传记录" lines={5} />
        </section>
      ) : (
        <UploadTaskView
          task={uploads.task}
          progress={uploads.progress}
          busy={uploads.busy}
          onRetryUpload={(item) => void uploads.retryFailedUpload(item)}
          onRetryProcessing={(item) => void uploads.retryFailedProcessing(item)}
          onStop={() => void uploads.stop()}
        />
      )}

      <section className="recent-uploads">
        <header>
          <h2>最近传过的</h2>
          <Button type="button" variant="text" loading={uploads.loading} onClick={() => void uploads.refresh()}>
            刷新
          </Button>
        </header>
        {uploads.recentTasks.length === 0 ? (
          <p>还没有传过照片。</p>
        ) : (
          <div>
            {uploads.recentTasks.map((task) =>
              task.batch ? (
                <article key={task.batch.id}>
                  <span>{formatDateTime(task.batch.createdAt)} · {task.batch.totalCount} 张</span>
                  <strong>{uploadBatchStatusText(task.batch.status)}</strong>
                </article>
              ) : null,
            )}
          </div>
        )}
      </section>
    </main>
  );
}
