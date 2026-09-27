package cmd

import (
	"github.com/kyriosdata/assinatura/internal/server"
	"github.com/spf13/cobra"
)

var statusPort int

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Exibe o status do assinador.jar",
	Long: `Verifica se o assinador.jar está em execução e exibe informações do processo.

Exemplos:
  assinatura status
  assinatura status --port 9090`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return server.ShowStatus(statusPort)
	},
}

func init() {
	statusCmd.Flags().IntVar(&statusPort, "port", server.DefaultPort, "Porta do servidor a verificar")
	rootCmd.AddCommand(statusCmd)
}
