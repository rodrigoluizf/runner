package cmd

import (
	"fmt"

	"github.com/kyriosdata/assinatura/internal/invoker"
	"github.com/kyriosdata/assinatura/internal/server"
	"github.com/spf13/cobra"
)

var signContent string
var signToken string
var signLocal bool

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Cria uma assinatura digital simulada",
	Long: `Invoca o assinador.jar para criar uma assinatura digital simulada.

Por padrão usa o modo servidor (HTTP) se houver uma instância em execução,
com fallback automático para o modo local. Use --local para forçar modo local.

Exemplos:
  assinatura sign --content "documento a assinar"
  assinatura sign --content "documento" --token "meu-token"
  assinatura sign --content "documento" --local`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !signLocal {
			if baseURL, ok := server.ActiveBaseURL(); ok {
				resp, err := invoker.SignHTTP(baseURL, signContent, signToken)
				if err != nil {
					return err
				}
				fmt.Println(resp.Format())
				return nil
			}
		}
		resp, err := invoker.Sign(signContent, signToken)
		if err != nil {
			return err
		}
		fmt.Println(resp.Format())
		return nil
	},
}

func init() {
	signCmd.Flags().StringVar(&signContent, "content", "", "Conteúdo a ser assinado (obrigatório)")
	signCmd.Flags().StringVar(&signToken, "token", "", "Token de autenticação (opcional)")
	signCmd.Flags().BoolVar(&signLocal, "local", false, "Forçar invocação direta do assinador.jar (modo local)")
	signCmd.MarkFlagRequired("content")
	rootCmd.AddCommand(signCmd)
}
