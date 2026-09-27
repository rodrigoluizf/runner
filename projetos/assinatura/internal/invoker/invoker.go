package invoker

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kyriosdata/assinatura/internal/jdk"
)

// Response representa a resposta JSON do assinador.jar.
type Response struct {
	Signature string `json:"signature"`
	Valid     bool   `json:"valid"`
	Message   string `json:"message"`
}

// Format retorna uma representação legível da resposta para exibição no terminal.
func (r *Response) Format() string {
	if r.Valid {
		if r.Signature != "" {
			return fmt.Sprintf("[OK] %s\nAssinatura: %s", r.Message, r.Signature)
		}
		return fmt.Sprintf("[OK] %s", r.Message)
	}
	return fmt.Sprintf("[FALHA] %s", r.Message)
}

// Sign invoca o assinador.jar com a operação de criação de assinatura.
func Sign(content, token string) (*Response, error) {
	args := []string{"sign", "--content", content}
	if token != "" {
		args = append(args, "--token", token)
	}
	return invoke(args)
}

// Validate invoca o assinador.jar com a operação de validação de assinatura.
func Validate(content, signature string) (*Response, error) {
	args := []string{"validate", "--content", content, "--signature", signature}
	return invoke(args)
}

func invoke(args []string) (*Response, error) {
	java, err := jdk.Ensure()
	if err != nil {
		return nil, fmt.Errorf("JDK não disponível: %w", err)
	}

	jar, err := findJar()
	if err != nil {
		return nil, err
	}

	cmdArgs := append([]string{"-jar", jar}, args...)
	out, err := exec.Command(java, cmdArgs...).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("assinador encerrou com erro: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("erro ao executar assinador: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("resposta inválida do assinador: %w", err)
	}
	return &resp, nil
}

func findJar() (string, error) {
	if jar := os.Getenv("ASSINADOR_JAR"); jar != "" {
		if _, err := os.Stat(jar); err == nil {
			return jar, nil
		}
		return "", fmt.Errorf("ASSINADOR_JAR definido mas arquivo não encontrado: %s", jar)
	}

	// Mesmo diretório do executável
	if exe, err := os.Executable(); err == nil {
		jar := filepath.Join(filepath.Dir(exe), "assinador.jar")
		if _, err := os.Stat(jar); err == nil {
			return jar, nil
		}
	}

	// ~/.hubsaude/assinador.jar
	if home, err := os.UserHomeDir(); err == nil {
		jar := filepath.Join(home, ".hubsaude", "assinador.jar")
		if _, err := os.Stat(jar); err == nil {
			return jar, nil
		}
	}

	return "", fmt.Errorf("assinador.jar não encontrado: defina ASSINADOR_JAR ou coloque o JAR no mesmo diretório do CLI")
}
