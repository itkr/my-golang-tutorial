package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"io"
)

func main(){
	var fp *os.File
	var err error

	// ファイル読み込み
	fp, err = os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	reader := csv.NewReader(fp)
// 	// TSV
// 	reader.Comma = '\t'
	reader.LazyQuotes = true
	
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		fmt.Println(record)
	}
}









