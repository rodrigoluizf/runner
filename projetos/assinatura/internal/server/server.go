package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/kyriosdata/assinatura/internal/jdk"
)

const DefaultPort = 8080

// State persiste informações do servidor em ~/.hubsaude/assinador.json.
type State struct {
	PID  int `json:"pid"`
	Port int `json:"port"`
}

// BaseURL retorna a URL base do servidor na porta indicada.
func BaseURL(port int) string {
	return fmt.Sprintf("http://localhost:%d", port)
}

// IsAlive verifica se há um servidor respondendo na porta indicada.
func IsAlive(port int) bool {
	resp, err := http.Get(fmt.Sprintf("%s/health", BaseURL(port))) //nolint:gosec
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// ActiveBaseURL retorna a URL do servidor ativo registrado no estado local,
// ou vazio se nenhum servidor estiver respondendo.
func ActiveBaseURL() (string, bool) {
	state, err := loadState()
	if err != nil {
		return "", false
	}
	if IsAlive(state.Port) {
		return BaseURL(state.Port), true
	}
	return "", false
}

// Start inicia o assinador.jar como servidor HTTP em background.
func Start(jar string, port, timeoutMinutes int) error {
	if IsAlive(port) {
		fmt.Printf("assinador.jar já está em execução na porta %d\n", port)
		return nil
	}

	java, err := jdk.Ensure()
	if err != nil {
		return fmt.Errorf("JDK não disponível: %w", err)
	}

	args := []string{"-jar", jar, "server", "--port", fmt.Sprintf("%d", port)}
	if timeoutMinutes > 0 {
		args = append(args, "--timeout", fmt.Sprintf("%d", timeoutMinutes))
	}

	cmd := exec.Command(java, args...) //nolint:gosec
	cmd.SysProcAttr = daemonAttr()

	logPath := logFilePath()
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err == nil {
		if lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			cmd.Stdout = lf
			cmd.Stderr = lf
		}
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("erro ao iniciar assinador.jar: %w", err)
	}

	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if IsAlive(port) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if !IsAlive(port) {
		_ = cmd.Process.Kill()
		return fmt.Errorf("assinador.jar não respondeu na porta %d (veja %s)", port, logFilePath())
	}

	state := State{PID: cmd.Process.Pid, Port: port}
	if err := saveState(state); err != nil {
		fmt.Fprintf(os.Stderr, "aviso: não foi possível salvar estado: %v\n", err)
	}
	_ = cmd.Process.Release()

	fmt.Printf("assinador.jar iniciado na porta %d (PID %d)\n", port, state.PID)
	return nil
}

// Stop encerra o servidor enviando POST /shutdown.
func Stop(port int) error {
	if !IsAlive(port) {
		clearState()
		return fmt.Errorf("nenhum servidor em execução na porta %d", port)
	}

	resp, err := http.Post( //nolint:gosec
		fmt.Sprintf("%s/shutdown", BaseURL(port)),
		"application/json",
		nil,
	)
	if err != nil {
		return fmt.Errorf("erro ao enviar shutdown: %w", err)
	}
	defer resp.Body.Close()

	clearState()
	fmt.Printf("assinador.jar na porta %d encerrado\n", port)
	return nil
}

// ShowStatus exibe o status atual do servidor.
func ShowStatus(port int) error {
	if !IsAlive(port) {
		if state, err := loadState(); err == nil && state.Port == port {
			fmt.Printf("assinador.jar (porta %d, PID %d): não está respondendo\n", state.Port, state.PID)
		} else {
			fmt.Printf("assinador.jar não está em execução na porta %d\n", port)
		}
		return nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/health", BaseURL(port))) //nolint:gosec
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var info map[string]any
	_ = json.Unmarshal(body, &info)

	pid := "desconhecido"
	if p, ok := info["pid"].(float64); ok {
		pid = fmt.Sprintf("%.0f", p)
	}

	fmt.Printf("assinador.jar em execução na porta %d (PID %s)\n", port, pid)
	return nil
}

func statePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".hubsaude", "assinador.json")
}

func logFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".hubsaude", "assinador.log")
}

func saveState(s State) error {
	path := statePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func loadState() (*State, error) {
	data, err := os.ReadFile(statePath())
	if err != nil {
		return nil, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func clearState() {
	_ = os.Remove(statePath())
}
