package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "assinatura",
	Short: "CLI para operações de assinatura digital simulada",
	Long:  `CLI multiplataforma para invocar o assinador.jar e realizar operações de assinatura digital.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
