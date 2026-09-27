package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// resetFlags restaura todas as flags de cmd (e subcomandos) ao valor padrão.
// Os comandos Cobra são singletons de pacote, então, sem este reset, o estado
// de uma flag (valor e marca Changed) vaza de um teste para o seguinte —
// fazendo, por exemplo, MarkFlagRequired passar indevidamente.
func resetFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
	for _, sub := range cmd.Commands() {
		resetFlags(sub)
	}
}

// runCLI executa o rootCmd com os args fornecidos e retorna a saída combinada e o erro.
func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	resetFlags(rootCmd)
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestVersionCmd_ExibeVersao(t *testing.T) {
	out, err := runCLI(t, "version")
	if err != nil {
		t.Fatalf("version retornou erro: %v", err)
	}
	// Sem injeção de -ldflags (build de desenvolvimento), a variável version
	// mantém o valor padrão "dev". Em release, o CI sobrescreve via ldflags.
	if !strings.Contains(out, "dev") {
		t.Errorf("saída do version não contém a versão de desenvolvimento %q: %q", "dev", out)
	}
}

func TestSignCmd_ContentObrigatorio(t *testing.T) {
	_, err := runCLI(t, "sign")
	if err == nil {
		t.Error("sign sem --content deveria retornar erro")
	}
}

func TestValidateCmd_ContentObrigatorio(t *testing.T) {
	_, err := runCLI(t, "validate", "--signature", "SIG==")
	if err == nil {
		t.Error("validate sem --content deveria retornar erro")
	}
}

func TestValidateCmd_SignatureObrigatoria(t *testing.T) {
	_, err := runCLI(t, "validate", "--content", "doc")
	if err == nil {
		t.Error("validate sem --signature deveria retornar erro")
	}
}

func TestSimuladorCmd_ExibeSubcomandos(t *testing.T) {
	out, _ := runCLI(t, "simulador", "--help")
	for _, sub := range []string{"start", "stop", "status"} {
		if !strings.Contains(out, sub) {
			t.Errorf("help do simulador deveria mencionar subcomando %q, mas não encontrado em: %q", sub, out)
		}
	}
}

func TestStartCmd_HelpMencaoTimeout(t *testing.T) {
	out, _ := runCLI(t, "start", "--help")
	if !strings.Contains(out, "timeout") {
		t.Errorf("help do start deveria mencionar --timeout, mas não encontrado em: %q", out)
	}
}

func TestSimuladorStartCmd_HelpMencaoSource(t *testing.T) {
	out, _ := runCLI(t, "simulador", "start", "--help")
	if !strings.Contains(out, "source") {
		t.Errorf("help do simulador start deveria mencionar --source, mas não encontrado em: %q", out)
	}
}
