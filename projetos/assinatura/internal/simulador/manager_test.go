package simulador

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestPortFree_FreePort(t *testing.T) {
	// Pega uma porta disponível e verifica que está livre após fechar
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	if !portFree(port) {
		t.Errorf("portFree(%d) = false; porta deveria estar livre", port)
	}
}

func TestPortFree_BusyPort(t *testing.T) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	if portFree(port) {
		t.Errorf("portFree(%d) = true; porta deveria estar ocupada", port)
	}
}

func TestStatus_NoPidFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	running, pid := Status()
	if running {
		t.Error("Status() = true; esperado false quando não há arquivo de PID")
	}
	if pid != 0 {
		t.Errorf("Status() PID = %d; esperado 0", pid)
	}
}

func TestStop_NotRunning(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := Stop(); err == nil {
		t.Error("Stop() deveria retornar erro quando simulador não está em execução")
	}
}

func TestStatus_StalePid(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	const impossiblePID = 2147483647
	sf := stateFile()
	data, _ := json.Marshal(State{PID: impossiblePID, Port: 8443})
	if err := os.WriteFile(sf, data, 0644); err != nil {
		t.Fatal(err)
	}

	running, _ := Status()
	if running {
		t.Error("Status() = true para PID inexistente; esperado false")
	}
	if _, err := os.Stat(sf); err == nil {
		t.Error("arquivo de estado stale deveria ter sido removido por Status()")
	}
}

func TestIsAlive_CurrentProcess(t *testing.T) {
	pid := os.Getpid()
	if !isAlive(pid) {
		t.Errorf("isAlive(%d) = false para o próprio processo", pid)
	}
}

func TestIsAlive_InvalidPid(t *testing.T) {
	const impossiblePID = 2147483647
	if isAlive(impossiblePID) {
		t.Errorf("isAlive(%d) = true para PID impossível", impossiblePID)
	}
}

func TestLocalVersion_Missing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if v := localVersion(); v != "" {
		t.Errorf("localVersion() = %q; esperado vazio quando arquivo não existe", v)
	}
}

func TestLocalVersion_Written(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	_ = os.MkdirAll(hubDir(), 0755)
	if err := os.WriteFile(versionFile(), []byte("1.2.3\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if v := localVersion(); v != "1.2.3" {
		t.Errorf("localVersion() = %q; esperado %q", v, "1.2.3")
	}
}

func TestVerifySHA256_Correto(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "test.jar")
	content := []byte("conteudo de teste")
	if err := os.WriteFile(f, content, 0644); err != nil {
		t.Fatal(err)
	}

	h := sha256.New()
	h.Write(content)
	expected := hex.EncodeToString(h.Sum(nil))

	if err := verifySHA256(f, expected); err != nil {
		t.Errorf("verifySHA256() erro inesperado: %v", err)
	}
}

func TestVerifySHA256_Incorreto(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "test.jar")
	if err := os.WriteFile(f, []byte("conteudo"), 0644); err != nil {
		t.Fatal(err)
	}

	err := verifySHA256(f, "0000000000000000000000000000000000000000000000000000000000000000")
	if err == nil {
		t.Error("verifySHA256() deveria retornar erro para checksum incorreto")
	}
}

func TestVerifySHA256_ArquivoInexistente(t *testing.T) {
	err := verifySHA256("/nao/existe/arquivo.jar", "abc123")
	if err == nil {
		t.Error("verifySHA256() deveria retornar erro para arquivo inexistente")
	}
}

func TestEnsureJar_SourceURL_ServidorIndisponivel(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	_, err := EnsureJar("http://localhost:19998/simulador.jar")
	if err == nil {
		t.Error("EnsureJar com source inválido deveria retornar erro")
	}
}
