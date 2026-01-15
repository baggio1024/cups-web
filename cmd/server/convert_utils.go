package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func convertOfficeToPDF(ctx context.Context, inputPath string) (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "convert-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(tmpDir) }

	cmd := exec.CommandContext(ctx, "libreoffice", "--headless", "--convert-to", "pdf", "--outdir", tmpDir, inputPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("conversion failed: %w - %s", err, string(out))
	}

	base := filepath.Base(inputPath)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	outPath := filepath.Join(tmpDir, name+".pdf")
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		matches, _ := filepath.Glob(filepath.Join(tmpDir, "*.pdf"))
		if len(matches) == 0 {
			cleanup()
			return "", nil, fmt.Errorf("conversion produced no PDF")
		}
		outPath = matches[0]
	}

	return outPath, cleanup, nil
}

func convertTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 60*time.Second)
}
