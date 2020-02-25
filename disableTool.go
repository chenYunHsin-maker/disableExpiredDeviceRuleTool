package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	_ "github.com/syhlion/sqlwrapper"
	"github.com/tidwall/gjson"
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
	detailTime              = "2006-01-02_150405"
)

var (
	siteToB         = make(map[string][]string)
	siteToF         = make(map[string][]string)
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
func ShortDateFromString(ds string) (time.Time, error) {
	t, err := time.Parse(timeFormat, ds)
	if err != nil {
		return t, err
	}
	return t, err
}
func ShortDateFromString2(ds string) (time.Time, error) {
	t, err := time.Parse(detailTime, ds)
	//fmt.Println("s:", t)
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
func getSpecStr(bName string) string {
	var body string
	switch bName[0:1] {
	case "b":
		//fmt.Println("case b")
		fmt.Println(apiserverDomain, buisnessPolicyUrl+bName)
		body = getApiserverBody(apiserverDomain, buisnessPolicyUrl+bName)

	case "f":
		//fmt.Println("case f")
		fmt.Println(apiserverDomain, firewallPolicyUrl+bName)
		body = getApiserverBody(apiserverDomain, firewallPolicyUrl+bName)
	}
	//fmt.Println("body:", body)
	val := gjson.Get(body, "spec")
	fmt.Println("spec: ", val)
	return val.String()
}
func getApiserverMap(apiserverDomain string, snExpiredMap map[string]string) (map[string]string, map[string]string, map[string]string, []string) {
	apiServerMap := make(map[string]string)
	siteNameToSiteIdMap := make(map[string]string)
	snToSiteId := make(map[string]string)
	var siteIds []string
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
			siteIds = append(siteIds, this_siteId)
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
	return apiServerMap, siteNameToSiteIdMap, snToSiteId, siteIds
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
		fmt.Println(command)
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
func updateApiserver(beName, fBname string, siteId string) {
	targetUrl := apiserverDomain + buisnessPolicyUrl + beName
	targetUrlF := apiserverDomain + firewallPolicyUrl + fBname
	putToApiserver(targetUrl, siteId)
	putToApiserver(targetUrlF, siteId)
}
func putToApiserver(targetUrl string, siteId string) {
	resp, err := http.Get(targetUrl)
	checkErr(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	checkErr(err)

	str := string(body[:])
	str = updateJson(str)

	oldSpec := gjson.Get(str, "spec")
	var newSpec string
	fmt.Println("spec: ", oldSpec.String())
	switch strings.Contains(targetUrl, "businesspolicy") {
	case true:
		newSpec = "{" + "\"disabledProfileRuleIds\":" + sliceToString(siteToB[siteId]) + "}"
	case false:
		newSpec = "{" + "\"disabledProfileRuleIds\":" + sliceToString(siteToF[siteId]) + "}"
	}
	str = strings.Replace(str, oldSpec.String(), newSpec, 1)
	fmt.Println(str)
	fmt.Println()
	putRequest(targetUrl, strings.NewReader(str))
}
func sliceToString(s []string) string {
	str := "["
	for i := 0; i < len(s); i++ {
		if i != len(s)-1 {
			str += s[i] + ","
		} else {
			str += s[i] + "]"
		}

	}
	return str
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
func checkDeviceLicense(snExpiredMap, siteNameToSnMap, siteNameToSiteIdMap map[string]string, policyIdToIdMap, FpolicyIdToIdMap, siteIdToPolicyBName, siteIdToFirewallBName, siteIdToBusinessNamesMap, siteIdToFirewallNamesMap map[string][]string, siteIdName map[string]string) {
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
				updateApiserver(siteIdToPolicyBName[this_site_id][0], siteIdToFirewallBName[this_site_id][0], this_site_id)
				pushLogToMysql(this_site_id, siteIdName[this_site_id], "Business Policy", siteIdToBusinessNamesMap[this_site_id])
				pushLogToMysql(this_site_id, siteIdName[this_site_id], "Firewall", siteIdToFirewallNamesMap[this_site_id])
				fmt.Printf("sn %s 's license is expired! today is %s expired_day is %s \n", this_sn, today.Format(timeFormat), expired_date.Format(timeFormat))
				fmt.Printf("change enabled to False. \n")
				fmt.Println("beName: ", siteIdToPolicyBName[this_site_id])
				fmt.Println("fire bName: ", siteIdToFirewallBName[this_site_id])
				fmt.Println("policy rule ids:", idMap, " firewall rule ids:", fIdMap, " site ids: ", this_site_id)

			}
		}
	}
}

func getSnExpiredQuery() *sql.Rows {
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName+"?charset=utf8&parseTime=True")
	checkErr(err)
	var rows *sql.Rows
	var command string
	switch {
	case from_date == "" && to_date == "":
		command = "SELECT contractId,last_expired_at FROM cubs.contract;"
		rows, _ = db.Query(command)
	case from_date == "":
		command = "SELECT contractId,last_expired_at FROM cubs.contract  WHERE DATE(last_expired_at) < '" + to_date + "'"
		rows, _ = db.Query(command)
	case to_date == "":
		command = "SELECT contractId,last_expired_at FROM cubs.contract  WHERE DATE(last_expired_at) > '" + from_date + "'"
		rows, _ = db.Query(command)
	case from_date != "" && to_date != "":
		command = "SELECT contractId,last_expired_at FROM cubs.contract  WHERE DATE(last_expired_at) BETWEEN  '" + from_date + "' AND '" + to_date + "'"
		rows, err = db.Query(command)
		//fmt.Println(command)
	}
	fmt.Println(command)
	return rows
}

func pushLogToMysql(siteId string, siteName string, ruleType string, rules []string) {

	userId := -1
	userName := "License Check"
	modulePage := "Site > Configuration > Device"
	var requestDesc string
	var feature string
	var newValue string
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName+"?charset=utf8&parseTime=True")
	switch ruleType {
	case "Business Policy":
		feature = "Business Policy"
		requestDesc = "PUT /rest/site/updatesitepolicy/" + siteId
		//newValue = "[Update]/rest/site/updatesitepolicy/" + siteId
	case "Firewall":
		feature = "Firewall"
		requestDesc = "PUT /rest/site/updatesitefirewall/" + siteId
		//newValue = "[Update]/rest/site/updatesitefirewall/" + siteId
	}
	newValue += "<br>Disable the following rule(s): <br>"

	for i := 0; i < len(rules); i++ {
		newValue += rules[i] + "<br>"
	}
	//newValue += "wioeuvwmiecrawurioauwpoieruaemcrweurcmioawecrmpaweucmaewauweopiurcu"
	_, err = db.Exec(
		"INSERT INTO cubs.auth_useractivitylog(userId,userName,layerId,layerName,modulePage,feature,requestDesc,newValue) VALUES (?,?,?,?,?,?,?,?)",
		userId,
		userName,
		siteId,
		siteName,
		modulePage,
		feature,
		requestDesc,
		newValue,
	)
	checkErr(err)
	fmt.Println("pushLog:", userName, userId, siteId, siteName, modulePage, feature, requestDesc)
}
func getMysqlSiteIdToSiteNameMap(siteIds []string) map[string]string {
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName+"?charset=utf8&parseTime=True")
	checkErr(err)
	siteIdToName := make(map[string]string)
	var rows *sql.Rows
	for i := 0; i < len(siteIds); i++ {
		command := "SELECT siteName FROM cubs.site WHERE siteId = " + siteIds[i]
		rows, _ = db.Query(command)
		for rows.Next() {
			var name sql.NullString
			if err := rows.Scan(&name); err != nil {
				fmt.Println(" err :", err)
			}
			siteIdToName[siteIds[i]] = name.String
		}
		rows.Close()
	}
	return siteIdToName
}
func getSiteIdToOrderedIdMaps(siteIds []string) (map[string][]string, map[string][]string) {
	//siteIds include expired and unexpired!
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName+"?charset=utf8&parseTime=True")
	checkErr(err)
	var siteId string
	var profileId string

	siteToP := make(map[string]string)
	//pToOrders := make(map[string][]string)
	siteToBOrders := make(map[string][]string)
	siteToFOrders := make(map[string][]string)

	cmd := "SELECT siteId,profileId FROM cubs.site WHERE siteId = "
	for i := 0; i < len(siteIds); i++ {
		//fmt.Println(cmd + siteIds[i] + ";")
		row := db.QueryRow(cmd + siteIds[i] + ";")

		err = row.Scan(&siteId, &profileId)
		siteToP[siteId] = profileId
		//checkTableS(siteToP)
	}

	for key, value := range siteToP {
		cmd = "SELECT orderId,policyId FROM cubs.profile_policy_rule WHERE ruleType='profile' && policyId="
		cmd2 := "SELECT orderId,firewallId FROM cubs.profile_firewall_rule  WHERE ruleType='profile' && firewallId="
		cmd = cmd + value + ";"

		rows, _ := db.Query(cmd)
		for rows.Next() {
			var orderId string
			var policyId string
			if err := rows.Scan(&orderId, &policyId); err != nil {
				checkErr(err)
			}
			siteToBOrders[key] = append(siteToBOrders[key], orderId)
		}

		cmd2 = cmd2 + value + ";"
		rows2, _ := db.Query(cmd2)
		for rows2.Next() {
			var orderId string
			var firewallId string
			if err := rows2.Scan(&orderId, &firewallId); err != nil {
				checkErr(err)
			}
			siteToFOrders[key] = append(siteToFOrders[key], orderId)
		}
	}
	return siteToBOrders, siteToFOrders
}
func tmpF(m1 map[string][]string, m2 map[string][]string) {
	for key, _ := range m1 {
		//fmt.Println(key, " ", m2[key])
		if m2[key] != nil {
			fmt.Println("spec: ", getSpecStr(m2[key][0]))
		} else {
			fmt.Println("nil:", m2[key])
		}

	}
}
func main() {
	flag.Parse()
	defer glog.Flush()
	//initDb()
	/*
		for i := 0; i < len(os.Args); i++ {
			fmt.Println("arg:", os.Args[i])
		}*/
	mysqlDomain = os.Args[1]
	username = os.Args[2]
	password = os.Args[3]
	apiserverDomain = os.Args[4]
	from_date = os.Args[5]
	to_date = os.Args[6]
	fmt.Println("timestamp:", GetTaiwanTime2().Format(detailTime))
	fmt.Println("mysql domain:", mysqlDomain)
	fmt.Println("mysql username:", username)
	fmt.Println("mysql password:", password)
	fmt.Println("apiserver domain:", apiserverDomain)
	fmt.Println("from date", from_date)
	fmt.Println("to date:", to_date)
	//fmt.Println("args4:",os.Args[4])

	snExpiredMap := getMysqlMap(getSnExpiredQuery())
	siteNameToSnMap, siteNameToSiteIdMap, snToSiteId, siteIds := getApiserverMap(apiserverDomain, snExpiredMap)
	checkTableS(siteNameToSnMap)
	checkTableS(siteNameToSiteIdMap)
	checkTableS(snToSiteId)

	siteIdName := getMysqlSiteIdToSiteNameMap(siteIds)

	BpolicyIdToIdMap, siteIdToBusnessNamesMap := getMysqlProfilePolicyRule(snToSiteId)
	FpolicyIdToIdMap, siteIdToFirewallNamesMap := getMysqlFirewallRule(snToSiteId)
	siteIdToPolicyBName := getSiteIdToPolicyBName(snToSiteId)
	siteIdToFirewallBName := getSiteIdToFirewallBName(snToSiteId)
	checkDeviceLicense(snExpiredMap, siteNameToSnMap, siteNameToSiteIdMap, BpolicyIdToIdMap, FpolicyIdToIdMap, siteIdToPolicyBName, siteIdToFirewallBName, siteIdToBusnessNamesMap, siteIdToFirewallNamesMap, siteIdName)

	siteToB, siteToF = getSiteIdToOrderedIdMaps(siteIds)
	//checkTable(siteToB)
	//checkTable(siteToF)
	//tmpF(siteToB, siteIdToPolicyBName)
	//tmpF(siteToF, siteIdToFirewallBName)
	//checkTableS(siteIdName)
	//getSpecStr("fw-q6f9n")
	//getSpecStr("bp-xqsc8")
}
