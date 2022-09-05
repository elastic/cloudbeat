// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

// https://kb.objectrocket.com/elasticsearch/how-to-get-elasticsearch-documents-using-golang-448
package console_actions

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

type ActionDoc struct {
	Resource string `json:"resource,omitempty"`
	Command  string `json:"command,omitempty"`
	Mode     int    `json:"mode,omitempty"`
}

func getConfig() elasticsearch.Config {
	cfg := elasticsearch.Config{
		Addresses: []string{
			//"http://localhost:9200",
			"http://host.docker.internal:9200",
		},
		Username: "elastic",
		Password: "changeme",
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second,
			//DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	return cfg
}

func getQuery() string {
	var query = `
			{
"query": {
"bool": {
"filter": {"range": {"timestamp": {"gte": "now-10s", "lte": "now" }}
}}
},
"sort": [
  {
    "timestamp": {
      "order": "desc"
    }
  }
],
    "size": 1
  }
			`
	return query
}

func runEsSearch(client *elasticsearch.Client, ctx context.Context, read *strings.Reader) (*esapi.Response, error) {
	res, err := client.Search(
		client.Search.WithContext(ctx),
		client.Search.WithIndex("csp-actions-1"),
		client.Search.WithBody(read),
		client.Search.WithTrackTotalHits(true),
		client.Search.WithPretty(),
	)
	return res, err
}

func changeResourceMode(resource string, mode int) {
	err := os.Chmod(resource, os.FileMode(mode))
	if err != nil {
		log.Fatal(err)
	}
}

func RunActionsRoutine() {
	for {
		newAction, _ := FetchNewActions()
		if newAction.Command != "" {
			fmt.Println("exec command")
			mode := newAction.Mode
			//resource := "/hostfs/etc/kubernetes/manifests/etcd.yaml"
			resource := newAction.Resource
			changeResourceMode(resource, mode)
		}
		time.Sleep(5 * time.Second)
	}
}

func FetchNewActions() (ActionDoc, error) {
	ctx := context.Background()

	cfg := getConfig()
	client, err := elasticsearch.NewClient(cfg)

	// Exit the system if connection raises an error
	if err != nil {
		fmt.Println("Elasticsearch connection error:", err)
	}

	// Instantiate a mapping interface for API response
	var mapResp map[string]interface{}
	var buf bytes.Buffer

	query := getQuery()

	//Build and read the query string for the Elasticsearch Search() method
	var b strings.Builder
	b.WriteString(query)
	read := strings.NewReader(b.String())

	// Attempt to encode the JSON query and look for errors
	if err := json.NewEncoder(&buf).Encode(read); err != nil {
		log.Fatalf("Error encoding query: %s", err)

		// Query is a valid JSON object
	} else {
		fmt.Println("json.NewEncoder encoded query:", read)

		res, err := runEsSearch(client, ctx, read)
		// Check for any errors returned by API call to Elasticsearch
		if err != nil {
			fmt.Println("Elasticsearch Search() API ERROR:", err)

		} else {
			fmt.Println("############################")
			fmt.Println(res)
			fmt.Println("############################")
			// Close the result body when the function call is complete
			defer res.Body.Close()
			// Decode the JSON response and using a pointer
			if err := json.NewDecoder(res.Body).Decode(&mapResp); err != nil {
				log.Fatalf("Error parsing the response body: %s", err)

				// If no error, then convert response to a map[string]interface
			} else {
				fmt.Println("mapResp TYPE:", reflect.TypeOf(mapResp))
				for _, hit := range mapResp["hits"].(map[string]interface{})["hits"].([]interface{}) {

					// Parse the attributes/fields of the document
					doc := hit.(map[string]interface{})
					var source = doc["_source"]

					jsonbody, _ := json.Marshal(source)
					newAction := ActionDoc{}
					json.Unmarshal(jsonbody, &newAction)

					return newAction, nil

				}
			}

		}

	}
	return ActionDoc{}, err
}
