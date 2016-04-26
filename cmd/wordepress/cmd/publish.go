package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/weaveworks/wordepress"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func headImage(image *wordepress.Image) (bool, error) {
	url := baseURL + "/wp-content/uploads/" + image.Hash + image.Extension
	request, err := http.NewRequest("HEAD", url, nil)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return false, err
	}
	response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return false, nil
	}
	return true, nil
}

func postImage(image *wordepress.Image) error {
	url := baseURL + "/wp-json/wp/v2/media"
	name := image.Hash + image.Extension
	request, err := http.NewRequest("POST", url, bytes.NewReader(image.Content))
	request.SetBasicAuth(user, password)
	request.Header.Set("Content-Type", image.MimeType)
	request.Header.Set("Content-Disposition", `attachment; filename="`+name+`"`)
	request.Header.Set("Accept", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("post failed: %v %s", response.Status, string(responseBytes))
	}

	var media wordepress.Media
	err = json.Unmarshal(responseBytes, &media)
	if err != nil {
		return err
	}

	if media.MediaDetails.File != name {
		return fmt.Errorf("duplicate attachment: requested %s, response %s",
			name, media.MediaDetails.File)
	}

	return nil
}

func toMap(rds []*wordepress.Document) map[string]*wordepress.Document {
	rdm := make(map[string]*wordepress.Document)
	for i, _ := range rds {
		rdm[rds[i].Slug] = rds[i]
	}
	return rdm
}

func identical(local *wordepress.Document, remote *wordepress.Document) bool {
	return local.MenuOrder == remote.MenuOrder &&
		local.Title.Raw == remote.Title.Raw &&
		local.Content.Raw == remote.Content.Raw &&
		local.Parent == remote.Parent
}

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a site into WordPress",
	Long:  `Publish a site into WordPress`,
	Run: func(cmd *cobra.Command, args []string) {
		if product == "" || version == "" || user == "" || password == "" || len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		}

		// Load local site
		localDocuments, images, err := wordepress.ParseSite(product, version, args[0])
		if err != nil {
			log.Fatalf("Error parsing site: %v", err)
		}

		// Load remote site. context=edit is required to populate the Raw field
		// of the title and content JSON for comparison with local values
		endpoint := fmt.Sprintf("%s/wp-json/wp/v2/documentation", baseURL)
		query := fmt.Sprintf(
			"context=edit&per_page=100&"+
				"filter[meta_query][0][key]=wpcf-product&"+
				"filter[meta_query][0][value]=%s&"+
				"filter[meta_query][1][key]=wpcf-version&"+
				"filter[meta_query][1][value]=%s", product, version)

		remoteDocuments, err := wordepress.GetDocuments(user, password, endpoint, query)
		if err != nil {
			log.Fatalf("Unable to get JSON documents: %v", err)
		}

		// Create/update documents
		existing := toMap(remoteDocuments)
		for _, localDocument := range localDocuments {
			if localDocument.LocalParent != nil {
				// Pre-order traversal guarantees the remote document will be set
				localDocument.Parent = localDocument.LocalParent.RemoteDocument.ID
			}
			if remoteDocument, ok := existing[localDocument.Slug]; ok {
				if identical(localDocument, remoteDocument) {
					if dryRun {
						log.Printf("Would skip document: %s", localDocument.Slug)
					} else {
						log.Printf("Skipping document: %s", localDocument.Slug)
					}
				} else {
					if dryRun {
						log.Printf("Would update document: %s", localDocument.Slug)
					} else {
						log.Printf("Updating document: %s", localDocument.Slug)
						remoteDocument, err = wordepress.PutDocument(user, password, endpoint, remoteDocument.ID, localDocument)
						if err != nil {
							log.Fatalf("Error updating document: %v", err)
						}
					}
				}
				localDocument.RemoteDocument = remoteDocument
				delete(existing, localDocument.Slug)
			} else {
				if dryRun {
					log.Printf("Would upload document: %s", localDocument.Slug)
					localDocument.RemoteDocument = &wordepress.Document{}
				} else {
					log.Printf("Uploading document: %s", localDocument.Slug)
					remoteDocument, err := wordepress.PostDocument(user, password, endpoint, localDocument)
					if err != nil {
						log.Fatalf("Error uploading document: %v", err)
					}
					localDocument.RemoteDocument = remoteDocument
				}
			}
		}

		// Upload new images
		for _, image := range images {
			exists, err := headImage(image)
			if err != nil {
				log.Fatalf("Error testing image existence: %v", err)
			}
			if exists {
				if dryRun {
					log.Printf("Would skip image: %s%s", image.Hash, image.Extension)
				} else {
					log.Printf("Skipping image: %s%s", image.Hash, image.Extension)
				}
				continue
			}
			if dryRun {
				log.Printf("Would upload image: %s%s", image.Hash, image.Extension)
			} else {
				log.Printf("Uploading image: %s%s", image.Hash, image.Extension)
				err = postImage(image)
				if err != nil {
					log.Fatalf("Error uploading image: %v", err)
				}
			}
		}

		// Remove residual remote documents
		for _, remoteDocument := range existing {
			if remoteDocument.Product != product || remoteDocument.Version != version {
				// meta_query filter was ignored, most likely due to wrong plugin version
				log.Printf("Skipping delete of %s due to product/version mismatch. "+
					"Is your plugin up to date?", remoteDocument.Slug)
				continue
			}
			if dryRun {
				log.Printf("Would delete document: %s", remoteDocument.Slug)
			} else {
				log.Printf("Deleting document: %s", remoteDocument.Slug)
				err := wordepress.DeleteDocument(user, password, endpoint, remoteDocument)
				if err != nil {
					log.Fatalf("Error deleting document: %v", err)
				}
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(publishCmd)
}
