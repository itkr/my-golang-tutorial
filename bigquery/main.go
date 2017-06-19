package main

import (
	"encoding/csv"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/bigquery/v2"
	"log"
	"os"
	"strconv"
	"time"
	"reflect"
)

const tokenURL = "https://accounts.google.com/o/oauth2/token"

const queryBase = `SELECT
  DATE(UTC_USEC_TO_MONTH(event_time)) AS segment, os_name, COUNT(DISTINCT app_user_id) AS user_count, '%s' AS game_id
FROM
  TABLE_DATE_RANGE([%s:%s.%s], TIMESTAMP('%s'), TIMESTAMP('%s'))
WHERE
  event_time >= TIMESTAMP('%s')
  AND event_time < TIMESTAMP('%s')
GROUP BY
  segment, os_name, game_id
ORDER BY
  segment, os_name`

const segmentLayout = "2006-01-02"

type QueryResult struct {
	Segment   time.Time
	OSName    string
	UserCount int
	GameID    string
}

func (queryResult QueryResult) CSVHeader() []string {
	rt := reflect.TypeOf(queryResult)
	fields := make([]string, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		fields[i] = rt.Field(i).Name
	}
	return fields
}

func (queryResult *QueryResult) ToCSVRaw() []string {
	return []string{
		queryResult.Segment.Format(segmentLayout),
		queryResult.OSName,
		strconv.Itoa(queryResult.UserCount),
		queryResult.GameID,
	}
}

func writeToCSV(queryResults []*QueryResult, dstFilePath string) error {
	file, err := os.OpenFile(dstFilePath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	if len(queryResults) != 0 {
		writer.Write(queryResults[0].CSVHeader())
		for _, queryResult := range queryResults {
			writer.Write(queryResult.ToCSVRaw())
		}
	}
	writer.Flush()
	return nil
}

func query(service *bigquery.Service, projectID string, sql string) ([]*QueryResult, error) {
	queryResponse, err := service.Jobs.Query(projectID, &bigquery.QueryRequest{
		Query:     sql,
		TimeoutMs: 60000,
	}).Do()
	queryResults := make([]*QueryResult, len(queryResponse.Rows))
	for i, row := range queryResponse.Rows {
		segment, err := time.Parse(segmentLayout, row.F[0].V.(string))
		if err != nil {
			continue
		}
		userCount, err := strconv.Atoi(row.F[2].V.(string))
		if err != nil {
			continue
		}
		queryResults[i] = &QueryResult{
			Segment:   segment,
			OSName:    row.F[1].V.(string),
			UserCount: userCount,
			GameID:    row.F[3].V.(string),
		}
	}
	return queryResults, err
}

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

func getBigQueryService(email string, keyFilePath string) (*bigquery.Service, error) {
	key, err := getKey(keyFilePath, 2048)
	if err != nil {
		return nil, err
	}
	cfg := jwt.Config{
		Email:      email,
		PrivateKey: key,
		Scopes:     []string{bigquery.BigqueryScope},
		TokenURL:   tokenURL,
	}
	ctx := context.Background()
	client := cfg.Client(ctx)
	return bigquery.New(client)
}

func getFirstDayOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func getLastDayOfMonth(t time.Time) time.Time {
	return getFirstDayOfMonth(t).AddDate(0, 1, -1)
}

func makeQuery(projectID string, t time.Time, appID string) string {
	dataSet := "dummy"
        tableNameBase := "dummy"
        timeStampLayout := "2006-01-02"

	from := getFirstDayOfMonth(t).Format(timeStampLayout)
	to := getFirstDayOfMonth(t.AddDate(0, 1, 0)).Format(timeStampLayout)

	return fmt.Sprintf(queryBase, appID, projectID, dataSet, tableNameBase, from, to, from, to)
}

func getAggregationFromBigQuery(config Config, t time.Time) ([]*QueryResult, error) {
	service, err := getBigQueryService(config.Email, config.FilePath)
	if err != nil {
		return nil, err
	}
	sql := makeQuery(config.ProjectID, t, config.AppID)
	rows, err := query(service, config.ProjectID, sql)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func main() {
	appID := "dummy"
	config := getConfig()[appID]
	lastMonth := time.Now().AddDate(0, -1, 0)
	rows, err := getAggregationFromBigQuery(config, lastMonth)
	writeToCSV(rows, "/tmp/mau.csv")
	if err != nil {
		log.Fatalf("%v", err)
		fmt.Printf("%v", err)
	}
}
