package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Item struct {
	Rendered_body string `json:"rendered_body"`
	Body          string `json:"body"`
	Coediting     bool   `json:"coediting"`
	Created_at    string `json:"created_at"`
	Group         string `json:"group"`
	Id            string `json:"id"`
	Private       bool   `json:"private"`
	Tags          []Tag  `json:"tag"`
	Title         string `json:"title"`
	Updated_at    string `json:"updated_at"`
	Url           string `json:"url"`
	User          User   `json:"user"`
}

type Tag struct {
	Name     string   `json:"name"`
	Versions []string `json:versions`
}

type User struct {
	Description         string `json:"description"`
	Facebook_id         string `json:"facebook_id"`
	Followees_count     int    `json:"followees_count"`
	Followers_count     int    `json:"followers_count"`
	Github_login_name   string `json:"github_login_name"`
	Id                  string `json:"id"`
	Items_count         int    `json:"items_count"`
	Linkedin_id         string `json:"linkedin_id"`
	Location            string `json:"location"`
	Name                string `json:"name"`
	Organization        string `json:"organization"`
	Permanent_id        int    `json:"permanent_id"`
	Profile_image_url   string `json:"profile_image_url"`
	Twitter_screen_name string `json:"twitter_screen_name"`
	Website_url         string `json:"website_url"`
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

func getIdList(userId string) ([]string, error) {
	var items []Item
	err := getItems(userId, &items)
	if err != nil {
		return nil, err
	}
	urlList := make([]string, len(items))
	for i, item := range items {
		urlList[i] = item.Id
	}
	return urlList, nil
}

func getLikeCount(userId string, itemId string) (int, error) {
	url := fmt.Sprintf("http://qiita.com/%s/items/%s", userId, itemId)
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return 0, err
	}
	count := doc.Find(".js-likecount").Text()
	return strconv.Atoi(count)
}

func getSumLikeCount(userId string) (int, error) {
	var count int
	urlList, err := getIdList(userId)
	if err != nil {
		return 0, err
	}
	for _, url := range urlList {
		like, err := getLikeCount(userId, url)
		if err != nil {
			return count, err
		}
		count += like
	}
	return count, nil
}

func main() {
	count, err := getSumLikeCount("itkr")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(count)
}
