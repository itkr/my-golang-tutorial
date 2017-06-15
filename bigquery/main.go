package main

import (
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/bigquery/v2"
	"log"
	"os"
)

const tokenURL = "https://accounts.google.com/o/oauth2/token"

func insrtQuery(service *bigquery.Service, projectID string, sql string) (string, error) {
	jobInsertCall := service.Jobs.Insert(projectID, &bigquery.Job{
		Configuration: &bigquery.JobConfiguration{
			Query: &bigquery.JobConfigurationQuery{
				Query: sql,
			},
		},
	})
	job, err := jobInsertCall.Do()
	return job.JobReference.JobId, err
}

// BigQueryから検索jobの結果を取得する
func getQueryResult(service *bigquery.Service, projectID string, jobId string) (*bigquery.GetQueryResultsResponse, error) {
	responseCall := service.Jobs.GetQueryResults(projectID, jobId)
	response, err := responseCall.Do()
	if err != nil {
		return nil, err
	}
	return response, nil
}

func getKey(filePath string, bufferSize int) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return []byte{}, err
	}
	defer file.Close()
	key := make([]byte, bufferSize)
	file.Read(key)
	return key, err
}

func getBigQueryService(config Config) (*bigquery.Service, error) {
	key, err := getKey(config.FilePath, 2048)
	if err != nil {
		return nil, err
	}
	cfg := jwt.Config{
		Email:      config.Email,
		PrivateKey: key,
		Scopes:     []string{bigquery.BigqueryScope},
		TokenURL:   tokenURL,
	}
	ctx := context.Background()
	client := cfg.Client(ctx)
	return bigquery.New(client)
}

func main() {
	projectID := "project-9999"
	configures := getConfig()

	service, err := getBigQueryService(configures[projectID])
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	query := "SELECT COUNT(time) FROM [hoge] LIMIT 1"
	jobID, err := insrtQuery(service, projectID, query)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	fmt.Println(jobID)
}
