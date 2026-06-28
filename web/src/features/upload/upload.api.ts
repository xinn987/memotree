// 上传 API 适配器：保留现有 intent、直传回报、重试和停止契约。
import type {
  UploadBatch,
  UploadIntentResponse,
  UploadItem,
  UploadTask,
  UploadTaskListResponse,
} from "../../api/contracts";
import { requestJSON } from "../../api/client";

export type UploadItemMutation = {
  batch: UploadBatch;
  item: UploadItem;
};

export function getActiveUploadTask(familyId: number) {
  return requestJSON<UploadTask>(`/families/${familyId}/uploads/active`);
}

export function listRecentUploadTasks(familyId: number) {
  return requestJSON<UploadTaskListResponse>(`/families/${familyId}/uploads/recent`);
}

export function createUploadIntent(familyId: number, files: File[]) {
  return requestJSON<UploadIntentResponse>(`/families/${familyId}/media/upload-intents`, {
    method: "POST",
    body: JSON.stringify({
      files: files.map((file) => ({
        filename: file.name,
        contentType: file.type || fallbackContentType(file.name),
        byteSize: file.size,
      })),
    }),
  });
}

export function completeUpload(familyId: number, batchId: number, itemId: number) {
  return requestJSON<UploadItemMutation>(
    `/families/${familyId}/uploads/${batchId}/items/${itemId}/complete-upload`,
    { method: "POST", body: JSON.stringify({}) },
  );
}

export function failUpload(familyId: number, batchId: number, itemId: number, errorMessage: string) {
  return requestJSON<UploadItemMutation>(
    `/families/${familyId}/uploads/${batchId}/items/${itemId}/fail-upload`,
    { method: "POST", body: JSON.stringify({ errorMessage }) },
  );
}

export function retryUpload(familyId: number, batchId: number, itemId: number) {
  return requestJSON<UploadItemMutation>(
    `/families/${familyId}/uploads/${batchId}/items/${itemId}/retry-upload`,
    { method: "POST", body: JSON.stringify({}) },
  );
}

export function retryProcessing(familyId: number, batchId: number, itemId: number) {
  return requestJSON<UploadItemMutation>(
    `/families/${familyId}/uploads/${batchId}/items/${itemId}/retry-processing`,
    { method: "POST", body: JSON.stringify({}) },
  );
}

export function stopUploadTask(familyId: number, batchId: number) {
  return requestJSON<UploadTask>(`/families/${familyId}/uploads/${batchId}/stop`, {
    method: "POST",
    body: JSON.stringify({}),
  });
}

// 原文件直传对象存储不走 JSON client，XHR 用于获取可靠的上传进度。
export function putOriginalFile(item: UploadItem, file: File, onProgress: (progress: number) => void): Promise<void> {
  return new Promise((resolve, reject) => {
    if (!item.uploadUrl) {
      reject(new Error("这张照片缺少上传地址，请重新选择"));
      return;
    }
    const request = new XMLHttpRequest();
    request.open(item.method ?? "PUT", item.uploadUrl);
    request.setRequestHeader("Content-Type", item.contentType || file.type || "application/octet-stream");
    request.upload.onprogress = (event) => {
      if (event.lengthComputable) {
        onProgress((event.loaded / event.total) * 100);
      }
    };
    request.onload = () => {
      if (request.status >= 200 && request.status < 300) {
        resolve();
      } else {
        reject(new Error(`照片没有传上去（${request.status}）`));
      }
    };
    request.onerror = () => reject(new Error("网络中断，这张照片没传上去"));
    request.onabort = () => reject(new Error("这张照片已经停下"));
    request.send(file);
  });
}

export function isSupportedImageFile(file: File): boolean {
  const type = file.type.toLowerCase();
  const name = file.name.toLowerCase();
  return (
    ["image/jpeg", "image/png", "image/gif"].includes(type) ||
    [".jpg", ".jpeg", ".png", ".gif"].some((extension) => name.endsWith(extension))
  );
}

function fallbackContentType(filename: string): string {
  const name = filename.toLowerCase();
  if (name.endsWith(".png")) return "image/png";
  if (name.endsWith(".gif")) return "image/gif";
  return "image/jpeg";
}
