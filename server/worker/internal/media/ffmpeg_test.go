package media

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFFmpegTranscoderDerivesFFprobePath(t *testing.T) {
	transcoder := FFmpegTranscoder{Path: filepath.Join("C:\\tools", "ffmpeg.exe")}

	got := transcoder.ffprobePath()

	if got != filepath.Join("C:\\tools", "ffprobe.exe") {
		t.Fatalf("expected adjacent ffprobe path, got %q", got)
	}
}

func TestSafeVideoExtension(t *testing.T) {
	if got := safeVideoExtension("clip.MOV"); got != ".mov" {
		t.Fatalf("expected mov extension, got %q", got)
	}
	if got := safeVideoExtension("clip.sh"); got != ".video" {
		t.Fatalf("expected unsafe extension fallback, got %q", got)
	}
}

func TestParseDurationMillis(t *testing.T) {
	if got := parseDurationMillis("4.245"); got != 4245 {
		t.Fatalf("expected 4245ms, got %d", got)
	}
	if got := parseDurationMillis("bad"); got != 0 {
		t.Fatalf("expected invalid duration fallback, got %d", got)
	}
}

func TestCheckVideoToolsReturnsVersions(t *testing.T) {
	oldCommand := execCommandContext
	execCommandContext = fakeVideoToolCommand
	defer func() {
		execCommandContext = oldCommand
	}()

	versions, err := CheckVideoTools(context.Background(), "ffmpeg", "ffprobe")

	if err != nil {
		t.Fatalf("check video tools: %v", err)
	}
	if versions.FFmpeg != "ffmpeg version helper" {
		t.Fatalf("expected ffmpeg version line, got %q", versions.FFmpeg)
	}
	if versions.FFprobe != "ffprobe version helper" {
		t.Fatalf("expected ffprobe version line, got %q", versions.FFprobe)
	}
}

func TestCheckVideoToolsReportsFailure(t *testing.T) {
	oldCommand := execCommandContext
	execCommandContext = failingVideoToolCommand
	defer func() {
		execCommandContext = oldCommand
	}()

	_, err := CheckVideoTools(context.Background(), "ffmpeg", "ffprobe")

	if err == nil || !strings.Contains(err.Error(), "ffmpeg") {
		t.Fatalf("expected ffmpeg failure, got %v", err)
	}
}

func fakeVideoToolCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	executable, err := os.Executable()
	if err != nil {
		panic(err)
	}
	commandArgs := []string{"-test.run=TestVideoToolHelperProcess", "--", name}
	commandArgs = append(commandArgs, args...)
	cmd := exec.CommandContext(ctx, executable, commandArgs...)
	cmd.Env = append(os.Environ(), "MEMOTREE_VIDEO_TOOL_HELPER=1")
	return cmd
}

func failingVideoToolCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	executable, err := os.Executable()
	if err != nil {
		panic(err)
	}
	commandArgs := []string{"-test.run=TestVideoToolHelperProcess", "--", name}
	commandArgs = append(commandArgs, args...)
	cmd := exec.CommandContext(ctx, executable, commandArgs...)
	cmd.Env = append(os.Environ(), "MEMOTREE_VIDEO_TOOL_HELPER=fail")
	return cmd
}

func TestVideoToolHelperProcess(t *testing.T) {
	mode := os.Getenv("MEMOTREE_VIDEO_TOOL_HELPER")
	if mode == "" {
		return
	}
	if mode == "fail" {
		os.Exit(7)
	}
	name := "unknown"
	for i, arg := range os.Args {
		if arg == "--" && i+1 < len(os.Args) {
			name = os.Args[i+1]
			break
		}
	}
	_, _ = os.Stdout.WriteString(name + " version helper\n")
	os.Exit(0)
}
