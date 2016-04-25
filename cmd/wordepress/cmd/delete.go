package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/weaveworks/wordepress"
	"log"
	"os"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a site from WordPress",
	Long:  `Delete a site from WordPress`,
	Run: func(cmd *cobra.Command, args []string) {
		if product == "" || version == "" || user == "" || password == "" {
			cmd.Usage()
			os.Exit(1)
		}

		endpoint := fmt.Sprintf("%s/wp-json/wp/v2/documentation", baseURL)
		query := fmt.Sprintf(
			"filter[meta_query][0][key]=wpcf-product&"+
				"filter[meta_query][0][value]=%s&"+
				"filter[meta_query][1][key]=wpcf-version&"+
				"filter[meta_query][1][value]=%s", product, version)

		documents, err := wordepress.GetDocuments(user, password, endpoint, query)
		if err != nil {
			log.Fatalf("Unable to get documents: %v", err)
		}

		for _, document := range documents {
			if document.Product != product || document.Version != version {
				// meta_query filter was ignored, most likely due to wrong plugin version
				log.Printf("Skipping delete of %s due to product/version mismatch. "+
					"Is your plugin up to date?", document.Slug)
				continue
			}

			log.Printf("Deleting document: %s", document.Slug)
			err := wordepress.DeleteDocument(user, password, endpoint, document)
			if err != nil {
				log.Fatalf("Error deleting document: %v", err)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
