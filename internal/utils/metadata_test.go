package utils_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"videogen/internal/utils"
)

// 75. Testar extração de metadados com ffprobe mockado
func TestProbeVideoMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Pasta bin temporária para mockar o ffprobe
	mockBinDir := filepath.Join(tmpDir, "mockbin")
	if err := os.MkdirAll(mockBinDir, 0700); err != nil {
		t.Fatalf("falha ao criar mockbin: %v", err)
	}

	ffprobeMockPath := filepath.Join(mockBinDir, "ffprobe")
	ffprobeScript := `#!/bin/sh
echo '{"format": {"duration": "10.000000"}, "streams": [{"width": 1920, "height": 1080, "r_frame_rate": "30/1"}]}'
exit 0
`
	if err := os.WriteFile(ffprobeMockPath, []byte(ffprobeScript), 0755); err != nil {
		t.Fatalf("falha ao mockar ffprobe: %v", err)
	}

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", mockBinDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	dummyFile := filepath.Join(tmpDir, "dummy.mp4")
	if err := os.WriteFile(dummyFile, []byte("data"), 0600); err != nil {
		t.Fatalf("falha ao criar arquivo mp4 dummy: %v", err)
	}

	res, err := utils.ProbeVideoMetadata(dummyFile)
	if err != nil {
		t.Fatalf("ProbeVideoMetadata falhou: %v", err)
	}

	if !strings.Contains(res, "duration") || !strings.Contains(res, "1920") {
		t.Errorf("Esperava metadados válidos contendo duration e 1920, obtido: %s", res)
	}
}
