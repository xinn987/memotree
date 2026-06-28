// 后端暂缺能力的集中注册表。
// 页面可以保留 demo 版位，但不得散落临时布尔值或伪造持久化成功。

export type FrontendCapability =
  | "transferAdmin"
  | "downloadOriginal"
  | "editCapturedAt"
  | "adjacentMedia"
  | "familyNotes";

type CapabilityDefinition = {
  available: boolean;
  unavailableMessage: string;
};

const capabilityRegistry: Record<FrontendCapability, CapabilityDefinition> = {
  transferAdmin: {
    available: false,
    unavailableMessage: "管理员转让暂未开放",
  },
  downloadOriginal: {
    available: false,
    unavailableMessage: "原图下载暂未开放",
  },
  editCapturedAt: {
    available: false,
    unavailableMessage: "修改拍摄时间暂未开放",
  },
  adjacentMedia: {
    available: false,
    unavailableMessage: "直接打开详情时暂时不能切换前后照片",
  },
  familyNotes: {
    available: false,
    unavailableMessage: "家庭手记暂未开放",
  },
};

export function isCapabilityAvailable(capability: FrontendCapability): boolean {
  return capabilityRegistry[capability].available;
}

export function capabilityMessage(capability: FrontendCapability): string {
  return capabilityRegistry[capability].unavailableMessage;
}
