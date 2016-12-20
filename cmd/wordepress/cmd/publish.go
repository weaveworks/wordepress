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

var (
	version string
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

func toMap(rds []*wordepress.CustomPost) map[string]*wordepress.CustomPost {
	rdm := make(map[string]*wordepress.CustomPost)
	for i, _ := range rds {
		rdm[rds[i].Slug] = rds[i]
	}
	return rdm
}

func identical(local *wordepress.CustomPost, remote *wordepress.CustomPost) bool {
	return local.MenuOrder == remote.MenuOrder &&
		local.Title.Raw == remote.Title.Raw &&
		local.Content.Raw == remote.Content.Raw &&
		local.Parent == remote.Parent &&
		local.Version == remote.Version
}

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a site into WordPress",
	Long:  `Publish a site into WordPress`,
	Run: func(cmd *cobra.Command, args []string) {
		if product == "" || tag == "" || version == "" || user == "" || password == "" || len(args) != 1 {
			cmd.UsageFunc()(cmd)
			os.Exit(1)
		}

		// Load local site
		localPosts, images, err := wordepress.ParseDir(product, version, tag, args[0])
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
				"filter[meta_query][1][key]=wpcf-tag&"+
				"filter[meta_query][1][value]=%s", product, tag)

		remotePosts, err := wordepress.Get(user, password, endpoint, query)
		if err != nil {
			log.Fatalf("Unable to get JSON post: %v", err)
		}

		// Create/update posts
		existing := toMap(remotePosts)
		for _, localPost := range localPosts {
			if localPost.LocalParent != nil {
				// Pre-order traversal guarantees the remote post will be set
				localPost.Parent = localPost.LocalParent.RemotePost.ID
			}
			if remotePost, ok := existing[localPost.Slug]; ok {
				if identical(localPost, remotePost) {
					if dryRun {
						log.Printf("Would skip post: %s", localPost.Slug)
					} else {
						log.Printf("Skipping post: %s", localPost.Slug)
					}
				} else {
					if dryRun {
						log.Printf("Would update post: %s", localPost.Slug)
					} else {
						log.Printf("Updating post: %s", localPost.Slug)
						remotePost, err = wordepress.Put(user, password, endpoint, remotePost.ID, localPost)
						if err != nil {
							log.Fatalf("Error updating post: %v", err)
						}
					}
				}
				localPost.RemotePost = remotePost
				delete(existing, localPost.Slug)
			} else {
				if dryRun {
					log.Printf("Would upload post: %s", localPost.Slug)
					localPost.RemotePost = &wordepress.CustomPost{}
				} else {
					log.Printf("Uploading post: %s", localPost.Slug)
					remotePost, err := wordepress.Post(user, password, endpoint, localPost)
					if err != nil {
						log.Fatalf("Error uploading post: %v", err)
					}
					localPost.RemotePost = remotePost
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

		// Remove residual remote posts
		for _, remotePost := range existing {
			if remotePost.Product != product || remotePost.Tag != tag {
				// meta_query filter was ignored, most likely due to wrong plugin version
				log.Printf("Skipping delete of %s due to product/tag mismatch. "+
					"Is your plugin up to date?", remotePost.Slug)
				continue
			}
			if dryRun {
				log.Printf("Would delete post: %s", remotePost.Slug)
			} else {
				log.Printf("Deleting post: %s", remotePost.Slug)
				err := wordepress.Delete(user, password, endpoint, remotePost)
				if err != nil {
					log.Fatalf("Error deleting post: %v", err)
				}
			}
		}
	},
}

func init() {
	publishCmd.Flags().StringVarP(&version, "version", "", "", "Value for custom post version field")
	RootCmd.AddCommand(publishCmd)
}
