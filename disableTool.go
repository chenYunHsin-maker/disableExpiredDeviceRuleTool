package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	_ "github.com/syhlion/sqlwrapper"
)

const (
	buisnessPolicyUrl       = "/apis/businesspolicy/v1alpha1/namespaces/default/bprulesets/"
	firewallPolicyUrl       = "/apis/firewall/v1alpha1/namespaces/default/fwpolicies/"
	siteconfigUrl           = "/apis/site/v1alpha1/namespaces/default/siteconfigs"
	timeFormat              = "2006-01-02"
	dbName_default          = "cubs"
	mysqlDomain_default     = "127.0.0.1:3308"
	apiserverDomain_default = "http://127.0.0.1:8080"
	username_default        = "root"
	password_default        = "root"
	from_date_default       = ""
	to_date_default         = ""
)

var (
	dbName          = dbName_default
	mysqlDomain     = mysqlDomain_default
	apiserverDomain = apiserverDomain_default
	username        = username_default
	password        = password_default
	from_date       = from_date_default
	to_date         = to_date_default
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

func putRequest(url string, data io.Reader) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, data)
	checkErr(err)
	_, err = client.Do(req)
	checkErr(err)
}
func updateJson(str string) string {
	i := 0
	for strings.Index(str, "\"enabled\":true,") != -1 {
		i++
		str = strings.Replace(str, "\"enabled\":true,", "\"enabled\":false,", 1)
	}
	fmt.Println("update ", i, "apiserver enabled fileds")
	return str
}
func GetTaiwanTime() time.Time {
	loc, _ := time.LoadLocation("Asia/Taipei")
	//fmt.Println(time.Now().In(loc))
	t, _ := ShortDateFromString(time.Now().In(loc).Format(timeFormat))
	return t
}
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
func getApiserverBody(apiserverDomain, siteUrl string) string {
	//fmt.Println("start to get body")
	resp, err := http.Get(apiserverDomain + siteUrl)
	checkErr(err)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	s := string(body)
	//fmt.Println(s)
	return s
}
func getApiserverMap(apiserverDomain string, snExpiredMap map[string]string) (map[string]string, map[string]string, map[string]string) {
	apiServerMap := make(map[string]string)
	siteNameToSiteIdMap := make(map[string]string)
	snToSiteId := make(map[string]string)
	//apiServerDevice2Map := make(map[string][]string)
	body := getApiserverBody(apiserverDomain, siteconfigUrl)
	var document Document
	var data []byte = []byte(body)
	json.Unmarshal(data, &document)
	//apiServerMap["S162L45290036"] = append(apiServerMap["S162L45290036"], "88")
	for i := 0; i < len(document.Items); i++ {
		if document.Items[i].Spec.Sn != "" {
			//_, ok := apiServerMap[document.Items[i].Spec.Sn]
			this_sn := document.Items[i].Spec.Sn
			this_siteName := document.Items[i].Metadata.Name
			this_siteId := strconv.Itoa(document.Items[i].Spec.SiteID)
			//this_device2Sn := document.Items[i].Spec.Device2.Sn
			//fmt.Println(snName)
			if _, ok := snExpiredMap[this_sn]; ok {
				snToSiteId[this_sn] = this_siteId
				siteNameToSiteIdMap[this_siteName] = this_siteId
				apiServerMap[this_siteName] = this_sn
			}

			//fmt.Println("device2 ", this_device2Sn)
			//apiServerDevice2Map[this_siteName] = append(apiServerDevice2Map[this_siteName], this_device2Sn)
		}
	}
	return apiServerMap, siteNameToSiteIdMap, snToSiteId
}
func checkTable(snSiteLinkedMap map[string][]string) {
	fmt.Println("start to check map......")
	fmt.Println("your map: ")
	for key, value := range snSiteLinkedMap {
		fmt.Println("Key:", key, "Value:", value)
	}
	fmt.Println("map check end :D")
}
func checkTableS(snSiteLinkedMap map[string]string) {
	fmt.Println("start to check map......")
	fmt.Println("your map: ")
	for key, value := range snSiteLinkedMap {
		fmt.Println("Key:", key, "Value:", value)
	}
	fmt.Println("map check end :D")
}
func getMysqlMap(rows *sql.Rows) map[string]string {
	snExpiredMap := make(map[string]string)
	for rows.Next() {
		var expired sql.NullTime
		var sn sql.NullString
		if err := rows.Scan(&sn, &expired); err != nil {
			fmt.Println(" err :", err)
		}
		if sn.Valid == true {
			if sn.String != "" {
				snExpiredMap[strings.Replace(sn.String, "default", "", 1)] = expired.Time.Format(timeFormat)
			}

			//fmt.Println("sn:", sn.String, " linked site id:", site.String)
		}
	}
	return snExpiredMap
}
func getMysqlProfilePolicyRule(snToSite map[string]string) (map[string][]string, map[string][]string) {
	//policyIdToRuleTypeMap := make(map[string][]string)

	//policyIdToRuleNameMap := make(map[string][]string)
	var id sql.NullString
	var ruleName sql.NullString
	var ruleType sql.NullString
	var enabled sql.NullString
	var policyId sql.NullString
	siteIdToIdMap := make(map[string][]string)
	siteIdToBusnessNamesMap := make(map[string][]string)
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName)
	checkErr(err)
	command_part := "SELECT  id,ruleName,ruleType,enabled,policyId FROM cubs.profile_policy_rule WHERE `policyId`="

	for key, _ := range snToSite {
		command := command_part + snToSite[key] + " AND ruleType='site'"
		//fmt.Println(command)
		rows, _ := db.Query(command)

		for rows.Next() {
			if err := rows.Scan(&id, &ruleName, &ruleType, &enabled, &policyId); err != nil {
				fmt.Println(" err :", err)
			}
			siteIdToIdMap[policyId.String] = append(siteIdToIdMap[policyId.String], id.String)
			siteIdToBusnessNamesMap[policyId.String] = append(siteIdToBusnessNamesMap[policyId.String], ruleName.String)
		}
		rows.Close()
	}

	return siteIdToIdMap, siteIdToBusnessNamesMap
}
func getMysqlFirewallRule(snToSiteid map[string]string) (map[string][]string, map[string][]string) {
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName+"?charset=utf8&parseTime=True")
	checkErr(err)

	var id sql.NullString
	var ruleName sql.NullString
	var ruleType sql.NullString
	var firewallId sql.NullString
	command_part := "SELECT  id,ruleName,ruleType,firewallId FROM cubs.profile_firewall_rule WHERE ruleType='site' AND firewallId="
	siteIdToFirewallIdMap := make(map[string][]string)
	siteIdToFirewallNamesMap := make(map[string][]string)
	for key, _ := range snToSiteid {
		command := command_part + snToSiteid[key]
		fmt.Println(command)
		rows, _ := db.Query(command)
		for rows.Next() {
			if err := rows.Scan(&id, &ruleName, &ruleType, &firewallId); err != nil {
				fmt.Println(" err :", err)
			}
			siteIdToFirewallIdMap[firewallId.String] = append(siteIdToFirewallIdMap[firewallId.String], id.String)
			siteIdToFirewallNamesMap[firewallId.String] = append(siteIdToFirewallNamesMap[firewallId.String], ruleName.String)
		}
		rows.Close()

	}

	return siteIdToFirewallIdMap, siteIdToFirewallNamesMap
}
func getSiteIdToPolicyBName(snToSite map[string]string) map[string][]string {
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName+"?charset=utf8&parseTime=True")
	checkErr(err)
	siteIdToPolicyBName := make(map[string][]string)
	command_part := "SELECT siteId,beName FROM cubs.site_policy WHERE siteId="
	var id sql.NullString
	var beName sql.NullString
	for key, _ := range snToSite {
		command := command_part + snToSite[key]
		fmt.Println(command)
		rows, _ := db.Query(command)
		for rows.Next() {
			if err := rows.Scan(&id, &beName); err != nil {
				fmt.Println(" err :", err)
			}
			siteIdToPolicyBName[id.String] = append(siteIdToPolicyBName[id.String], beName.String)
		}
		rows.Close()

	}
	/*
		for rows.Next() {
			var id sql.NullString
			var beName sql.NullString
			if err := rows.Scan(&id, &beName); err != nil {
				fmt.Println(" err :", err)
			}
			siteIdToPolicyBName[id.String] = append(siteIdToPolicyBName[id.String], beName.String)
		}*/
	return siteIdToPolicyBName
}
func getSiteIdToFirewallBName(snToSiteMap map[string]string) map[string][]string {
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName+"?charset=utf8&parseTime=True")
	checkErr(err)
	siteIdToFirewallBName := make(map[string][]string)
	command_part := "SELECT siteId,beName FROM cubs.site_firewall WHERE siteId="
	var id sql.NullString
	var beName sql.NullString
	for key, _ := range snToSiteMap {
		command := command_part + snToSiteMap[key]
		rows, _ := db.Query(command)
		for rows.Next() {
			if err := rows.Scan(&id, &beName); err != nil {
				fmt.Println(" err :", err)
			}
			siteIdToFirewallBName[id.String] = append(siteIdToFirewallBName[id.String], beName.String)
		}
		rows.Close()

	}
	/*
		for rows.Next() {
			var id sql.NullString
			var beName sql.NullString
			if err := rows.Scan(&id, &beName); err != nil {
				fmt.Println(" err :", err)
			}
			siteIdToFirewallBName[id.String] = append(siteIdToFirewallBName[id.String], beName.String)
		}*/
	return siteIdToFirewallBName
}
func initDb() {

	fmt.Println("input mysql domain: ")
	fmt.Scanf("%s", &mysqlDomain)
	if mysqlDomain == "" {
		mysqlDomain = mysqlDomain_default
	}
	fmt.Println("input mysql username:")
	fmt.Scanf("%s", &username)
	fmt.Println("input mysql password:")
	fmt.Scanf("%s", &password)
	fmt.Println("input apiserver domain")
	fmt.Scanf("%s", &apiserverDomain)
	fmt.Println("watch license from date(format: 2020-02-12)")
	fmt.Scanf("%s", &from_date)
	fmt.Println("watch license to date(format: 2020-02-12)")
	fmt.Scanf("%s", &to_date)
	if apiserverDomain == "" {
		apiserverDomain = apiserverDomain_default
	}
	if username == "" {
		username = username_default
	}
	if password == "" {
		password = password_default
	}

	fmt.Println("mysql login in db:", dbName)
}
func updateMysqlEnableStatement(command string, idArr []string) {
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName)
	checkErr(err)
	for i := 0; i < len(idArr); i++ {
		command := command + idArr[i] + ";"
		db.Exec(command)
	}
}
func updateApiserver(beName, fBname string) {
	targetUrl := apiserverDomain + buisnessPolicyUrl + beName
	targetUrlF := apiserverDomain + firewallPolicyUrl + fBname
	putToApiserver(targetUrl)
	putToApiserver(targetUrlF)
}
func putToApiserver(targetUrl string) {
	resp, err := http.Get(targetUrl)
	checkErr(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	checkErr(err)
	str := string(body[:])
	str = updateJson(str)
	putRequest(targetUrl, strings.NewReader(str))
}
func generateLog(this_sn, this_site_id string, siteIdToBusinessNamesMap, siteIdToFirewallNamesMap, siteIdToPolicyBName, siteIdToFirewallBName map[string][]string, today, expired_date time.Time) {
	glog.Infof("sn %s 's license is expired! today is %s expired_day is %s,will disable %d business rules %d firewall rules\n", this_sn, today.Format(timeFormat), expired_date.Format(timeFormat), len(siteIdToBusinessNamesMap[this_site_id]), len(siteIdToFirewallNamesMap[this_site_id]))
	for i := 0; i < len(siteIdToBusinessNamesMap[this_site_id]); i++ {
		glog.Infof("disable business rule: %s\n", siteIdToBusinessNamesMap[this_site_id][i])
	}
	for i := 0; i < len(siteIdToFirewallNamesMap[this_site_id]); i++ {
		glog.Infof("disable firewall rule: %s\n", siteIdToFirewallNamesMap[this_site_id][i])
	}
	glog.Infof("business rule beName: %s\n", siteIdToPolicyBName[this_site_id][0])
	glog.Infof("firewall rule beName: %s\n", siteIdToFirewallBName[this_site_id][0])
}
func checkDeviceLicense(snExpiredMap, siteNameToSnMap, siteNameToSiteIdMap map[string]string, policyIdToIdMap, FpolicyIdToIdMap, siteIdToPolicyBName, siteIdToFirewallBName, siteIdToBusinessNamesMap, siteIdToFirewallNamesMap map[string][]string) {
	policyUpdateCmd := "UPDATE cubs.profile_policy_rule SET `enabled` = 0  WHERE `id`="
	firewallUpdateCmd := "UPDATE cubs.profile_firewall_rule SET `enabled` = 0  WHERE `id`="
	for key, _ := range siteNameToSnMap {
		this_site_name := key
		this_sn := siteNameToSnMap[this_site_name]
		//today := GetTaiwanTime()
		if snExpiredMap[this_sn] != "" {
			today := GetTaiwanTime()
			expired_date, _ := ShortDateFromString(snExpiredMap[this_sn])
			if expired_date.Before(today) {
				this_site_id := siteNameToSiteIdMap[key]
				idMap := policyIdToIdMap[this_site_id]
				fIdMap := FpolicyIdToIdMap[this_site_id]
				generateLog(this_sn, this_site_id, siteIdToBusinessNamesMap, siteIdToFirewallNamesMap, siteIdToPolicyBName, siteIdToFirewallBName, today, expired_date)
				updateMysqlEnableStatement(policyUpdateCmd, idMap)
				updateMysqlEnableStatement(firewallUpdateCmd, fIdMap)
				updateApiserver(siteIdToPolicyBName[this_site_id][0], siteIdToFirewallBName[this_site_id][0])
				fmt.Printf("sn %s 's license is expired! today is %s expired_day is %s \n", this_sn, today.Format(timeFormat), expired_date.Format(timeFormat))
				fmt.Printf("change enabled to False. \n")
				fmt.Println("beName: ", siteIdToPolicyBName[this_site_id])
				fmt.Println("fire bName: ", siteIdToFirewallBName[this_site_id])
				fmt.Println("policy rule ids:", idMap, " firewall rule ids:", fIdMap, " site ids: ", this_site_id)

			}
		}
	}
}
func ShortDateFromString(ds string) (time.Time, error) {
	t, err := time.Parse(timeFormat, ds)
	if err != nil {
		return t, err
	}
	return t, err
}
func getSnExpiredQuery() *sql.Rows {
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName+"?charset=utf8&parseTime=True")
	checkErr(err)
	var rows *sql.Rows
	switch {
	case from_date == "" && to_date == "":
		command := "SELECT contractId,last_expired_at FROM cubs.contract;"
		rows, _ = db.Query(command)
	case from_date == "":
		command := "SELECT contractId,last_expired_at FROM cubs.contract  WHERE DATE(last_expired_at) < '" + to_date + "'"
		rows, _ = db.Query(command)
	case to_date == "":
		command := "SELECT contractId,last_expired_at FROM cubs.contract  WHERE DATE(last_expired_at) > '" + from_date + "'"
		rows, _ = db.Query(command)
	case from_date != "" && to_date != "":
		command := "SELECT contractId,last_expired_at FROM cubs.contract  WHERE DATE(last_expired_at) BETWEEN  '" + from_date + "' AND '" + to_date + "'"
		rows, err = db.Query(command)
		fmt.Println(command)
	}

	return rows
}
func tmp(a map[string]string, b map[string]string, c map[string]string) {

}

func main() {
	flag.Parse()
	defer glog.Flush()
	initDb()
	//db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName)
	//checkErr(err)

	//rows2, _ := db.Query("SELECT  id,ruleName,ruleType,enabled,policyId FROM cubs.profile_policy_rule;")
	//rows3, _ := db.Query("SELECT  id,ruleName,ruleType,firewallId FROM cubs.profile_firewall_rule;")
	//rows4, _ := db.Query("SELECT siteId,beName FROM cubs.site_policy;")
	//rows5, _ := db.Query("SELECT siteId,beName FROM cubs.site_firewall;")

	//defer rows2.Close()
	//defer rows3.Close()
	//defer rows4.Close()
	//defer rows5.Close()

	snExpiredMap := getMysqlMap(getSnExpiredQuery())
	//checkTableS(snExpiredMap)

	siteNameToSnMap, siteNameToSiteIdMap, snToSiteId := getApiserverMap(apiserverDomain, snExpiredMap)
	tmp(snExpiredMap, siteNameToSnMap, siteNameToSiteIdMap)
	//checkTableS(snExpiredMap)
	//checkTableS(snToSiteId)
	BpolicyIdToIdMap, siteIdToBusnessNamesMap := getMysqlProfilePolicyRule(snToSiteId)
	//checkTable(BpolicyIdToIdMap)
	//fmt.Println("==================================")
	//checkTable(siteIdToBusnessNamesMap)

	FpolicyIdToIdMap, siteIdToFirewallNamesMap := getMysqlFirewallRule(snToSiteId)
	//====before ok
	//checkTable(FpolicyIdToIdMap)
	//checkTable(siteIdToFirewallNamesMap)
	siteIdToPolicyBName := getSiteIdToPolicyBName(snToSiteId)
	//checkTable(siteIdToPolicyBName)
	siteIdToFirewallBName := getSiteIdToFirewallBName(snToSiteId)
	//checkTable(siteIdToFirewallBName)

	checkDeviceLicense(snExpiredMap, siteNameToSnMap, siteNameToSiteIdMap, BpolicyIdToIdMap, FpolicyIdToIdMap, siteIdToPolicyBName, siteIdToFirewallBName, siteIdToBusnessNamesMap, siteIdToFirewallNamesMap)

}
