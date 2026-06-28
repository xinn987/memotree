// 统一 HTTP 客户端：页面只处理业务状态，不重复拼接 API 前缀和解析错误。

const apiBasePath = import.meta.env.VITE_API_BASE_PATH ?? "/api";

type ErrorPayload = {
  error?: string;
  message?: string;
};

// ApiError 保留 HTTP 状态码，便于页面区分无权限、失效和普通网络错误。
export class ApiError extends Error {
  readonly status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

// requestJSON 只承诺 JSON/空响应；文件直传由上传 feature 使用独立 XHR。
export async function requestJSON<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set("Accept", "application/json");
  if (options.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  let response: Response;
  try {
    response = await fetch(apiURL(path), {
      ...options,
      credentials: "include",
      headers,
    });
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError") {
      throw error;
    }
    throw new ApiError("暂时连不上家里的相册，请稍后再试", 0);
  }

  if (!response.ok) {
    throw new ApiError(await readErrorMessage(response), response.status);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  const text = await response.text();
  return (text ? JSON.parse(text) : undefined) as T;
}

// apiURL 接受现有 feature 使用的斜杠路径，避免出现双斜杠。
function apiURL(path: string): string {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${apiBasePath.replace(/\/$/, "")}${normalizedPath}`;
}

// 错误响应可能来自 Go API 或代理层，因此同时兼容 JSON 与纯文本。
async function readErrorMessage(response: Response): Promise<string> {
  const fallback = response.status >= 500 ? "家里的相册暂时忙不过来，请稍后再试" : "这次操作没有完成，请再试一次";
  const text = await response.text();
  if (!text) {
    return fallback;
  }

  try {
    const payload = JSON.parse(text) as ErrorPayload;
    return payload.error?.trim() || payload.message?.trim() || fallback;
  } catch {
    return text.trim() || fallback;
  }
}
