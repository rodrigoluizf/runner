package cmd

import (
	"github.com/kyriosdata/assinatura/internal/invoker"
	"github.com/kyriosdata/assinatura/internal/server"
	"github.com/spf13/cobra"
)

var startPort int
var startTimeout int

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Inicia o assinador.jar no modo servidor HTTP",
	Long: `Inicia o assinador.jar como servidor HTTP em background.

Exemplos:
  assinatura start
  assinatura start --port 9090
  assinatura start --port 9090 --timeout 30`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jar, err := invoker.FindJar()
		if err != nil {
			return err
		}
		return server.Start(jar, startPort, startTimeout)
	},
}

func init() {
	startCmd.Flags().IntVar(&startPort, "port", server.DefaultPort, "Porta em que o servidor irá escutar")
	startCmd.Flags().IntVar(&startTimeout, "timeout", 0, "Encerrar automaticamente após N minutos de inatividade (0 = nunca)")
	rootCmd.AddCommand(startCmd)
}
