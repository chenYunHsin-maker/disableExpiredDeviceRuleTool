package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/syhlion/sqlwrapper"
)

const (
	buisnessPolicyUrl = "/apis/businesspolicy/v1alpha1/namespaces/default/bprulesets/"
	firewallPolicyUrl = "/apis/firewall/v1alpha1/namespaces/default/fwpolicies/"
	siteconfigUrl     = "/apis/site/v1alpha1/namespaces/default/siteconfigs"
	timeFormat        = "2006-01-02"
)

var (
	dbName          = "cubs"
	mysqlDomain     = "127.0.0.1:3308"
	apiserverDomain = "127.0.0.1:8080"
	username        = "root"
	password        = "root"
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
	resp, err := http.Get("http://" + apiserverDomain + siteUrl)
	checkErr(err)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	s := string(body)
	//fmt.Println(s)
	return s
}
func getApiserverMap(apiserverDomain string) (map[string]string, map[string]string) {
	apiServerMap := make(map[string]string)
	siteNameToSiteIdMap := make(map[string]string)

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
			siteNameToSiteIdMap[this_siteName] = this_siteId
			apiServerMap[this_siteName] = this_sn

			//fmt.Println("device2 ", this_device2Sn)
			//apiServerDevice2Map[this_siteName] = append(apiServerDevice2Map[this_siteName], this_device2Sn)
		}
	}
	return apiServerMap, siteNameToSiteIdMap
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
		var expired sql.NullString
		var sn sql.NullString
		if err := rows.Scan(&sn, &expired); err != nil {
			fmt.Println(" err :", err)
		}
		if sn.Valid == true {
			if sn.String != "" {
				snExpiredMap[sn.String] = expired.String
			}

			//fmt.Println("sn:", sn.String, " linked site id:", site.String)
		}
	}
	return snExpiredMap
}
func getMysqlProfilePolicyRule(rows *sql.Rows) map[string][]string {
	//policyIdToRuleTypeMap := make(map[string][]string)

	//policyIdToRuleNameMap := make(map[string][]string)
	siteIdToIdMap := make(map[string][]string)
	//ruleName,ruleType,enabled,policyId
	for rows.Next() {
		var id sql.NullString
		var ruleName sql.NullString
		var ruleType sql.NullString
		var enabled sql.NullString
		var policyId sql.NullString
		if err := rows.Scan(&id, &ruleName, &ruleType, &enabled, &policyId); err != nil {
			fmt.Println(" err :", err)
		}

		if ruleType.String == "site" {
			//idToRuleNameMap

			//policyIdToRuleNameMap[policyId.String] = append(policyIdToRuleNameMap[policyId.String], ruleName.String)
			siteIdToIdMap[policyId.String] = append(siteIdToIdMap[policyId.String], id.String)
		}

		//fmt.Println("sn:", sn.String, " linked site id:", site.String)

	}
	return siteIdToIdMap
}
func initDb() {

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
}
func checkDeviceLicense(snExpiredMap, siteNameToSnMap, siteNameToSiteIdMap map[string]string, policyIdToIdMap map[string][]string) {
	for key, _ := range siteNameToSnMap {
		this_site_name := key
		this_sn := siteNameToSnMap[this_site_name]
		today := GetTaiwanTime()
		expired_date, _ := ShortDateFromString(snExpiredMap[this_sn])

		if expired_date.After(today) {
			fmt.Println("license is working!")
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
func main() {
	initDb()
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName)
	checkErr(err)
	rows, _ := db.Query("SELECT serial,new_expired FROM cubs.license_key;")
	rows2, _ := db.Query("SELECT  id,ruleName,ruleType,enabled,policyId FROM cubs.profile_policy_rule;")
	defer rows.Close()
	defer rows2.Close()

	snExpiredMap := getMysqlMap(rows)
	policyIdToIdMap := getMysqlProfilePolicyRule(rows2)
	siteNameToSnMap, siteNameToSiteIdMap := getApiserverMap(apiserverDomain)
	checkDeviceLicense(snExpiredMap, siteNameToSnMap, siteNameToSiteIdMap, policyIdToIdMap)

}
