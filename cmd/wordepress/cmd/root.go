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
	tag      string
)

var rootCmd = &cobra.Command{
	Use:   "wordepress",
	Short: "Technical documentation importer for WordPress",
	Long:  `Technical documentation importer for WordPress`,
}

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish posts into WordPress",
	Long:  `Publish reference documentation or tutorials into WordPress`,
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete posts from WordPress",
	Long:  `Delete reference documentation or tutorials from WordPress`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "", false, "Dry run only - make no remote changes")
	rootCmd.PersistentFlags().StringVarP(&baseURL, "url", "", "http://wordpress.local", "WordPress URL")
	rootCmd.PersistentFlags().StringVarP(&user, "user", "", "", "Username for WordPress authentication")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "", "", "Password for WordPress authentication")
	rootCmd.AddCommand(publishCmd)
	rootCmd.AddCommand(deleteCmd)
}
