package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"
)

func LogFatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Data struct {
	ID     string
	Random string
}

func ReadFile(datas chan []Data) {
	csvFile, err := os.Open("data.csv")
	LogFatalIfErr(err)
	defer csvFile.Close()

	csvLines, err := csv.NewReader(csvFile).ReadAll()
	LogFatalIfErr(err)

	res := make([]Data, 0)
	counter := 0
	for _, line := range csvLines {
		res = append(res, Data{
			ID:     line[0],
			Random: line[1],
		})
		counter++
	}
	fmt.Println(fmt.Sprintf("===== total data berjumlah %d data", counter))
	datas <- res
}

func main() {
	startTime := time.Now()
	datas := make(chan []Data)
	go ReadFile(datas)
	<-datas
	fmt.Println(time.Since(startTime))
}
