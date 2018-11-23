package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/olivere/elastic.v5"
)

const kibanaConfigFile = "kibana.json"

func main() {
	// Setup dashboards in Kibana (not required step)
	if err := setupDashboards(); err != nil {
		fmt.Printf("failed to setup Kibana dashboards, error: %s\n", err.Error())
	}

	// Starting with elastic.v5, you must pass a context to execute each service
	ctx := context.Background()

	url := os.Getenv("ELASTIC_URL")
	if len(url) == 0 {
		panic(fmt.Sprintf("wrong Elastic url: %s", url))
	}
	// Obtain a client and connect to the default Elasticsearch installation
	client, err := elastic.NewClient(elastic.SetURL(url))
	if err != nil {
		panic(err)
	}

	// Ping the Elasticsearch server to get e.g. the version number
	info, code, err := client.Ping(url).Do(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

	// Getting the ES version number is quite common, so there's a shortcut
	esVersion, err := client.ElasticsearchVersion(url)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Elasticsearch version %s\n", esVersion)

	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists("random").Do(ctx)
	if err != nil {
		panic(err)
	}
	if !exists {
		// Create a new index.
		_, err := client.CreateIndex("random").BodyString(randomMapping).Do(ctx)
		if err != nil {
			panic(err)
		}
	}

	names := []string{"Pavel", "John", "Mark", "Patrick", "Rex", "Julia"}
	// Index a random (using JSON serialization)
	i := 0
	fmt.Println("Start sending data")
	for i <= 10000 {
		i++
		err = sendRandomData(client, ctx, i, names)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Data successfully sent")

	// Flush to make sure the documents got written.
	_, err = client.Flush().Index("random").Do(ctx)
	if err != nil {
		panic(err)
	}

	// Search with a term query
	searchByUser := "Pavel"
	termQuery := elastic.NewTermQuery("creator", searchByUser)
	searchResult, err := client.Search().
		Index("random").       // search in index "random"
		Query(termQuery).      // specify the query
		Sort("created", true). // sort by "created" field, ascending
		From(0).Size(15).      // take documents 0-9
		Pretty(true).          // pretty print request and response JSON
		Do(ctx)                // execute
	if err != nil {
		panic(err)
	}

	// searchResult is of type SearchResult and returns hits, suggestions,
	// and all kinds of other information from Elasticsearch
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

	// Here's how you iterate through results with full control over each step.
	if searchResult.Hits.TotalHits > 0 {
		fmt.Printf("Found a total of %d random numbers from %s\n", searchResult.Hits.TotalHits, searchByUser)

		// Iterate through results
		for _, hit := range searchResult.Hits.Hits {
			// Deserialize hit.Source into a NormalDistribution (could also be just a map[string]interface{}).
			var r normalDistribution
			err := json.Unmarshal(*hit.Source, &r)
			if err != nil {
				panic(err)
			}

			fmt.Printf("Random number %f created at %s by %s\n", r.RandomNumber, r.Created.String(), r.Creator)
		}
	} else {
		fmt.Print("Found no numbers\n")
	}

	// Continue sending data
	fmt.Println("Launch endless data sending every second...")
	for {
		i++
		time.Sleep(time.Second)
		err := sendRandomData(client, ctx, i, names)
		if err != nil {
			panic(err)
		}
	}
}

const randomMapping = `
{
	"settings":{
		"number_of_shards": 1,
		"number_of_replicas": 0
	},
	"mappings":{
		"rand":{
			"properties":{
				"created":{
					"type":"date"
				},
				"creator":{
					"type":"keyword"
				}
			}
		}
	}
}`

func sendRandomData(client *elastic.Client, ctx context.Context, i int, names []string) (err error) {
	_, err = client.Index().
		Index("random").
		Type("rand").
		Id(strconv.Itoa(i)).
		BodyJson(normalDistribution{
			Created:      time.Now(),
			RandomNumber: rand.NormFloat64(),
			Creator:      names[rand.Intn(len(names))],
		}).
		Do(ctx)
	return
}

// Random is a structure used for serializing/deserializing data in Elasticsearch
type normalDistribution struct {
	Created      time.Time `json:"created"`
	Creator      string    `json:"creator"`
	RandomNumber float64   `json:"randomNumber"`
}

// setupDashboards put graphs and dashboards inside kibana
func setupDashboards() error {
	f, err := os.Open(kibanaConfigFile)
	if err != nil {
		return err
	}
	defer f.Close()
	url := os.Getenv("KIBANA_URL") + "/api/kibana/dashboards/import"
	req, err := http.NewRequest("POST", url, f)
	if err != nil {
		return err
	}

	req.Header.Add("Kbn-Xsrf", "true")
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("error responce from kibana: %s", string(body))
	}
	return nil
}
