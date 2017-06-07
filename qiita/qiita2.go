package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const organizationID = "gumi"
const region = "ap-northeast-1"
const bucket = "gumi-qiita"
const tmpFilePath = "/tmp/qiita.csv"
const fileFormat = "qiita-2006-01-02.csv"

// Member : メンバーのデータを扱う
type Member struct {
	Name                   string
	Contributions          int
	Posts                  int
	YesterdayContributions int
	YesterdayPosts         int
	DiffPosts              int
	DiffContributions      int
}

func getS3Client() (*s3.S3, error) {
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
		return nil, err
	}
	client := s3.New(sess)
	return client, nil
}

func getObjectFromS3(keyName string) (*s3.GetObjectOutput, error) {
	client, err := getS3Client()
	if err != nil {
		return nil, err
	}
	return client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(keyName),
	})
}

func putToS3(srcFilePath string, destKeyName string) error {
	file, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	client, err := getS3Client()
	if err != nil {
		return err
	}
	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(destKeyName),
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
		// name
		name := s.Find(".organizationMemberList_userName").Text()
		// contoributions
		contributions := s.Find(".organizationMemberList_memberStats").Last().Text()
		countStr := strings.Split(contributions, " ")[0]
		count, err := strconv.Atoi(countStr)
		if err != nil {
			count = 0
		}
		// posts
		posts := s.Find(".organizationMemberList_memberStats").First().Text()
		postStr := strings.Split(posts, " ")[0]
		post, err := strconv.Atoi(postStr)
		if err != nil {
			count = 0
		}
		// struct
		members = append(members, Member{
			Name:          name,
			Contributions: count,
			Posts:         post,
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

func writeToCSV(members []*Member, dstFilePath string) error {
	file, err := os.OpenFile(dstFilePath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	writer.Write([]string{
		"Name",
		"Contributions",
		"Posts",
		"YesterdayContributions",
		"YesterdayPosts",
		"DiffContributions",
		"DiffPosts",
	})
	for _, member := range members {
		writer.Write([]string{
			member.Name,
			strconv.Itoa(member.Contributions),
			strconv.Itoa(member.Posts),
			strconv.Itoa(member.YesterdayContributions),
			strconv.Itoa(member.YesterdayPosts),
			strconv.Itoa(member.DiffContributions),
			strconv.Itoa(member.DiffPosts),
		})
	}
	writer.Flush()
	return nil
}

func getMembersFromS3(keyName string) ([]Member, error) {
	obj, err := getObjectFromS3(keyName)
	if err != nil {
		return []Member{}, err
	}
	buffer := make([]byte, 1024)
	obj.Body.Read(buffer)
	defer obj.Body.Close()
	reader := csv.NewReader(bytes.NewReader(buffer))

	members := []Member{}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return members, err
		}
		contributions := 0
		if len(record) > 1 {
			contributions, err = strconv.Atoi(record[1])
			if err != nil {
				contributions = 0
			}
		}
		posts := 0
		if len(record) > 2 {
			posts, err = strconv.Atoi(record[2])
			if err != nil {
				posts = 0
			}
		}
		members = append(members, Member{
			Name:          record[0],
			Contributions: contributions,
			Posts:         posts,
		})
	}
	return members, nil
}

func comparisonMembers(oldMembers []Member, newMembers []Member) []*Member {
	oldMap := make(map[string]Member, len(oldMembers))
	for _, member := range oldMembers {
		oldMap[member.Name] = member
	}
	result := make([]*Member, len(newMembers))
	for i, member := range newMembers {
		result[i] = &Member{
			Name:          member.Name,
			Contributions: member.Contributions,
			Posts:         member.Posts,
			YesterdayContributions: oldMap[member.Name].Contributions,
			YesterdayPosts:         oldMap[member.Name].Posts,
			DiffContributions:      member.Contributions - oldMap[member.Name].Contributions,
			DiffPosts:              member.Posts - oldMap[member.Name].Posts,
		}
	}
	return result
}

func main() {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1).Format(fileFormat)

	newMembers, err := getMembers()
	if err != nil {
		fmt.Println(err.Error())
	}

	oldMembers, err := getMembersFromS3(yesterday)
	if err != nil {
		fmt.Println(err.Error())
	}

	diffMembers := comparisonMembers(oldMembers, newMembers)

	// tmpファイルをローカルに
	err = writeToCSV(diffMembers, tmpFilePath)
	if err != nil {
		fmt.Println(err.Error())
	}
	// 日付なしで保存
	_, destKeyName := filepath.Split(tmpFilePath)
	err = putToS3(tmpFilePath, destKeyName)
	if err != nil {
		fmt.Println(err.Error())
	}
	// 日付つきで保存
	destKeyName = time.Now().Format(fileFormat)
	err = putToS3(tmpFilePath, destKeyName)
	if err != nil {
		fmt.Println(err.Error())
	}
	// print
	for _, member := range newMembers {
		fmt.Printf("%s,%d,%d\n", member.Name, member.Contributions, member.Posts)
	}
}
