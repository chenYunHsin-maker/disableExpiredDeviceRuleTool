package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/syhlion/sqlwrapper"
)

var (
	err01cnt = 0
	err02cnt = 0
	err03cnt = 0
	err04cnt = 0
	err05cnt = 0
	okCnt    = 0
)

const (
	apiserverPodName = "sdwan-api-01-apiserver-6bfc7d5b64-t4ffv"
	timeFormat       = "20060102_15_04_05"
)

type Document struct {
	APIVersion string `json:"apiVersion"`
	Items      []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Spec struct {
			SiteID  int    `json:"siteId"`
			Sn      string `json:"sn,omitempty"`
			Device2 struct {
				Sn string `json:"sn,omitempty"`
			}
		} `json:"spec"`
	} `json:"items"`
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
func getMysqlMap(rows *sql.Rows) map[string][]string {
	snSiteLinkedMap := make(map[string][]string)
	for rows.Next() {
		var site sql.NullString
		var sn sql.NullString
		if err := rows.Scan(&sn, &site); err != nil {
			fmt.Println(" err :", err)
		}
		if site.Valid == true {
			if site.String != "0" {
				snSiteLinkedMap[site.String] = append(snSiteLinkedMap[site.String], sn.String)
			}

			//fmt.Println("sn:", sn.String, " linked site id:", site.String)
		}
	}
	return snSiteLinkedMap
}
func portForwardData() {
	fmt.Println("start to port forward data from api server......")
	cmd := exec.Command("kubectl", "port-forward", apiserverPodName, "8080:8080")
	_ = cmd

	stdout, err := cmd.Output()
	checkErr(err)
	fmt.Println(string(stdout))
	//cmd.Wait()
	time.Sleep(1000 * time.Millisecond)
	fmt.Println("end kubectl connection")
}
func checkTable(snSiteLinkedMap map[string][]string) {
	fmt.Println("start to check map......")
	fmt.Println("your map: ")
	for key, value := range snSiteLinkedMap {
		fmt.Println("Key:", key, "Value:", value)
	}
	fmt.Println("map check end :D")
}
func getApiserverBody(apiserverDomain string) string {
	//fmt.Println("start to get body")
	resp, err := http.Get("http://" + apiserverDomain + "/apis/site/v1alpha1/namespaces/default/siteconfigs")
	checkErr(err)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	s := string(body)
	//fmt.Println(s)
	return s
}
func getApiserverMap(apiserverDomain string) (map[string][]string, map[string][]string) {
	apiServerMap := make(map[string][]string)
	apiServerDevice2Map := make(map[string][]string)
	body := getApiserverBody(apiserverDomain)
	var document Document
	var data []byte = []byte(body)
	json.Unmarshal(data, &document)
	//apiServerMap["S162L45290036"] = append(apiServerMap["S162L45290036"], "88")
	for i := 0; i < len(document.Items); i++ {
		if document.Items[i].Spec.Sn != "" {
			//_, ok := apiServerMap[document.Items[i].Spec.Sn]
			this_sn := document.Items[i].Spec.Sn
			this_siteId := strconv.Itoa(document.Items[i].Spec.SiteID)
			this_device2Sn := document.Items[i].Spec.Device2.Sn
			//fmt.Println(snName)
			apiServerMap[this_siteId] = append(apiServerMap[this_siteId], this_sn)
			//fmt.Println("device2 ", this_device2Sn)
			apiServerDevice2Map[this_siteId] = append(apiServerDevice2Map[this_siteId], this_device2Sn)
		}
	}
	return apiServerMap, apiServerDevice2Map
}

/*check len different or link different condition*/
func testEq(mysqlMap, apiserverMap, apiServerDevice2SnMap map[string][]string) string {
	var errorMsg = ""
	for key, _ := range mysqlMap {
		if len(mysqlMap[key]) == 1 && len(apiserverMap[key]) > 1 {
			fmt.Println("Err01: mysql site links to ", mysqlMap[key][0], " but apiserver site links to ", strconv.Itoa(len(apiserverMap[key])), " sites, they are:")
			errorMsg += "Err01: mysql site links to " + mysqlMap[key][0] + " but apiserver site links to " + strconv.Itoa(len(apiserverMap[key])) + " sites, they are:"
			err01cnt++
			for i := 0; i < len(apiserverMap[key]); i++ {
				fmt.Println(apiserverMap[key][i])
				errorMsg += "\n" + apiserverMap[key][i] + "\n"
			}
		} else if len(mysqlMap[key]) > 1 && len(apiserverMap[key]) == 1 {
			/*when mysqlMap has two device and it's hadevice, it's ok*/
			for i := 0; i < len(mysqlMap[key]); i++ {
				if mysqlMap[key][i] == apiserverMap[key][0] {
					continue
				} else if mysqlMap[key][i] == apiServerDevice2SnMap[key][0] {
					okCnt++
					fmt.Println("OK: siteId:", key, " mysql: ", mysqlMap[key], " apiserver: ", apiserverMap[key], " Device2: ", apiServerDevice2SnMap[key])
					continue

				} else {
					err02cnt++
					fmt.Println("Err02: site: ", key, " apiserver links to ", apiserverMap[key], "device 2 is ", apiServerDevice2SnMap[key], " but mysql links to ", mysqlMap[key])
					errorMsg += "Err02: site: " + key + " apiserver links to " + strings.Join(apiserverMap[key], ",") + "device 2 is " + strings.Join(apiServerDevice2SnMap[key], ",") + " but mysql links to " + strings.Join(mysqlMap[key], ",") + "\n"
				}
			}
		} else if len(mysqlMap[key]) > 1 && len(apiserverMap[key]) > 1 {
			err03cnt++
			fmt.Println("Err03: site: ", key, " mysql links to ", mysqlMap[key], " but apiserver links to ", apiserverMap[key])
			errorMsg += "Err03: site: " + key + " mysql links to " + strings.Join(mysqlMap[key], ",") + " but apiserver links to " + strings.Join(apiserverMap[key], ",")
		} else if len(mysqlMap[key]) == 1 && len(apiserverMap[key]) == 1 {
			if mysqlMap[key][0] != apiserverMap[key][0] {
				err04cnt++
				fmt.Println("Err04: mysql sn ", key, " link to ", mysqlMap[key][0], " but apiserver sn ", key, " link to ", apiserverMap[key])
				errorMsg += "Err04: mysql sn " + key + " link to " + mysqlMap[key][0] + " but apiserver sn " + key + " link to " + strings.Join(apiserverMap[key], ",")
			} else {
				okCnt++
				fmt.Println("OK: siteId:", key, " mysql: ", mysqlMap[key], " apiserver: ", apiserverMap[key])
				continue

			}
		} else {
			/*mysql always return string whatever it has item*/
			if mysqlMap[key][0] == "" && len(apiserverMap[key]) == 0 {
				okCnt++
				fmt.Println("OK: siteId:", key, " mysql: ", mysqlMap[key], " apiserver: ", apiserverMap[key])
				continue

			} else {
				err05cnt++
				fmt.Println("Err05:  key: ", key, " mysql links to ", mysqlMap[key], " apiserver links to ", apiserverMap[key])
				errorMsg += "Err05:  key: " + key + " mysql links to " + strings.Join(mysqlMap[key], ",") + " apiserver links to " + strings.Join(apiserverMap[key], ",")
			}
		}
	}
	return errorMsg
}

func GetTaipeiTime() string {
	loc, _ := time.LoadLocation("Asia/Taipei")
	//fmt.Println(time.Now().In(loc))
	return time.Now().In(loc).Format(timeFormat)
}
func main() {
	path := "./resultLogs"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}
	file, err := os.Create(path + "/" + GetTaipeiTime() + ".txt")
	checkErr(err)
	defer file.Close()
	dbName := "cubs"
	mysqlDomain := "127.0.0.1:3308"
	apiserverDomain := "127.0.0.1:8080"
	username := "root"
	password := "root"
	fmt.Println("input mysql domain: ")
	fmt.Scanf("%s", &mysqlDomain)
	if mysqlDomain == "" {
		mysqlDomain = "127.0.0.1:3308"
	}
	fmt.Println("input mysql username:")
	fmt.Scanf("%s", &username)
	fmt.Println("input mysql password:")
	fmt.Scanf("%s", &password)
	fmt.Println("input apiserver domain")
	fmt.Scanf("%s", &apiserverDomain)
	if apiserverDomain == "" {
		apiserverDomain = "127.0.0.1:8080"
	}
	fmt.Println("mysql login as root/root db:", dbName)
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName)
	checkErr(err)
	err = db.Ping()
	checkErr(err)
	rows, _ := db.Query("SELECT serial,siteId FROM cubs.device;")
	defer rows.Close()
	mysqlMap := getMysqlMap(rows)
	apiServerMap, apiServerDevice2SnMap := getApiserverMap(apiserverDomain)
	fmt.Println("=========================================Message==================================================")
	errorMsg := testEq(mysqlMap, apiServerMap, apiServerDevice2SnMap)
	fmt.Println("=========================================Error Statistics==========================================")
	fmt.Println("err01: ", err01cnt, " err02: ", err02cnt, " err03: ", err03cnt, " err04: ", err04cnt, " err05: ", err05cnt, " ok: ", okCnt)
	fmt.Printf("save msg to %s/%s.txt \n", path, GetTaipeiTime())
	file.WriteString(errorMsg)

}
