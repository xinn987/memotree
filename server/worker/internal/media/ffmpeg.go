package media

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	defaultFFmpegPath  = "ffmpeg"
	defaultFFprobePath = "ffprobe"
	videoDisplayMaxW   = 1280
)

var execCommandContext = exec.CommandContext

type VideoToolVersions struct {
	FFmpeg  string
	FFprobe string
}

// FFmpegTranscoder 使用本机 FFmpeg 生成浏览器友好的视频派生资源。
type FFmpegTranscoder struct {
	Path        string
	FFprobePath string
	TempDir     string
}

// CheckVideoTools 在 Worker 启动阶段验证视频处理依赖，避免领取视频任务后才失败。
func CheckVideoTools(ctx context.Context, ffmpegPath string, ffprobePath string) (VideoToolVersions, error) {
	if ffmpegPath == "" {
		ffmpegPath = defaultFFmpegPath
	}
	if ffprobePath == "" {
		ffprobePath = defaultFFprobePath
	}
	ffmpegVersion, err := versionLine(ctx, ffmpegPath)
	if err != nil {
		return VideoToolVersions{}, fmt.Errorf("check ffmpeg %q: %w", ffmpegPath, err)
	}
	ffprobeVersion, err := versionLine(ctx, ffprobePath)
	if err != nil {
		return VideoToolVersions{}, fmt.Errorf("check ffprobe %q: %w", ffprobePath, err)
	}
	return VideoToolVersions{
		FFmpeg:  ffmpegVersion,
		FFprobe: ffprobeVersion,
	}, nil
}

func (t FFmpegTranscoder) TranscodeVideo(ctx context.Context, input VideoTranscodeInput) (VideoTranscodeOutput, error) {
	if len(input.Original) == 0 {
		return VideoTranscodeOutput{}, errors.New("video original is empty")
	}
	workDir, err := os.MkdirTemp(t.TempDir, "memotree-video-*")
	if err != nil {
		return VideoTranscodeOutput{}, err
	}
	defer os.RemoveAll(workDir)

	inputPath := filepath.Join(workDir, "input"+safeVideoExtension(input.OriginalFilename))
	thumbnailPath := filepath.Join(workDir, "thumbnail.jpg")
	displayPath := filepath.Join(workDir, "display.mp4")
	if err := os.WriteFile(inputPath, input.Original, 0600); err != nil {
		return VideoTranscodeOutput{}, err
	}

	if err := t.run(ctx,
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-i", inputPath,
		"-frames:v", "1",
		"-vf", "scale=w='min(320,iw)':h=-2",
		thumbnailPath,
	); err != nil {
		return VideoTranscodeOutput{}, err
	}
	if err := t.run(ctx,
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-i", inputPath,
		"-map", "0:v:0",
		"-map", "0:a?",
		"-vf", fmt.Sprintf("scale=w='min(%d,iw)':h=-2", videoDisplayMaxW),
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-crf", "23",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", "128k",
		"-movflags", "+faststart",
		displayPath,
	); err != nil {
		return VideoTranscodeOutput{}, err
	}

	thumbnail, err := os.ReadFile(thumbnailPath)
	if err != nil {
		return VideoTranscodeOutput{}, err
	}
	display, err := os.ReadFile(displayPath)
	if err != nil {
		return VideoTranscodeOutput{}, err
	}
	thumbWidth, thumbHeight := imageSize(thumbnail)
	width, height, durationMillis := t.probe(ctx, displayPath)
	if width == 0 || height == 0 {
		width = thumbWidth
		height = thumbHeight
	}
	return VideoTranscodeOutput{
		Thumbnail:       thumbnail,
		DisplayVideo:    display,
		Width:           width,
		Height:          height,
		DurationMillis:  durationMillis,
		ThumbnailWidth:  thumbWidth,
		ThumbnailHeight: thumbHeight,
	}, nil
}

func (t FFmpegTranscoder) run(ctx context.Context, args ...string) error {
	path := t.Path
	if path == "" {
		path = defaultFFmpegPath
	}
	var stderr bytes.Buffer
	cmd := execCommandContext(ctx, path, args...)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			return err
		}
		return fmt.Errorf("%w: %s", err, message)
	}
	return nil
}

func (t FFmpegTranscoder) probe(ctx context.Context, path string) (int, int, int64) {
	probePath := t.ffprobePath()
	var stdout bytes.Buffer
	cmd := execCommandContext(ctx, probePath,
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height:format=duration",
		"-of", "json",
		path,
	)
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return 0, 0, 0
	}
	var result struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return 0, 0, 0
	}
	var width, height int
	if len(result.Streams) > 0 {
		width = result.Streams[0].Width
		height = result.Streams[0].Height
	}
	durationMillis := parseDurationMillis(result.Format.Duration)
	return width, height, durationMillis
}

func (t FFmpegTranscoder) ffprobePath() string {
	if t.FFprobePath != "" {
		return t.FFprobePath
	}
	if t.Path == "" || t.Path == defaultFFmpegPath {
		return defaultFFprobePath
	}
	base := filepath.Base(t.Path)
	dir := filepath.Dir(t.Path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	if strings.EqualFold(name, "ffmpeg") {
		if runtime.GOOS == "windows" && ext == "" {
			ext = ".exe"
		}
		return filepath.Join(dir, "ffprobe"+ext)
	}
	return defaultFFprobePath
}

func versionLine(ctx context.Context, path string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := execCommandContext(ctx, path, "-version")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			return "", err
		}
		return "", fmt.Errorf("%w: %s", err, message)
	}
	output := strings.TrimSpace(stdout.String())
	if output == "" {
		output = strings.TrimSpace(stderr.String())
	}
	line, _, _ := strings.Cut(output, "\n")
	return strings.TrimSpace(line), nil
}

func safeVideoExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp4", ".mov", ".m4v", ".webm", ".avi", ".mkv":
		return ext
	default:
		return ".video"
	}
}

func imageSize(body []byte) (int, int) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(body))
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

func parseDurationMillis(value string) int64 {
	if value == "" {
		return 0
	}
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil || seconds <= 0 {
		return 0
	}
	return int64(math.Round(seconds * 1000))
}
