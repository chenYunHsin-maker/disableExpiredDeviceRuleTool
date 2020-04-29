package main

import (
	"database/sql"
	"flag"
	"net/http"
	"os/exec"
	"strings"

	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
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

type Cronjob struct {
	Name string `json:"name" form:"name" query:"name"`
	Freq string `json:"freq" form:"freq" query:"freq"`
	Cmd  string `json:"cmd" form:"cmd" query:"cmd"`
}

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
	t, _ := ShortDateFromString2(time.Now().In(loc).Format(detailTime))
	return t
}
func ShortDateFromString2(ds string) (time.Time, error) {
	t, err := time.Parse(detailTime, ds)
	if err != nil {
		return t, err
	}
	return t, err
}
func getCronjobsMap() map[string]string {
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+"cubs"+"?charset=utf8&parseTime=True")
	checkErr(err)
	command := "SELECT  cronCmd,cronjobName,freq FROM cubs.crontab"
	rows, _ := db.Query(command)
	cronjobs := make(map[string]string)
	for rows.Next() {
		var cronCmd sql.NullString
		var cronjobName sql.NullString
		var freq sql.NullString
		if err := rows.Scan(&cronCmd, &cronjobName, &freq); err != nil {
			glog.Infoln(err.Error())
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
			glog.Infoln("Cronjob: Check License")
		}
	}
	return cronjobList
}

func syncCrontab() {
	crontabFileNm := "./crontabFile.txt"

	file, err := os.Create(crontabFileNm)
	checkErr(err)
	cronjobs := getCronjobsMap()
	file.WriteString(getCronjobList(cronjobs))
	file.Close()
	cmd := exec.Command("crontab", "./crontabFile.txt")
	_, err = cmd.Output()
	checkErr(err)
	glog.Infoln("crontab added! use \"crontab -e\" and \"grep CRON /var/log/syslog\" to check!")
}
func init() {
	flag.StringVar(&mysqlDomain, "mysqlDomain", "sdwan-orch-db-orchestrator-db:3306", "it's mysql domain")
	flag.StringVar(&username, "username", "root", "mysql login username")
	flag.StringVar(&password, "password", "root", "mysql login password")
	flag.StringVar(&apiserverDomain, "apiserverDomain", "http://sdwan-api-01-apiserver:80", "it's apiserver domain")
}
func createCronjob(job Cronjob) {
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+"cubs?charset=utf8&parseTime=True")
	cmd := "INSERT INTO cubs.crontab(cronjobName,cronCmd,freq) VALUES (?,?,?);"
	_, err = db.Exec(cmd, job.Name, job.Cmd, job.Freq)
	checkErr(err)
	syncCrontab()
}
func readCronjob(name string) string {
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+"cubs?charset=utf8&parseTime=True")
	checkErr(err)
	command := "SELECT cronjobName,cronCmd,freq FROM cubs.crontab WHERE cronjobName='" + name + "';"
	rows, _ := db.Query(command)
	var resultStr string
	var count int64

	checkErr(err)
	for rows.Next() {
		err = rows.Scan(&count)
		if count == 0 {
			resultStr = "please verify if your cronjobName: " + name + " is exist"
		}
		var cronjobName sql.NullString
		var cronjobCmd sql.NullString
		var freq sql.NullString
		if err := rows.Scan(&cronjobName, &cronjobCmd, &freq); err != nil {
			glog.Infoln(err.Error())
		}
		if cronjobName.Valid {
			resultStr = "cronjobName: " + cronjobName.String + " cronjobCmd:" + cronjobCmd.String + " freq:" + freq.String
		}
	}

	rows.Close()
	return resultStr

}
func createTable() {
	dbName := "cubs"
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName+"?charset=utf8&parseTime=True")
	stmt, err := db.Prepare("CREATE Table crontab(cronjogId int NOT NULL AUTO_INCREMENT  PRIMARY KEY, cronjobName varchar(50) NOT NULL UNIQUE, cronCmd varchar(200), freq varchar(50));")
	if err != nil {
		glog.Infoln(err.Error())
	}
	_, err = stmt.Exec()
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			glog.Infoln("Crontab table created successfully :D")
		} else {
			glog.Infoln(err.Error())
		}
	} else {
		glog.Infoln("Crontab table created successfully :D")
	}
}
func updateCronjob(job Cronjob) {

	//UPDATE cubs.crontab SET `cronCmd` = './maomao',`freq`='1 * * * *'  WHERE cronjobName= 'vicky';
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+"cubs?charset=utf8&parseTime=True")
	cmd := "UPDATE cubs.crontab SET `cronCmd` = '" + job.Cmd + "',`freq`='" + job.Freq + "' " + "WHERE cronjobName= '" + job.Name + "';"
	glog.Infoln(cmd)
	_, err = db.Exec(cmd)
	checkErr(err)
	syncCrontab()
}
func deleteCronjob(name string) string {
	var resultStr string
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+"cubs?charset=utf8&parseTime=True")
	checkErr(err)
	command := "DELETE FROM cubs.crontab WHERE cronjobName='" + name + "';"
	res, err := db.Exec(command)
	if err == nil {
		cnt, err := res.RowsAffected()
		if err == nil {
			switch cnt {
			case 0:
				resultStr = "please verify if " + name + " is exist"
			case 1:
				resultStr = "delete " + name + " successfully!"
				syncCrontab()
			}
		}
	}
	checkErr(err)

	return resultStr

}
func main() {

	flag.Parse()

	//glog.Infoln()
	createTable()
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.POST("/createCronjob", func(c echo.Context) error {
		job := new(Cronjob)
		if err := c.Bind(job); err != nil {
			return err
		}
		createCronjob(*job)
		return c.JSON(http.StatusOK, job)
	})
	e.POST("/updateCronjob", func(c echo.Context) error {

		job := new(Cronjob)
		if err := c.Bind(job); err != nil {
			return err
		}
		updateCronjob(*job)
		return c.JSON(http.StatusOK, job)
	})
	e.GET("/readCronjob/:name", func(c echo.Context) error {
		name := c.Param("name")
		resultStr := readCronjob(name)
		return c.String(http.StatusOK, resultStr)
	})
	e.GET("/deleteCronjob/:name", func(c echo.Context) error {
		name := c.Param("name")
		resultStr := deleteCronjob(name)
		return c.String(http.StatusOK, resultStr)
	})
	e.Logger.Fatal(e.Start(":3143"))

}
