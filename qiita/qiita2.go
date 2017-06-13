package main

import (
	"fmt"
	"path/filepath"
	"time"
)


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
	err = copyToS3(tmpFilePath, destKeyName)
	if err != nil {
		fmt.Println(err.Error())
	}
	// 日付つきで保存
	destKeyName = time.Now().Format(fileFormat)
	err = copyToS3(tmpFilePath, destKeyName)
	if err != nil {
		fmt.Println(err.Error())
	}
	// print
	for _, member := range newMembers {
		fmt.Printf("%s,%d,%d\n", member.Name, member.Contributions, member.Posts)
	}
}
