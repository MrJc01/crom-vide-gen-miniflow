package utils

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 12. Suporte a leitura de arquivos remotos (URLs)
// 99. Prevenir SSRF
func DownloadRemoteFile(rawURL string, destDir string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("URL inválida: %w", err)
	}

	// 99. SSRF Protection: Apenas HTTP e HTTPS permitidos
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("esquema de protocolo não permitido: %s", u.Scheme)
	}

	// Resolver hostname para IPs
	ips, err := net.LookupIP(u.Hostname())
	if err != nil {
		return "", fmt.Errorf("falha ao resolver host %s: %w", u.Hostname(), err)
	}

	// 99. SSRF Protection: Bloquear IPs privados/locais
	for _, ip := range ips {
		if isPrivateIP(ip) || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return "", fmt.Errorf("acesso a endereço IP restrito não permitido: %s", ip.String())
		}
	}

	// Fazer requisição HTTP segura
	client := &http.Client{
		Timeout: 10 * time.Second, // Timeout razoável
	}

	resp, err := client.Get(rawURL)
	if err != nil {
		return "", fmt.Errorf("falha ao fazer download do arquivo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("resposta HTTP inválida: %s", resp.Status)
	}

	// 89. Limite de tamanho de download (max 50MB)
	const MaxDownloadSize = 50 * 1024 * 1024
	limitReader := io.LimitReader(resp.Body, MaxDownloadSize)

	// Nome do arquivo temporário a partir do path da URL
	fileName := filepath.Base(u.Path)
	if fileName == "." || fileName == "/" || strings.TrimSpace(fileName) == "" {
		fileName = "downloaded_asset"
	}

	destPath := filepath.Join(destDir, fileName)
	
	// Sanitização de Directory Traversal
	destPath, err = SanitizePath(destDir, destPath)
	if err != nil {
		return "", err
	}

	// #nosec G304
	out, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("falha ao criar arquivo de destino: %w", err)
	}
	defer out.Close()

	copied, err := io.Copy(out, limitReader)
	if err != nil {
		return "", fmt.Errorf("falha ao salvar conteúdo: %w", err)
	}

	// Se a leitura foi limitada e pode ter cortado o arquivo
	if copied >= MaxDownloadSize {
		_ = os.Remove(destPath)
		return "", errors.New("tamanho máximo de download excedido")
	}

	// 98. MIME validation
	if err := ValidateMIMEType(destPath); err != nil {
		_ = os.Remove(destPath)
		return "", fmt.Errorf("arquivo baixado tem formato inválido: %w", err)
	}

	return destPath, nil
}

func isPrivateIP(ip net.IP) bool {
	if ip.To4() != nil {
		// RFC 1918
		privateBlocks := []string{
			"10.0.0.0/8",
			"172.16.0.0/12",
			"192.168.0.0/16",
		}
		for _, block := range privateBlocks {
			_, subnet, _ := net.ParseCIDR(block)
			if subnet.Contains(ip) {
				return true
			}
		}
	}
	return false
}
