package cmd

import (
	"github.com/kyriosdata/assinatura/internal/server"
	"github.com/spf13/cobra"
)

var stopPort int

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Encerra o assinador.jar em execução",
	Long: `Envia uma requisição de encerramento ao assinador.jar.

Exemplos:
  assinatura stop
  assinatura stop --port 9090`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return server.Stop(stopPort)
	},
}

func init() {
	stopCmd.Flags().IntVar(&stopPort, "port", server.DefaultPort, "Porta do servidor a encerrar")
	rootCmd.AddCommand(stopCmd)
}
