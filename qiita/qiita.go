package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Item struct {
	RenderedBody string `json:"rendered_body"`
	Body         string `json:"body"`
	Coediting    bool   `json:"coediting"`
	CreatedAt    string `json:"created_at"`
	Group        string `json:"group"`
	Id           string `json:"id"`
	Private      bool   `json:"private"`
	Tags         []Tag  `json:"tag"`
	Title        string `json:"title"`
	UpdatedAt    string `json:"updated_at"`
	Url          string `json:"url"`
	User         User   `json:"user"`
}

type Tag struct {
	Name     string   `json:"name"`
	Versions []string `json:versions`
}

type User struct {
	Description       string `json:"description"`
	FacebookId        string `json:"facebook_id"`
	FolloweesCount    int    `json:"followees_count"`
	FollowersCount    int    `json:"followers_count"`
	GithubLoginName   string `json:"github_login_name"`
	Id                string `json:"id"`
	ItemsCount        int    `json:"items_count"`
	LinkedinId        string `json:"linkedin_id"`
	Location          string `json:"location"`
	Name              string `json:"name"`
	Organization      string `json:"organization"`
	PermanentId       int    `json:"permanent_id"`
	ProfileImageUrl   string `json:"profile_image_url"`
	TwitterScreenName string `json:"twitter_screen_name"`
	WebsiteUrl        string `json:"website_url"`
}

func strptime(timeStr string) (time.Time, error) {
	layout := "2006-01-02T15:04:05-07:00"
	return time.Parse(layout, timeStr)
}

func getItems(userId string, items *[]Item) error {
	url := fmt.Sprintf("http://qiita.com/api/v2/users/%s/items?per_page=100", userId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	client := new(http.Client)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(byteArray, items); err != nil {
		return err
	}
	return nil
}

func main() {
	var items []Item
	members, err := getMembers()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for _, member := range members {
		err := getItems(member.Name, &items)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(member.Name)
		fmt.Println(items[0].Title)
		fmt.Println(items[0].CreatedAt)
		fmt.Println("======")
	}
}
