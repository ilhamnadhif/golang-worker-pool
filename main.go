package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
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

const (
	host           = "localhost"
	port           = 5432
	user           = "root"
	password       = "root"
	dbname         = "worker_pool"
	dbMaxIdleConns = 4
	dbMaxConns     = 100
	totalWorker    = 100
)

func OpenDbConnection() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	log.Println("=> open db connection")

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(dbMaxConns)
	db.SetMaxIdleConns(dbMaxIdleConns)

	return db, nil
}

func ReadFile(datas chan Data) {
	defer close(datas)
	csvFile, err := os.Open("data.csv")
	LogFatalIfErr(err)
	defer csvFile.Close()

	csvLines, err := csv.NewReader(csvFile).ReadAll()
	LogFatalIfErr(err)

	counter := 0
	for _, line := range csvLines {
		datas <- Data{
			ID:     line[0],
			Random: line[1],
		}
		counter++
	}
	fmt.Println(fmt.Sprintf("===== total data berjumlah %d data", counter))
}

func Worker(stmt *sql.Stmt, wg *sync.WaitGroup, data chan Data) {
	for v := range data {
		for {
			_, err := stmt.Exec(v.ID, v.Random)
			if err != nil {
				errPq, ok := err.(*pq.Error)
				if ok {
					if errPq.Code == "53300" {
						log.Println(errPq.Code, errPq.Message, err.Error())
						continue
					} else {
						logrus.Error(err.Error())
						break
					}
				} else {
					break
				}
			} else {
				break
			}
		}
	}
	wg.Done()
}

func main() {
	startTime := time.Now()

	file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	logrus.SetReportCaller(true)
	logrus.SetOutput(file)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	db, err := OpenDbConnection()
	LogFatalIfErr(err)
	defer db.Close()

	datas := make(chan Data)
	wg := new(sync.WaitGroup)

	go ReadFile(datas)
	prepare, err := db.Prepare("insert into data values ($1, $2)")
	if err != nil {
		fmt.Println("error db prepare", err)
	}

	for i := 0; i < totalWorker; i++ {
		wg.Add(1)
		go Worker(prepare, wg, datas)
	}
	wg.Wait()

	duration := time.Since(startTime)
	fmt.Println("done in", int(math.Ceil(duration.Seconds())), "seconds")
}
