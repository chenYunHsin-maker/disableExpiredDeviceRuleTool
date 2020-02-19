package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/syhlion/sqlwrapper"
)

const (
	dbName_default          = "cubs"
	mysqlDomain_default     = "127.0.0.1:3308"
	apiserverDomain_default = "http://127.0.0.1:8080"
	username_default        = "root"
	password_default        = "root"
	timeFormat              = "2006-01-02"
	detailTime              = "2006-01-02_150405"
)

type Config struct {
	FileNms []struct {
		Config	struct {
			ApiserverDomain	string	`json:"apiserverDomain,omitempty"`
			FromDate	string	`json:"from_date,omitempty"`
			MysqlDomain	string	`json:"mysqlDomain,omitempty"`
			Password	string	`json:"password,omitempty"`
			ToDate		string	`json:"to_date,omitempty"`
			Username	string	`json:"username,omitempty"`
		}	`json:"config"`
		FileNm	string	`json:"fileNm"`
	} `json:"FileNms"
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
func main() {
	crontabFileNm := "./crontabFile.txt"
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
