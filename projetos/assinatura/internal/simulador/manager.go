package simulador

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kyriosdata/assinatura/internal/jdk"
)

const (
	releaseURL  = "https://raw.githubusercontent.com/kyriosdata/runner/main/release.json"
	hubsaudeDir = ".hubsaude"
)

// State persiste informações do simulador em ~/.hubsaude/simulador.json.
type State struct {
	PID  int `json:"pid"`
	Port int `json:"port"`
}

type releaseManifest struct {
	Jar struct {
		URL     string `json:"url"`
		Version string `json:"version"`
		SHA256  string `json:"sha256,omitempty"`
	} `json:"jar"`
}

func hubDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, hubsaudeDir)
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func stateFile() string {
	return filepath.Join(hubDir(), "simulador.json")
}

func jarFile() string {
	return filepath.Join(hubDir(), "simulador.jar")
}

func versionFile() string {
	return filepath.Join(hubDir(), "simulador.version")
}

func saveState(s State) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(stateFile(), data, 0644)
}

func loadState() (*State, error) {
	data, err := os.ReadFile(stateFile())
	if err != nil {
		return nil, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func portFree(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

func fetchManifest() (*releaseManifest, error) {
	resp, err := http.Get(releaseURL) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar release.json: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("release.json: HTTP %d", resp.StatusCode)
	}
	var m releaseManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("release.json inválido: %w", err)
	}
	return &m, nil
}

func localVersion() string {
	data, err := os.ReadFile(versionFile())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d ao baixar %s", resp.StatusCode, url)
	}
	tmp := dest + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	f.Close()
	return os.Rename(tmp, dest)
}

func verifySHA256(path, expected string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if got != strings.ToLower(expected) {
		return fmt.Errorf("checksum SHA-256 divergente: esperado %s, obtido %s", expected, got)
	}
	return nil
}

// EnsureJar garante que simulador.jar está presente e atualizado.
// Se source for não-vazio, baixa diretamente dessa URL sem verificar versão.
// Caso contrário, busca release.json para comparar versões e baixa apenas se necessário.
func EnsureJar(source string) (string, error) {
	jar := jarFile()

	if source != "" {
		fmt.Fprintf(os.Stderr, "Baixando simulador.jar de %s...\n", source)
		if err := downloadFile(source, jar); err != nil {
			return "", fmt.Errorf("erro ao baixar simulador.jar: %w", err)
		}
		_ = os.WriteFile(versionFile(), []byte("custom"), 0644)
		fmt.Fprintln(os.Stderr, "simulador.jar baixado com sucesso.")
		return jar, nil
	}

	manifest, err := fetchManifest()
	if err != nil {
		if _, e := os.Stat(jar); e == nil {
			fmt.Fprintf(os.Stderr, "Aviso: não foi possível verificar atualizações (%v). Usando versão local.\n", err)
			return jar, nil
		}
		return "", fmt.Errorf("simulador.jar não encontrado e não foi possível baixar: %w", err)
	}

	if manifest.Jar.URL == "" {
		return "", fmt.Errorf("release.json não contém URL do simulador.jar")
	}

	if localVersion() == manifest.Jar.Version {
		if _, e := os.Stat(jar); e == nil {
			return jar, nil
		}
	}

	fmt.Fprintf(os.Stderr, "Baixando simulador.jar v%s...\n", manifest.Jar.Version)
	if err := downloadFile(manifest.Jar.URL, jar); err != nil {
		return "", fmt.Errorf("erro ao baixar simulador.jar: %w", err)
	}

	if manifest.Jar.SHA256 != "" {
		if err := verifySHA256(jar, manifest.Jar.SHA256); err != nil {
			_ = os.Remove(jar)
			return "", fmt.Errorf("verificação de integridade falhou: %w", err)
		}
	}

	_ = os.WriteFile(versionFile(), []byte(manifest.Jar.Version), 0644)
	fmt.Fprintln(os.Stderr, "simulador.jar baixado com sucesso.")
	return jar, nil
}

// Start inicia o simulador na porta indicada.
// source, se não-vazio, especifica uma URL alternativa para download do simulador.jar.
func Start(port int, source string) error {
	if !portFree(port) {
		return fmt.Errorf("porta %d ocupada", port)
	}

	jar, err := EnsureJar(source)
	if err != nil {
		return err
	}

	java, err := jdk.Ensure()
	if err != nil {
		return fmt.Errorf("JDK não disponível: %w", err)
	}

	cmd := exec.Command(java, "-jar", jar, "--server.port="+strconv.Itoa(port)) //nolint:gosec
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("erro ao iniciar simulador: %w", err)
	}
	return saveState(State{PID: cmd.Process.Pid, Port: port})
}

// Stop encerra o simulador via endpoint /shutdown; usa kill como fallback.
func Stop() error {
	state, err := loadState()
	if err != nil {
		return fmt.Errorf("simulador não está em execução")
	}

	shutdownURL := fmt.Sprintf("http://localhost:%d/shutdown", state.Port)
	if resp, err := http.Post(shutdownURL, "application/json", nil); err == nil { //nolint:gosec
		resp.Body.Close()
		_ = os.Remove(stateFile())
		return nil
	}

	// Fallback: encerra o processo diretamente
	proc, err := os.FindProcess(state.PID)
	if err != nil {
		_ = os.Remove(stateFile())
		return err
	}
	_ = proc.Kill()
	_ = os.Remove(stateFile())
	return nil
}

// Status informa se o simulador está em execução e seu PID.
// Remove o arquivo de estado se o processo não estiver mais vivo.
func Status() (bool, int) {
	state, err := loadState()
	if err != nil {
		return false, 0
	}
	if !isAlive(state.PID) {
		_ = os.Remove(stateFile())
		return false, 0
	}
	return true, state.PID
}

// InfoHTTP consulta /api/info no simulador em execução e retorna a resposta.
func InfoHTTP() (string, error) {
	state, err := loadState()
	if err != nil {
		return "", fmt.Errorf("simulador não está em execução")
	}
	url := fmt.Sprintf("http://localhost:%d/api/info", state.Port)
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("simulador não está respondendo na porta %d: %w", state.Port, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
