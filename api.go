package wordepress

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

func GetJSONDocuments(user, password, endpoint, query string) ([]DocumentJSON, error) {
	var jsonDocuments []DocumentJSON
	maxPage := 1
	for page := 1; page <= maxPage; page++ {
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

		maxPage, err = strconv.Atoi(response.Header.Get("X-WP-TotalPages"))
		if err != nil {
			return nil, err
		}

		var jsonPage []DocumentJSON
		err = json.Unmarshal(responseBytes, &jsonPage)
		if err != nil {
			return nil, err
		}

		jsonDocuments = append(jsonDocuments, jsonPage...)
	}
	return jsonDocuments, nil
}

func DeleteJSONDocuments(user, password, endpoint string, jsonDocuments []DocumentJSON) error {
	for _, jsonDocument := range jsonDocuments {
		url := fmt.Sprintf("%s/%d?force=true", endpoint, jsonDocument.Id)
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
	}
	return nil
}
