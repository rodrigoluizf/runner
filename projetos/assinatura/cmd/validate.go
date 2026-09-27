package cmd

import (
	"fmt"

	"github.com/kyriosdata/assinatura/internal/invoker"
	"github.com/kyriosdata/assinatura/internal/server"
	"github.com/spf13/cobra"
)

var validateContent string
var validateSignature string
var validateLocal bool

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Valida uma assinatura digital simulada",
	Long: `Invoca o assinador.jar para verificar se uma assinatura é válida.

Por padrão usa o modo servidor (HTTP) se houver uma instância em execução,
com fallback automático para o modo local. Use --local para forçar modo local.

Exemplos:
  assinatura validate --content "documento" --signature "MOCKED_SIGNATURE_BASE64_=="
  assinatura validate --content "documento" --signature "SIG==" --local`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !validateLocal {
			if baseURL, ok := server.ActiveBaseURL(); ok {
				resp, err := invoker.ValidateHTTP(baseURL, validateContent, validateSignature)
				if err != nil {
					return err
				}
				fmt.Println(resp.Format())
				return nil
			}
		}
		resp, err := invoker.Validate(validateContent, validateSignature)
		if err != nil {
			return err
		}
		fmt.Println(resp.Format())
		return nil
	},
}

func init() {
	validateCmd.Flags().StringVar(&validateContent, "content", "", "Conteúdo original que foi assinado (obrigatório)")
	validateCmd.Flags().StringVar(&validateSignature, "signature", "", "Assinatura a ser validada (obrigatório)")
	validateCmd.Flags().BoolVar(&validateLocal, "local", false, "Forçar invocação direta do assinador.jar (modo local)")
	validateCmd.MarkFlagRequired("content")
	validateCmd.MarkFlagRequired("signature")
	rootCmd.AddCommand(validateCmd)
}
