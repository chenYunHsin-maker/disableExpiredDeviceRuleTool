package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/syhlion/sqlwrapper"
)

const (
	dbName_default          = "cubs"
	mysqlDomain_default     = "127.0.0.1:3308"
	apiserverDomain_default = "http://127.0.0.1:8080"
	username_default        = "root"
	password_default        = "root"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	crontabFileNm := "./crontabFile.txt"
	file, err := os.Create(crontabFileNm)
	checkErr(err)
	db, err := sql.Open("mysql", username_default+":"+password_default+"@tcp("+mysqlDomain_default+")/"+dbName_default+"?charset=utf8&parseTime=True")
	checkErr(err)
	command := "SELECT  cronCmd,cronjobName,freq FROM cubs.crontab"
	rows, _ := db.Query(command)
	var cronjobList string
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
	for key, _ := range cronjobs {
		cronjobList += cronjobs[key] + "\n"
	}
	file.WriteString(cronjobList)
	file.Close()

	cmd := exec.Command("crontab", "./crontabFile.txt")
	checkErr(err)
	stdout, err := cmd.Output()
	checkErr(err)
	fmt.Println("crontab added! use \"crontab -e\" or \"grep CRON /var/log/syslog\" to check!")
}
