package elasticsearch_test

import (
	"context"
	"log"

	"github.com/disaster37/elasticsearch/v9"
	"github.com/sirupsen/logrus"
)

// ExampleNew demonstrates creating a client and pinging a cluster.
func ExampleNew() {
	logger := logrus.NewEntry(logrus.StandardLogger())
	client, err := elasticsearch.New(&elasticsearch.Config{
		URL:           "https://localhost:9200",
		Username:      "elastic",
		Password:      "changeme",
		TLSSkipVerify: true,
	}, logger)
	if err != nil {
		log.Fatal(err)
	}

	ok, err := client.Info().Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("cluster reachable: %v", ok)
}

// ExampleNew_apiKey demonstrates API-key authentication.
func ExampleNew_apiKey() {
	logger := logrus.NewEntry(logrus.StandardLogger())
	client, err := elasticsearch.New(&elasticsearch.Config{
		URL:    "https://localhost:9200",
		APIKey: "ZpGaRk1BQlJhc0hKQ3h6QjB4Z0I6cXhLZjJ3U0tqYzVfXzZJNjJ3Zw==",
	}, logger)
	if err != nil {
		log.Fatal(err)
	}
	_ = client
}
