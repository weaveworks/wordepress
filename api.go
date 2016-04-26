package wordepress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func PostDocument(user, password, endpoint string, document *Document) (*Document, error) {
	requestBytes, err := json.Marshal(document)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", endpoint, bytes.NewReader(requestBytes))
	request.SetBasicAuth(user, password)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("post failed: %v %s", response.Status, string(responseBytes))
	}

	var remoteDocument Document
	err = json.Unmarshal(responseBytes, &remoteDocument)
	if err != nil {
		return nil, err
	}

	// Ensure server honoured our slug
	if remoteDocument.Slug != document.Slug {
		return nil, fmt.Errorf("duplicate slug: requested %s, response %s",
			document.Slug, remoteDocument.Slug)
	}

	return &remoteDocument, nil
}

func PutDocument(user, password, endpoint string, ID int, document *Document) (*Document, error) {
	requestBytes, err := json.Marshal(document)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%d", endpoint, ID)
	request, err := http.NewRequest("PUT", url, bytes.NewReader(requestBytes))
	request.SetBasicAuth(user, password)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("post failed: %v %s", response.Status, string(responseBytes))
	}

	var remoteDocument Document
	err = json.Unmarshal(responseBytes, &remoteDocument)
	if err != nil {
		return nil, err
	}

	return &remoteDocument, nil
}

func GetDocuments(user, password, endpoint, query string) ([]*Document, error) {
	var jsonDocuments []*Document
	for page := 1; true; page++ {
		url := fmt.Sprintf("%s?%s&page=%d", endpoint, query, page)
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		request.SetBasicAuth(user, password)
		request.Header.Set("Accept", "application/json")

		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			return nil, err
		}

		responseBytes, err := ioutil.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			return nil, err
		}

		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("%v", response.Status)
		}

		var jsonPage []Document
		err = json.Unmarshal(responseBytes, &jsonPage)
		if err != nil {
			return nil, err
		}

		if len(jsonPage) == 0 {
			break
		}

		for i := 0; i < len(jsonPage); i++ {
			log.Printf("Loaded document %s", jsonPage[i].Slug)
			jsonDocuments = append(jsonDocuments, &jsonPage[i])
		}
	}
	return jsonDocuments, nil
}

func DeleteDocument(user, password, endpoint string, jsonDocument *Document) error {
	url := fmt.Sprintf("%s/%d?force=true", endpoint, jsonDocument.ID)
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	request.SetBasicAuth(user, password)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%v", response.Status)
	}
	return nil
}
