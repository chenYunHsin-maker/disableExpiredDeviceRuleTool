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
	siteToB               = make(map[string][]string)
	siteToF               = make(map[string][]string)
	siteIdToPolicyBName   = make(map[string][]string)
	siteIdToFirewallBName = make(map[string][]string)
	dbName                = dbName_default
	mysqlDomain           = mysqlDomain_default
	apiserverDomain       = apiserverDomain_default
	username              = username_default
	password              = password_default
	from_date             = from_date_default
	to_date               = to_date_default
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
	glog.Infoln("update ", i, "apiserver enabled fileds")
	return str
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
func ShortDateFromString2(ds string) (time.Time, error) {
	t, err := time.Parse(detailTime, ds)
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
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
func getApiserverBody(apiserverDomain, siteUrl string) string {
	resp, err := http.Get(apiserverDomain + siteUrl)
	checkErr(err)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	s := string(body)
	return s
}
func getSpecStr(bName string) string {
	var body string
	switch bName[0:1] {
	case "b":
		glog.Infoln(apiserverDomain, buisnessPolicyUrl+bName)
		body = getApiserverBody(apiserverDomain, buisnessPolicyUrl+bName)
	case "f":
		glog.Infoln(apiserverDomain, firewallPolicyUrl+bName)
		body = getApiserverBody(apiserverDomain, firewallPolicyUrl+bName)
	}
	val := gjson.Get(body, "spec")
	glog.Infoln("spec: ", val)
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
	//fmt.Println("start to check map......")
	//fmt.Println("your map: ")
	for key, value := range snSiteLinkedMap {
		fmt.Println("Key:", key, "Value:", value)
	}
	fmt.Println("map check end :D")
}
func checkTableS(snSiteLinkedMap map[string]string) {
	//fmt.Println("start to check map......")
	//fmt.Println("your map: ")
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
			glog.Infoln(err.Error())
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
			if checkMysqlSdOnePolicy(id.String) {
				siteIdToIdMap[policyId.String] = append(siteIdToIdMap[policyId.String], id.String)
				siteIdToBusnessNamesMap[policyId.String] = append(siteIdToBusnessNamesMap[policyId.String], ruleName.String)
			}
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
		//fmt.Println(command)
		rows, _ := db.Query(command)
		for rows.Next() {
			if err := rows.Scan(&id, &ruleName, &ruleType, &firewallId); err != nil {
				glog.Infoln(err.Error())
			}
			if checkMysqlSdOneFirewall(id.String) {
				siteIdToFirewallIdMap[firewallId.String] = append(siteIdToFirewallIdMap[firewallId.String], id.String)
				siteIdToFirewallNamesMap[firewallId.String] = append(siteIdToFirewallNamesMap[firewallId.String], ruleName.String)
			}

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
		//fmt.Println(command)
		rows, _ := db.Query(command)
		for rows.Next() {
			if err := rows.Scan(&id, &beName); err != nil {
				glog.Infoln(err.Error())
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
				glog.Infoln(err.Error())
			}
			siteIdToFirewallBName[id.String] = append(siteIdToFirewallBName[id.String], beName.String)
		}
		rows.Close()

	}
	return siteIdToFirewallBName
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
	targetUrl := buisnessPolicyUrl + beName
	targetUrlF := firewallPolicyUrl + fBname
	needClosedIdB := checkSdwanBusinessApi(targetUrl)
	needClosedIdF := checkSdwanFirewallApi(targetUrlF)
	if len(needClosedIdB) > 0 {
		glog.Infoln("update site policy:", needClosedIdB)
	} else {
		glog.Infoln("update site policy: no change")
	}
	if len(needClosedIdF) > 0 {
		glog.Infoln("update site firewall:", needClosedIdF)
	} else {
		glog.Infoln("update site firewall: no change")
	}
	putToApiserver(apiserverDomain+targetUrl, siteId, needClosedIdB)
	putToApiserver(apiserverDomain+targetUrlF, siteId, needClosedIdF)
}
func replaceNth(s string, n int) string {
	old := "\"orderId\":" + strconv.Itoa(n) + "," + "\"enabled\":true"
	new := "\"orderId\":" + strconv.Itoa(n) + "," + "\"enabled\":false"
	old2 := "\"orderId\":" + strconv.Itoa(n)
	new2 := "\"orderId\":" + strconv.Itoa(n) + "," + "\"enabled\":false"
	if strings.Index(s, old) != -1 {
		s = strings.Replace(s, old, new, 1)
		//fmt.Println("kind1")
	} else {
		s = strings.Replace(s, old2, new2, 1)
		//fmt.Println("kind2")
	}
	return s
}
func checkSdwanBusinessApi(url string) []int {
	body := getApiserverBody(apiserverDomain, url)
	ruleMap := gjson.Get(body, "spec.policyRules").Array()

	var needClosedIdB []int
	for i := 0; i < len(ruleMap); i++ {
		targetStr := ruleMap[i].String()
		isSdwan := false
		if gjson.Get(targetStr, "action.networkService.pathSelectClassParams").Bool() {
			isSdwan = true
		}
		if gjson.Get(targetStr, "action.networkService.aggregation").Bool() {
			isSdwan = true
		}
		if gjson.Get(ruleMap[i].String(), "action.networkService.forwardErrorCorrect").Bool() {
			isSdwan = true
		}
		if gjson.Get(targetStr, "service.serviceType").String() == "appGroup" {
			isSdwan = true
		}
		if gjson.Get(targetStr, "source.selectType").String() == "custom" && gjson.Get(ruleMap[i].String(), "source.customAddrType").String() == "country" {
			isSdwan = true
		}
		if gjson.Get(targetStr, "destination.selectType").String() == "custom" && gjson.Get(ruleMap[i].String(), "destination.customAddrType").String() == "country" {
			isSdwan = true
		}
		if isSdwan {
			needClosedIdB = append(needClosedIdB, i)
		}
	}
	return needClosedIdB
}

func checkSdwanFirewallApi(url string) []int {
	body := getApiserverBody(apiserverDomain, url)
	ruleMap := gjson.Get(body, "spec.firewallRules").Array()
	//fmt.Println(gjson.Get(targetStr, "source.selectType").String())
	var needClosedIdF []int
	for i := 0; i < len(ruleMap); i++ {
		targetStr := ruleMap[i].String()
		isSdwan := false
		//fmt.Println(gjson.Get(targetStr, "source.selectType").String(), gjson.Get(targetStr, "source.customAddrType").String())
		if gjson.Get(targetStr, "action.isAllow").Bool() {
			//fmt.Println("MAOMAO1", url)
			if gjson.Get(targetStr, "action.blockAppGroup.id").Exists() {
				isSdwan = true
			}
			if gjson.Get(targetStr, "action.customBlockSites.length").Int() > 0 {
				isSdwan = true
			}
			if gjson.Get(targetStr, "selectedBlockPages.length").Int() > 0 {
				isSdwan = true
			}
		}
		if gjson.Get(targetStr, "source.selectType").String() == "custom" && gjson.Get(targetStr, "source.customAddrType").String() == "country" {
			isSdwan = true
		}
		if gjson.Get(targetStr, "destination.selectType").String() == "custom" && gjson.Get(targetStr, "destination.customAddrType").String() == "country" {
			isSdwan = true
		}
		if isSdwan {
			needClosedIdF = append(needClosedIdF, i)
		}
	}
	return needClosedIdF
}
func putToApiserver(targetUrl string, siteId string, orderId []int) {
	resp, err := http.Get(targetUrl)
	checkErr(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	str := string(body[:])
	checkErr(err)

	for i := len(orderId) - 1; i >= 0; i-- {
		//fmt.Println("update order:", orderId[i]+1)
		str = replaceNth(str, orderId[i]+1)
	}
	str = addOrderIds(str, siteId)
	putRequest(targetUrl, strings.NewReader(str))

}
func addOrderIds(s, siteId string) string {
	oldSpec := gjson.Get(s, "spec").String()
	newSpec := oldSpec
	//fmt.Println("oldspec:", oldSpec)
	if strings.Index(s, "BPRuleSet") != -1 {
		if len(siteToB[siteId]) != 0 {
			glog.Infoln("update profile policy:", siteToB[siteId])
			newSpec = newSpec[0:len(newSpec)-1] + "," + "\"disabledProfileRuleIds\":" + sliceToString(siteToB[siteId]) + "}"

		}
	} else {
		if len(siteToF[siteId]) != 0 {
			glog.Infoln("update profile firewall:", siteToF[siteId])
			newSpec = newSpec[0:len(newSpec)-1] + "," + "\"disabledProfileRuleIds\":" + sliceToString(siteToF[siteId]) + "}"
		}
	}
	//fmt.Println("new:", newSpec)
	s = strings.Replace(s, oldSpec, newSpec, 1)
	return s
}

/*
	func:sliceToString
	usage: transform slice to string
*/
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

/*
	func: generateLog
	usage: generate crontab log
*/
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
				idMap = checkMysqlSdwanBusiness(idMap)
				//fmt.Println("sdWAN FUNC B:", idMap)
				fIdMap = checkMysqlSdwanFirewall(fIdMap)
				//fmt.Println("sdwan func f:", fIdMap)
				updateMysqlEnableStatement(policyUpdateCmd, idMap)
				updateMysqlEnableStatement(firewallUpdateCmd, fIdMap)
				fmt.Println("update:", siteIdToPolicyBName[this_site_id], siteIdToFirewallBName[this_site_id])
				updateApiserver(siteIdToPolicyBName[this_site_id][0], siteIdToFirewallBName[this_site_id][0], this_site_id)
				pushLogToMysql(this_site_id, siteIdName[this_site_id], "Business Policy", siteIdToBusinessNamesMap[this_site_id])
				pushLogToMysql(this_site_id, siteIdName[this_site_id], "Firewall", siteIdToFirewallNamesMap[this_site_id])
			}
		}
	}
}

/*
	func: checkMysqlSdOnePolicy
	input:policy ruleId
	output: ifUseSdwan (bool)
*/
func checkMysqlSdOnePolicy(ruleId string) bool {
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName+"?charset=utf8&parseTime=True")
	checkErr(err)
	var id string
	var action string
	var source string
	var destination string
	var service string
	isSdwan := false
	command := "SELECT id,action,source,service,destination FROM cubs.profile_policy_rule WHERE id=" + ruleId + ";"
	row := db.QueryRow(command)
	err = row.Scan(&id, &action, &source, &service, &destination)
	//fmt.Println("service:", destination)
	if gjson.Get(action, "networkService.pathSelectClassParams").Bool() {
		isSdwan = true
	}
	if gjson.Get(action, "networkService.aggregation").Bool() {
		isSdwan = true
	}
	if gjson.Get(action, "networkService.forwardErrorCorrect").Bool() {
		isSdwan = true
	}
	if gjson.Get(service, "serviceType").String() == "appGroup" {
		isSdwan = true
	}
	//fmt.Println("mysql check:", gjson.Get(source, "selectType").String(), gjson.Get(source, "customAddrType").String())
	if gjson.Get(source, "selectType").String() == "custom" && gjson.Get(source, "customAddrType").String() == "country" {
		//fmt.Println("mysql check:", "update to true")
		isSdwan = true
	}
	if gjson.Get(destination, "selectType").String() == "custom" && gjson.Get(destination, "customAddrtype").String() == "country" {
		isSdwan = true
	}
	return isSdwan
}

/*
	func: checkMysqlSdOneFirewall
	input:firewall ruleId
	output: ifUseSdwan (bool)
*/
func checkMysqlSdOneFirewall(ruleId string) bool {

	var id string
	var action string
	var source string
	var destination string

	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName+"?charset=utf8&parseTime=True")
	checkErr(err)
	command := "SELECT id,action,source,destination FROM cubs.profile_firewall_rule WHERE id=" + ruleId + ";"
	row := db.QueryRow(command)
	err = row.Scan(&id, &action, &source, &destination)
	//fmt.Println("tag:", gjson.Get(action, "customBlockSites.length").String())
	isSdwan := false
	if gjson.Get(action, "isAllow").Bool() {
		if gjson.Get(action, "blockAppGroup.id").Exists() {
			isSdwan = true
		}
		if gjson.Get(action, "customBlockSites.length").Int() > 0 {
			isSdwan = true
		}
		if gjson.Get(action, "selectedBlockPages.length").Int() > 0 {
			isSdwan = true
		}
	}
	if gjson.Get(source, "selectType").String() == "custom" && gjson.Get(source, "customAddrType").String() == "country" {
		isSdwan = true
	}
	if gjson.Get(destination, "selectType").String() == "custom" && gjson.Get(destination, "customAddrType").String() == "country" {
		isSdwan = true
	}

	return isSdwan
}

/*
	func: checkMysqlSdwanBusiness
	input: mysql profile_policy_rule id slice
	output: profile_policy_rule id which use sdwan function slice
*/
func checkMysqlSdwanBusiness(m []string) []string {
	var newArr []string
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName+"?charset=utf8&parseTime=True")
	checkErr(err)
	for i := 0; i < len(m); i++ {
		var id string
		var action string
		var source string
		var destination string
		var service string
		isSdwan := false
		command := "SELECT id,action,source,service,destination FROM cubs.profile_policy_rule WHERE id=" + m[i] + ";"
		row := db.QueryRow(command)
		err = row.Scan(&id, &action, &source, &service, &destination)
		//fmt.Println("service:", destination)
		if gjson.Get(action, "networkService.pathSelectClassParams").Bool() {
			isSdwan = true
		}
		if gjson.Get(action, "networkService.aggregation").Bool() {
			isSdwan = true
		}
		if gjson.Get(action, "networkService.forwardErrorCorrect").Bool() {
			isSdwan = true
		}
		if gjson.Get(service, "serviceType").String() == "appGroup" {
			isSdwan = true
		}
		//fmt.Println("mysql check:", gjson.Get(source, "selectType").String(), gjson.Get(source, "customAddrType").String())
		if gjson.Get(source, "selectType").String() == "custom" && gjson.Get(source, "customAddrType").String() == "country" {
			//fmt.Println("mysql check:", "update to true")
			isSdwan = true
		}
		if gjson.Get(destination, "selectType").String() == "custom" && gjson.Get(destination, "customAddrtype").String() == "country" {
			isSdwan = true
		}
		if isSdwan {
			newArr = append(newArr, id)
			//fmt.Println("mysql check:", newArr)
		}
	}
	return newArr
}

/*
	func: checkMysqlSdwanFirewall
	input: mysql profile_firewall_rule id slice
	output: profile_firewall_rule id which use sdwan function slice
*/
func checkMysqlSdwanFirewall(m []string) []string {
	var newArr []string
	var id string
	var action string
	var source string
	var destination string
	for i := 0; i < len(m); i++ {
		db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName+"?charset=utf8&parseTime=True")
		checkErr(err)
		command := "SELECT id,action,source,destination FROM cubs.profile_firewall_rule WHERE id=" + m[i] + ";"
		row := db.QueryRow(command)
		err = row.Scan(&id, &action, &source, &destination)
		//fmt.Println("tag:", gjson.Get(action, "customBlockSites.length").String())
		isSdwan := false
		if gjson.Get(action, "isAllow").Bool() {
			if gjson.Get(action, "blockAppGroup.id").Exists() {
				isSdwan = true
			}
			if gjson.Get(action, "customBlockSites.length").Int() > 0 {
				isSdwan = true
			}
			if gjson.Get(action, "selectedBlockPages.length").Int() > 0 {
				isSdwan = true
			}
		}
		if gjson.Get(source, "selectType").String() == "custom" && gjson.Get(source, "customAddrType").String() == "country" {
			isSdwan = true
		}
		if gjson.Get(destination, "selectType").String() == "custom" && gjson.Get(destination, "customAddrType").String() == "country" {
			isSdwan = true
		}
		if isSdwan {
			newArr = append(newArr, id)
		}
	}
	return newArr
}

/*
	func: getSnExpiredQuery
	output: filter from_date and to_date, return contractId, last_expired_at *sql.Roes
*/

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
	return rows
}

/*
	func: pushLogToMysql
	input: site id,site name(gui),ruleType(policy or firewall),rules slice
	usage: put log ro gui
*/
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

}

/*
func: getMysqlSiteIdToSiteNameMap
input: site id slice
output: site id to gui site name map
*/
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

/*
func: getSiteIdToOrderedIdMaps
input: site id slice
output: 1)site id to policy profile order ids map(which uses sdwan function) 2)site id to firewall profile order ids map(which uses sdwan function)
*/
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
	//checkSdwanBusinessApi
	for key, value := range siteToP {
		//cmd = "SELECT orderId,policyId FROM cubs.profile_policy_rule WHERE ruleType='profile' && policyId="
		//cmd2 := "SELECT orderId,firewallId FROM cubs.profile_firewall_rule  WHERE ruleType='profile' && firewallId="
		cmd3 := "SELECT beName FROM cubs.profile_policy WHERE `profileId`="
		cmd4 := "SELECT beName FROM cubs.profile_firewall WHERE `profileId`="
		var thisBNm string
		var thisBNmF string
		row := db.QueryRow(cmd3 + value + ";")
		err := row.Scan(&thisBNm)
		row = db.QueryRow(cmd4 + value + ";")
		err = row.Scan(&thisBNmF)
		checkErr(err)
		//fmt.Println("search profile:", buisnessPolicyUrl, thisBNm)
		disableIds := checkSdwanBusinessApi(buisnessPolicyUrl + thisBNm)
		//fmt.Println("disabledIds:", disableIds)
		for i := 0; i < len(disableIds); i++ {
			siteToBOrders[key] = append(siteToBOrders[key], strconv.Itoa(disableIds[i]))

		}
		disableIds = checkSdwanFirewallApi(firewallPolicyUrl + thisBNmF)
		//fmt.Println("disable:", disableIds)
		for i := 0; i < len(disableIds); i++ {
			siteToFOrders[key] = append(siteToFOrders[key], strconv.Itoa(disableIds[i]))

		}

	}
	return siteToBOrders, siteToFOrders
}
func init() {
	flag.StringVar(&mysqlDomain, "mysqlDomain", "sdwan-orch-db-orchestrator-db:3306", "it's mysql domain")
	flag.StringVar(&username, "username", "root", "mysql login username")
	flag.StringVar(&password, "password", "root", "mysql login password")
	flag.StringVar(&apiserverDomain, "apiserverDomain", "http://sdwan-api-01-apiserver:80", "it's apiserver domain")
}
func main() {
	flag.Parse()
	defer glog.Flush()
	//mysqlDomain = os.Args[1]
	//username = os.Args[2]
	//password = os.Args[3]
	//apiserverDomain = os.Args[4]
	//from_date = os.Args[5]
	//to_date = os.Args[6]
	glog.Infoln("timestamp:", GetTaiwanTime2().Format(detailTime))
	glog.Infoln("mysql domain:", mysqlDomain)
	glog.Infoln("mysql username:", username)
	glog.Infoln("mysql password:", password)
	glog.Infoln("apiserver domain:", apiserverDomain)

	//fmt.Println("from date", from_date)
	//fmt.Println("to date:", to_date)

	snExpiredMap := getMysqlMap(getSnExpiredQuery())

	siteNameToSnMap, siteNameToSiteIdMap, snToSiteId, siteIds := getApiserverMap(apiserverDomain, snExpiredMap)
	siteToB, siteToF = getSiteIdToOrderedIdMaps(siteIds)
	siteIdName := getMysqlSiteIdToSiteNameMap(siteIds)
	//checkTableS(siteIdName)
	BpolicyIdToIdMap, siteIdToBusnessNamesMap := getMysqlProfilePolicyRule(snToSiteId)
	FpolicyIdToIdMap, siteIdToFirewallNamesMap := getMysqlFirewallRule(snToSiteId)
	siteIdToPolicyBName = getSiteIdToPolicyBName(snToSiteId)
	siteIdToFirewallBName = getSiteIdToFirewallBName(snToSiteId)

	//siteToProBBname, siteToProFBname = getSiteIdToProfileBname()
	checkDeviceLicense(snExpiredMap, siteNameToSnMap, siteNameToSiteIdMap, BpolicyIdToIdMap, FpolicyIdToIdMap, siteIdToPolicyBName, siteIdToFirewallBName, siteIdToBusnessNamesMap, siteIdToFirewallNamesMap, siteIdName)
	glog.Infoln("check license end!")
}
