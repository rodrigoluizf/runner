package jdk

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const hubsaudeDir = ".hubsaude"

// Exe returns the Java binary name for the current OS.
func Exe() string {
	if runtime.GOOS == "windows" {
		return "java.exe"
	}
	return "java"
}

// Ensure returns the path to a Java 21+ executable.
// It checks JAVA_HOME, PATH, and ~/.hubsaude/jdk in order.
// If none is found, downloads JDK 21 from Eclipse Temurin (Adoptium).
func Ensure() (string, error) {
	if java, ok := findSystem(); ok {
		return java, nil
	}
	if java, ok := FindManaged(); ok {
		return java, nil
	}
	return download()
}

// FindManaged returns the Java 21+ binary from ~/.hubsaude/jdk if present.
func FindManaged() (string, bool) {
	base, err := managedBase()
	if err != nil {
		return "", false
	}
	return findIn(base)
}

func findSystem() (string, bool) {
	if javaHome := os.Getenv("JAVA_HOME"); javaHome != "" {
		java := filepath.Join(javaHome, "bin", Exe())
		if Is21Plus(java) {
			return java, true
		}
	}
	if java, err := exec.LookPath(Exe()); err == nil {
		if Is21Plus(java) {
			return java, true
		}
	}
	return "", false
}

func findIn(base string) (string, bool) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return "", false
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		java := filepath.Join(base, e.Name(), "bin", Exe())
		if Is21Plus(java) {
			return java, true
		}
	}
	return "", false
}

// Is21Plus reports whether the given java binary is version 21 or higher.
func Is21Plus(java string) bool {
	if _, err := os.Stat(java); err != nil {
		return false
	}
	out, err := exec.Command(java, "-version").CombinedOutput()
	if err != nil {
		return false
	}
	s := string(out)
	for _, v := range []string{`"21`, `"22`, `"23`, `"24`, `"25`, `"26`} {
		if strings.Contains(s, v) {
			return true
		}
	}
	return false
}

func managedBase() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, hubsaudeDir, "jdk"), nil
}

func adoptiumURL() (string, string) {
	goos := runtime.GOOS
	arch := "x64"
	if runtime.GOARCH == "arm64" {
		arch = "aarch64"
	}

	var platform, ext string
	switch goos {
	case "windows":
		platform, ext = "windows", ".zip"
	case "darwin":
		platform, ext = "mac", ".tar.gz"
	default:
		platform, ext = "linux", ".tar.gz"
	}

	url := fmt.Sprintf(
		"https://api.adoptium.net/v3/binary/latest/21/ga/%s/%s/jdk/hotspot/normal/eclipse",
		platform, arch,
	)
	return url, ext
}

func download() (string, error) {
	url, ext := adoptiumURL()

	base, err := managedBase()
	if err != nil {
		return "", fmt.Errorf("erro ao determinar diretório gerenciado: %w", err)
	}
	if err := os.MkdirAll(base, 0755); err != nil {
		return "", fmt.Errorf("erro ao criar diretório JDK: %w", err)
	}

	fmt.Fprintf(os.Stderr, "JDK 21 não encontrado. Baixando de %s...\n", url)

	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("erro ao baixar JDK: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("falha ao baixar JDK: HTTP %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "jdk-*"+ext)
	if err != nil {
		return "", fmt.Errorf("erro ao criar arquivo temporário: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return "", fmt.Errorf("erro ao salvar JDK: %w", err)
	}
	tmp.Close()

	switch ext {
	case ".tar.gz":
		if err := extractTarGz(tmpPath, base); err != nil {
			return "", fmt.Errorf("erro ao extrair JDK: %w", err)
		}
	case ".zip":
		if err := extractZip(tmpPath, base); err != nil {
			return "", fmt.Errorf("erro ao extrair JDK: %w", err)
		}
	default:
		return "", fmt.Errorf("formato de arquivo não suportado: %s", ext)
	}

	java, ok := findIn(base)
	if !ok {
		return "", fmt.Errorf("JDK instalado em %s mas executável java não encontrado", base)
	}

	fmt.Fprintln(os.Stderr, "JDK 21 instalado com sucesso.")
	return java, nil
}

func extractZip(src, destDir string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("erro ao abrir zip: %w", err)
	}
	defer r.Close()

	clean := filepath.Clean(destDir) + string(os.PathSeparator)

	for _, f := range r.File {
		target := filepath.Join(destDir, filepath.FromSlash(f.Name))
		if !strings.HasPrefix(target, clean) {
			return fmt.Errorf("entrada zip suspeita (path traversal): %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, f.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			out.Close()
			return err
		}

		_, copyErr := io.Copy(out, rc)
		rc.Close()
		out.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}

func extractTarGz(src, destDir string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("erro ao abrir gzip: %w", err)
	}
	defer gz.Close()

	clean := filepath.Clean(destDir) + string(os.PathSeparator)
	tr := tar.NewReader(gz)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("erro ao ler tar: %w", err)
		}

		target := filepath.Join(destDir, filepath.FromSlash(hdr.Name))
		if !strings.HasPrefix(target, clean) {
			return fmt.Errorf("entrada tar suspeita (path traversal): %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(out, tr)
			out.Close()
			if copyErr != nil {
				return copyErr
			}
		case tar.TypeSymlink:
			os.Remove(target)
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return err
			}
		}
	}
	return nil
}
