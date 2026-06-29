package utils_test

import (
	"testing"
	"videogen/internal/utils"
)

func TestDownloadRemoteFile_SSRF_Blocked(t *testing.T) {
	destDir := t.TempDir()
	
	// Testar URL local/loopback
	_, err := utils.DownloadRemoteFile("http://127.0.0.1/test.png", destDir)
	if err == nil {
		t.Fatalf("deveria bloquear loopback 127.0.0.1")
	}

	_, err = utils.DownloadRemoteFile("http://localhost/test.png", destDir)
	if err == nil {
		t.Fatalf("deveria bloquear localhost")
	}
}

func TestDownloadRemoteFile_InvalidProtocol(t *testing.T) {
	destDir := t.TempDir()
	
	_, err := utils.DownloadRemoteFile("ftp://example.com/test.png", destDir)
	if err == nil {
		t.Fatalf("deveria bloquear ftp protocol")
	}
}
