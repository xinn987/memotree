// HTTP 客户端契约测试：先固定请求、空响应、错误和取消信号的行为。
import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, requestJSON } from "./client";

describe("requestJSON", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("给相对路径补 API 前缀并携带会话", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(requestJSON<{ ok: boolean }>("/health")).resolves.toEqual({ ok: true });
    expect(fetchMock).toHaveBeenCalledWith("/api/health", expect.objectContaining({ credentials: "include" }));
    const requestHeaders = fetchMock.mock.calls[0]?.[1]?.headers as Headers;
    expect(requestHeaders.get("Accept")).toBe("application/json");
  });

  it("把没有响应体的成功请求转换为 undefined", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 204 })));

    await expect(requestJSON<void>("/auth/logout", { method: "POST" })).resolves.toBeUndefined();
  });

  it("把服务端错误转换为带状态码的 ApiError", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ error: "这个邀请已经失效" }), {
          status: 410,
          headers: { "Content-Type": "application/json" },
        }),
      ),
    );

    const promise = requestJSON("/invites/expired/join", { method: "POST" });
    await expect(promise).rejects.toMatchObject({
      name: "ApiError",
      message: "这个邀请已经失效",
      status: 410,
    } satisfies Partial<ApiError>);
  });

  it("把 AbortSignal 原样交给 fetch", async () => {
    const controller = new AbortController();
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ ok: true })));
    vi.stubGlobal("fetch", fetchMock);

    await requestJSON("/health", { signal: controller.signal });

    expect(fetchMock.mock.calls[0]?.[1]?.signal).toBe(controller.signal);
  });
});
