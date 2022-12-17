package main

import (
	"encoding/csv"
	"fmt"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

func GenerateLine() []string {
	return []string{uuid.NewString(), strconv.Itoa(rand.Int())}
}

func main() {
	timeStart := time.Now()
	csvFile, err := os.Create("data.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer csvFile.Close()

	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()

	wg := new(sync.WaitGroup)
	mutex := sync.Mutex{}
	for i := 1; i <= 1000; i++ {
		wg.Add(1)
		go func() {
			for j := 1; j <= 1000; j++ {
				mutex.Lock()
				err2 := csvwriter.Write(GenerateLine())
				if err2 != nil {
					log.Fatal(err2)
				}
				mutex.Unlock()
			}
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Println(time.Since(timeStart).Seconds())
}
