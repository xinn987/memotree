package logging

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigureWritesStandardLogToConsoleAndFile(t *testing.T) {
	var console bytes.Buffer
	logDir := t.TempDir()

	cleanup, err := Configure(ServiceConfig{
		ServiceName:   "api",
		LogDir:        logDir,
		ConsoleWriter: &console,
	})
	if err != nil {
		t.Fatalf("configure logging: %v", err)
	}
	defer cleanup()

	log.Print("api log line")

	fileContent, err := os.ReadFile(filepath.Join(logDir, "api.log"))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(console.String(), "api log line") {
		t.Fatalf("expected console output to contain log line, got %q", console.String())
	}
	if !strings.Contains(string(fileContent), "api log line") {
		t.Fatalf("expected file output to contain log line, got %q", string(fileContent))
	}
}
