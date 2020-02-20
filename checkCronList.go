package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"net/http"
	"os/exec"

	"fmt"
	"io/ioutil"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
	_ "github.com/syhlion/sqlwrapper"
)

const (
	dbName_default          = "cubs"
	mysqlDomain_default     = "sdwan-orch-db-orchestrator-db:3306"
	apiserverDomain_default = "http://sdwan-api-01-apiserver:80"
	username_default        = "root"
	password_default        = "root"
	timeFormat              = "2006-01-02"
	detailTime              = "2006-01-02_150405"
)

var (
	mysqlDomain     string
	apiserverDomain string
	username        string
	password        string
)

type Config struct {
	JobNms []struct {
		Config struct {
			JobNm           string `json:"jobNm"`
			ApiserverDomain string `json:"apiserverDomain,omitempty"`
			FromDate        string `json:"from_date,omitempty"`
			MysqlDomain     string `json:"mysqlDomain,omitempty"`
			Password        string `json:"password,omitempty"`
			ToDate          string `json:"to_date,omitempty"`
			Username        string `json:"username,omitempty"`
		} `json:"config"`
	} `json:"jobNms"`
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func GetTaiwanTime() time.Time {
	loc, _ := time.LoadLocation("Asia/Taipei")
	//fmt.Println(time.Now().In(loc))
	t, _ := ShortDateFromString(time.Now().In(loc).Format(timeFormat))
	return t
}
func ShortDateFromString(ds string) (time.Time, error) {
	t, err := time.Parse(timeFormat, ds)
	if err != nil {
		return t, err
	}
	return t, err
}
func GetTaiwanTime2() time.Time {
	loc, _ := time.LoadLocation("Asia/Taipei")
	//fmt.Println(time.Now().In(loc))
	t, _ := ShortDateFromString2(time.Now().In(loc).Format(detailTime))
	//fmt.Println("t:", t)
	return t
}
func ShortDateFromString2(ds string) (time.Time, error) {
	t, err := time.Parse(detailTime, ds)
	//fmt.Println("s:", t)
	if err != nil {
		return t, err
	}
	return t, err
}
func getCronjobsMap() map[string]string {
	db, err := sql.Open("mysql", username_default+":"+password_default+"@tcp("+mysqlDomain_default+")/"+dbName_default+"?charset=utf8&parseTime=True")
	checkErr(err)
	command := "SELECT  cronCmd,cronjobName,freq FROM cubs.crontab"
	rows, _ := db.Query(command)
	cronjobs := make(map[string]string)
	for rows.Next() {
		var cronCmd sql.NullString
		var cronjobName sql.NullString
		var freq sql.NullString
		if err := rows.Scan(&cronCmd, &cronjobName, &freq); err != nil {
			fmt.Println(" err :", err)

		}
		cronjobs[cronjobName.String] = freq.String + " " + cronCmd.String
	}

	return cronjobs
}
func getCronjobList(cronjobs map[string]string) string {
	var cronjobList string

	dir, err := os.Getwd()
	checkErr(err)
	for key, _ := range cronjobs {
		path := dir + "/" + key + "_log_" + GetTaiwanTime().Format(timeFormat)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.Mkdir(path, os.ModePerm)
		}
		cronjobList += cronjobs[key] + " >> " + path + "/" + GetTaiwanTime2().Format(detailTime) + ".log" + " 2>&1" + "\n"

		switch key {
		case "CheckLicense":
			fmt.Println("Cronjob: Check License")
		}
	}
	return cronjobList
}
func getJsonToStruct() Config {
	var configObj Config
	data, err := ioutil.ReadFile("./conf.json")
	err = json.Unmarshal(data, &configObj)
	checkErr(err)
	return configObj
}
func syncCrontab() {
	crontabFileNm := "./crontabFile.txt"

	var configObj Config
	configObj = getJsonToStruct()

	//fmt.Println(configObj)
	for i := 0; i < len(configObj.JobNms); i++ {
		fmt.Println("cronjob config name:", configObj.JobNms[i].Config.JobNm)
	}
	file, err := os.Create(crontabFileNm)
	checkErr(err)
	cronjobs := getCronjobsMap()
	file.WriteString(getCronjobList(cronjobs))
	file.Close()
	cmd := exec.Command("crontab", "./crontabFile.txt")
	_, err = cmd.Output()
	checkErr(err)
	fmt.Println("crontab added! use \"crontab -e\" and \"grep CRON /var/log/syslog\" to check!")
}
func init() {
	flag.StringVar(&mysqlDomain, "mysqlDomain", "sdwan-orch-db-orchestrator-db:3306", "it's mysql domain")
	flag.StringVar(&username, "username", "root", "mysql login username")
	flag.StringVar(&password, "password", "root", "mysql login password")
	flag.StringVar(&apiserverDomain, "apiserverDomain", "http://sdwan-api-01-apiserver:80", "it's apiserver domain")
}
func main() {
	/*
		cmd := exec.Command("crontab", "-e")
		_, err := cmd.Output()
		checkErr(err)*/
	flag.Parse()
	fmt.Println(apiserverDomain)
	e := echo.New()
	e.GET("/sync", func(c echo.Context) error {
		syncCrontab()

		return c.String(http.StatusOK, "crontab added! use \"crontab -e\" and \"grep CRON /var/log/syslog\" to check!")
	})
	e.Logger.Fatal(e.Start(":1323"))
	/*
		e.Post("/add", func(c echo.Context) error {

		})*/

}
