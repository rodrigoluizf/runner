package main

import (
	"fmt"
	"os"

	sim "github.com/kyriosdata/assinatura/internal/simulador"
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "simulador",
	Short: "Gerencia o Simulador do HubSaúde",
	Long:  "Gerencia o ciclo de vida do simulador.jar do HubSaúde: iniciar, parar e monitorar.",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Exibe a versão do CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("simulador " + version)
	},
}

var startPort int
var startSource string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Inicia o Simulador do HubSaúde",
	Long: `Inicia o simulador.jar na porta indicada.

O simulador.jar é baixado automaticamente se não estiver disponível localmente
ou se uma versão mais recente estiver disponível.

Exemplos:
  simulador start
  simulador start --port 9443
  simulador start --source https://exemplo.com/simulador.jar`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sim.Start(startPort, startSource); err != nil {
			return err
		}
		fmt.Printf("[OK] Simulador iniciado na porta %d\n", startPort)
		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Encerra o Simulador do HubSaúde",
	Long: `Encerra o Simulador via endpoint /shutdown.

Exemplos:
  simulador stop`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sim.Stop(); err != nil {
			return err
		}
		fmt.Println("[OK] Simulador encerrado")
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Exibe o status do Simulador do HubSaúde",
	Long: `Verifica se o Simulador está em execução e exibe informações via /api/info.

Exemplos:
  simulador status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		running, pid := sim.Status()
		if !running {
			fmt.Println("[INFO] Simulador não está em execução")
			return nil
		}
		fmt.Printf("[OK] Simulador ativo (PID %d)\n", pid)
		if info, err := sim.InfoHTTP(); err == nil {
			fmt.Println(info)
		}
		return nil
	},
}

func init() {
	startCmd.Flags().IntVar(&startPort, "port", 8443, "Porta em que o Simulador irá escutar")
	startCmd.Flags().StringVar(&startSource, "source", "", "URL alternativa para download do simulador.jar")
	rootCmd.AddCommand(versionCmd, startCmd, stopCmd, statusCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
