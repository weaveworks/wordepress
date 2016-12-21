package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var (
	version string
	product string
)

var publishDocumentationCmd = &cobra.Command{
	Use:   "documentation",
	Short: "Publish documentation into WordPress",
	Run: func(cmd *cobra.Command, args []string) {
		if product == "" || tag == "" || version == "" || user == "" || password == "" || len(args) != 1 {
			cmd.UsageFunc()(cmd)
			os.Exit(1)
		}

		path := "/wp-json/wp/v2/documentation"
		query := fmt.Sprintf(
			"filter[meta_query][0][key]=wpcf-product&"+
				"filter[meta_query][0][value]=%s&"+
				"filter[meta_query][1][key]=wpcf-tag&"+
				"filter[meta_query][1][value]=%s", product, tag)

		createOrUpdatePosts(user, password, path, query, product, version, tag, args[0])

	},
}

var deleteDocumentationCmd = &cobra.Command{
	Use:   "documentation",
	Short: "Delete documentation from WordPress",
	Run: func(cmd *cobra.Command, args []string) {
		if product == "" || tag == "" || user == "" || password == "" || len(args) > 0 {
			cmd.UsageFunc()(cmd)
			os.Exit(1)
		}

		path := "/wp-json/wp/v2/documentation"
		query := fmt.Sprintf(
			"filter[meta_query][0][key]=wpcf-product&"+
				"filter[meta_query][0][value]=%s&"+
				"filter[meta_query][1][key]=wpcf-tag&"+
				"filter[meta_query][1][value]=%s", product, tag)

		deletePosts(user, password, path, query, product, version, tag)
	},
}

func init() {
	publishDocumentationCmd.Flags().StringVarP(&version, "version", "", "", "Value for custom post version field")
	publishDocumentationCmd.Flags().StringVarP(&product, "product", "", "", "Value for custom post product field")
	publishDocumentationCmd.Flags().StringVarP(&tag, "tag", "", "", "Value for custom post tag field")
	publishCmd.AddCommand(publishDocumentationCmd)
	deleteDocumentationCmd.Flags().StringVarP(&version, "version", "", "", "Value for custom post version field")
	deleteDocumentationCmd.Flags().StringVarP(&product, "product", "", "", "Value for custom post product field")
	deleteDocumentationCmd.Flags().StringVarP(&tag, "tag", "", "", "Value for custom post tag field")
	deleteCmd.AddCommand(deleteDocumentationCmd)
}
