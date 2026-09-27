package cmd

import (
	"fmt"

	sim "github.com/kyriosdata/assinatura/internal/simulador"
	"github.com/spf13/cobra"
)

var simPort int
var simSource string

var simuladorCmd = &cobra.Command{
	Use:   "simulador",
	Short: "Gerencia o simulador.jar do HubSaúde",
}

var simStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Inicia o simulador",
	Long: `Inicia o simulador.jar do HubSaúde.

Se o simulador.jar não estiver disponível localmente, é baixado automaticamente
do repositório da disciplina. Use --source para especificar uma URL alternativa.

Exemplos:
  assinatura simulador start
  assinatura simulador start --port 9443
  assinatura simulador start --source https://exemplo.com/simulador.jar`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sim.Start(simPort, simSource); err != nil {
			return err
		}
		fmt.Println("[OK] Simulador iniciado")
		return nil
	},
}

var simStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Encerra o simulador",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sim.Stop(); err != nil {
			return err
		}
		fmt.Println("[OK] Simulador encerrado")
		return nil
	},
}

var simStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Exibe status do simulador",
	RunE: func(cmd *cobra.Command, args []string) error {
		running, pid := sim.Status()
		if running {
			fmt.Printf("[OK] Simulador ativo (PID %d)\n", pid)
		} else {
			fmt.Println("[INFO] Simulador parado")
		}
		return nil
	},
}

func init() {
	simStartCmd.Flags().IntVar(&simPort, "port", 8443, "Porta do simulador")
	simStartCmd.Flags().StringVar(&simSource, "source", "", "URL alternativa para download do simulador.jar")

	simuladorCmd.AddCommand(simStartCmd)
	simuladorCmd.AddCommand(simStopCmd)
	simuladorCmd.AddCommand(simStatusCmd)

	rootCmd.AddCommand(simuladorCmd)
}
