package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	dryRun   bool
	baseURL  string
	user     string
	password string
	product  string
	tag      string
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
	RootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "", false, "Dry run only - make no remote changes")
	RootCmd.PersistentFlags().StringVarP(&baseURL, "url", "", "http://wordpress.local", "WordPress URL")
	RootCmd.PersistentFlags().StringVarP(&user, "user", "", "", "Username for WordPress authentication")
	RootCmd.PersistentFlags().StringVarP(&password, "password", "", "", "Password for WordPress authentication")
	RootCmd.PersistentFlags().StringVarP(&product, "product", "", "", "Value for custom post product field")
	RootCmd.PersistentFlags().StringVarP(&tag, "tag", "", "", "Value for custom post tag field")
}
