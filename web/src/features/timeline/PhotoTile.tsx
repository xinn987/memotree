// 单张媒体保持真实比例；元信息只在 hover/focus 中浮出，但详情页始终可访问。
import { Play } from "lucide-react";
import { Link } from "react-router-dom";
import type { TimelineMedia } from "../../api/contracts";
import { formatDateTime, renditionAspectRatio } from "../../utils/format";

type PhotoTileProps = {
  familyId: number;
  media: TimelineMedia;
  mediaIds: number[];
};

export function PhotoTile({ familyId, media, mediaIds }: PhotoTileProps) {
  const isVideo = media.mediaType === "video";
  const mediaLabel = media.uploadedBy.displayName || "家人";

  return (
    <Link
      className="photo"
      to={`/families/${familyId}/media/${media.id}`}
      state={{ mediaIds }}
      aria-label={`打开${mediaLabel}上传的${isVideo ? "视频" : "照片"}`}
    >
      <img
        src={media.thumbnail.url}
        alt={`${mediaLabel}上传的${isVideo ? "视频预览" : "照片"}`}
        loading="lazy"
        style={{ aspectRatio: renditionAspectRatio(media.thumbnail) }}
      />
      {isVideo && (
        <span className="photo__video-mark" aria-label="视频">
          <Play aria-hidden="true" size={13} fill="currentColor" />
        </span>
      )}
      <span className="photo__meta">
        <strong>{mediaLabel}</strong>
        <span>{formatDateTime(media.capturedAt ?? media.uploadedAt)}</span>
      </span>
    </Link>
  );
}
