package main

import (
	"encoding/csv"
	"fmt"
	"path/filepath"
	"os"
	"strconv"
	"strings"
	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const organizationID = "gumi"
const region = "ap-northeast-1"
const bucket = "gumi-qiita"
const tmpFilePath = "/tmp/qiita.csv"

// Member : メンバーのデータを扱う
type Member struct {
	Name          string
	Contributions int
}

func putToS3(srcFilePath string) error {
	file, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 環境変数で AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY を指定する
	// accessKeyを直接指定も可能（publicリポジトリにコミットしないこと）
	// creds := credentials.NewStaticCredentials(accessKey, sercetAccessKey, "")
	creds := credentials.NewEnvCredentials()
	config := aws.Config{
		Credentials: creds,
		Region:      aws.String(region),
	}
	sess, err := session.NewSession(&config)
	if err != nil {
		return err
	}
	client := s3.New(sess)
	_, fileName := filepath.Split(srcFilePath)
	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
		Body:   file,
	})
	return err
}

func getMembersInAPage(organizationID string, page int) ([]Member, error) {
	var members []Member
	url := fmt.Sprintf("https://qiita.com/organizations/%s/members?page=%d", organizationID, page)
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return members, err
	}
	doc.Find(".organizationMemberList_memberProfile").Each(func(i int, s *goquery.Selection) {
		name := s.Find(".organizationMemberList_userName").Text()
		contributions := s.Find(".organizationMemberList_memberStats").Last().Text()
		countStr := strings.Split(contributions, " ")[0]
		count, err := strconv.Atoi(countStr)
		if err != nil {
			count = 0
		}
		members = append(members, Member{
			Name:          name,
			Contributions: count,
		})
	})
	return members, nil
}

func getMembers() ([]Member, error) {
	var allMembers []Member
	var maxMemberCountPerPage int
	for page := 1; ; page++ {
		membersInAPage, err := getMembersInAPage(organizationID, page)
		if err != nil {
			return allMembers, err
		}
		if page == 1 {
			maxMemberCountPerPage = len(membersInAPage)
		}
		if len(membersInAPage) == 0 {
			break
		}
		allMembers = append(allMembers, membersInAPage...)
		if len(membersInAPage) < maxMemberCountPerPage {
			break
		}
	}
	return allMembers, nil
}

func writeToCSV(members []Member, dstFilePath string) error {
	file, err := os.OpenFile(dstFilePath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	for _, member := range members {
		writer.Write([]string{member.Name, strconv.Itoa(member.Contributions)})
	}
	writer.Flush()
	return nil
}

func main() {
	members, err := getMembers()
	if err != nil {
		fmt.Println(err.Error())
	}
	for _, member := range members {
		fmt.Printf("%s: %d\n", member.Name, member.Contributions)
	}
	err = writeToCSV(members, tmpFilePath)
	if err != nil {
		fmt.Println(err.Error())
	}
	err = putToS3(tmpFilePath)
	if err != nil {
		fmt.Println(err.Error())
	}
}
