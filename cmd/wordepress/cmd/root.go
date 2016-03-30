package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	baseURL  string
	user     string
	password string
	product  string
	version  string
)

var RootCmd = &cobra.Command{
	Use:   "wordepress",
	Short: "Technical documentation importer for WordPress",
	Long:  `Technical documentation importer for WordPress`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&baseURL, "url", "", "http://wordpress.local", "WordPress URL")
	RootCmd.PersistentFlags().StringVarP(&user, "user", "", "", "Username for WordPress authentication")
	RootCmd.PersistentFlags().StringVarP(&password, "password", "", "", "Password for WordPress authentication")
	RootCmd.PersistentFlags().StringVarP(&product, "product", "", "", "Value for document product field")
	RootCmd.PersistentFlags().StringVarP(&version, "version", "", "", "Value for document version field")
}
