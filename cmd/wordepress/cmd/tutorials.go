package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var publishTutorialsCmd = &cobra.Command{
	Use:   "tutorials",
	Short: "Publish tutorials into WordPress",
	Run: func(cmd *cobra.Command, args []string) {
		if user == "" || password == "" || len(args) != 1 {
			cmd.UsageFunc()(cmd)
			os.Exit(1)
		}

		path := "/wp-json/wp/v2/tutorials-post-type"
		query := "filter[meta_query][1][key]=wpcf-tag&" +
			"filter[meta_query][1][value]=" + tag

		createOrUpdatePosts(user, password, path, query, "tutorials", tag, tag, args[0])
	},
}

var deleteTutorialsCmd = &cobra.Command{
	Use:   "tutorials",
	Short: "Delete a tutorials from WordPress",
	Run: func(cmd *cobra.Command, args []string) {
		if user == "" || password == "" || len(args) > 0 {
			cmd.UsageFunc()(cmd)
			os.Exit(1)
		}

		path := "/wp-json/wp/v2/tutorials-post-type"
		query := "filter[meta_query][1][key]=wpcf-tag&" +
			"filter[meta_query][1][value]=" + tag

		deletePosts(user, password, path, query, "tutorials", tag, tag)
	},
}

func init() {
	publishTutorialsCmd.Flags().StringVarP(&tag, "tag", "", "latest", "Value for custom post tag field")
	publishCmd.AddCommand(publishTutorialsCmd)
	deleteTutorialsCmd.Flags().StringVarP(&tag, "tag", "", "latest", "Value for custom post tag field")
	deleteCmd.AddCommand(deleteTutorialsCmd)
}
