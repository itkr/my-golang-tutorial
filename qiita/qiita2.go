package main

import (
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"os"
	"strconv"
	"strings"
)

const organizationId = "gumi"

type Member struct {
	Name          string
	Contributions int
}

func getMembersInAPage(organizationId string, page int) ([]Member, error) {
	var members []Member
	url := fmt.Sprintf("https://qiita.com/organizations/%s/members?page=%d", organizationId, page)
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
		membersInAPage, err := getMembersInAPage(organizationId, page)
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

func writeToCSV(members []Member) error {
	file, err := os.OpenFile("./qiita.csv", os.O_WRONLY|os.O_CREATE, 0600)
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
	err = writeToCSV(members)
	if err != nil {
		fmt.Println(err.Error())
	}
}
