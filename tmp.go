package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
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
func main() {
	s := "{sdf}"
	s = s[1 : len(s)-1]
	fmt.Println(s)
	/*
		s := "{\"kind\":\"BPRuleSet\",\"apiVersion\":\"businesspolicy/v1alpha1\",\"metadata\":{\"name\":\"bp-xqsc8\",\"generateName\":\"bp-\",\"namespace\":\"default\",\"selfLink\":\"/apis/businesspolicy/v1alpha1/namespaces/default/bprulesets/bp-xqsc8\",\"uid\":\"dc289f92-f157-11e8-9730-0a58646004fc\",\"resourceVersion\":\"333638\",\"creationTimestamp\":\"2018-11-26T08:47:09Z\",\"labels\":{\"site-4fsnz\":\"c379572b-fec0-11e8-ab5d-0a5864600476\",\"site-dlt89\":\"996aa6dc-e0b8-11e8-b8ac-0a5864600d4c\",\"site-h8rbd\":\"99609de7-e0b8-11e8-b8ac-0a5864600d4c\"},\"ownerReferences\":[{\"apiVersion\":\"profile/v1alpha1\",\"kind\":\"ProfileConfig\",\"name\":\"profile-n6v72\",\"uid\":\"dc163433-f157-11e8-9730-0a58646004fc\",\"controller\":false,\"blockOwnerDeletion\":true}]},\"spec\":{\"policyRules\":[{\"ruleName\":\"P_from-100\",\"orderId\":1,\"source\":{\"selectType\":\"gateway\",\"selectIds\":[{}]},\"userGroup\":{\"name\":\"Any\"},\"destination\":{\"selectType\":\"siteScope\",\"selectIds\":[{\"siteName\":\"site-h8rbd\",\"zoneName\":\"zone-941pr\"},{\"siteName\":\"site-dlt89\",\"zoneName\":\"zone-941pr\"}]},\"service\":{\"serviceType\":\"any\"},\"dscp\":{\"dscpType\":\"any\",\"dscpValue\":-1},\"action\":{\"networkService\":{\"serviceType\":\"transport\",\"pathSelection\":[{\"orderId\":1,\"path\":\"VPN\"}],\"nat\":true},\"dscp\":{\"dscpType\":\"preserve\",\"dscpValue\":-1},\"priority\":\"high\"}},{\"ruleName\":\"P_from-100-2\",\"orderId\":2,\"source\":{\"selectType\":\"gateway\",\"selectIds\":[{}]},\"userGroup\":{\"name\":\"Any\"},\"destination\":{\"selectType\":\"siteScope\",\"selectIds\":[{\"siteName\":\"site-h8rbd\",\"zoneName\":\"zone-941pr\"},{\"siteName\":\"site-dlt89\",\"zoneName\":\"zone-941pr\"}]},\"service\":{\"serviceType\":\"any\"},\"dscp\":{\"dscpType\":\"any\",\"dscpValue\":-1},\"action\":{\"networkService\":{\"serviceType\":\"backhaul\",\"pathSelection\":[{\"orderId\":1,\"path\":\"VPN\"}],\"nat\":true,\"backhaulType\":\"zyxel\",\"backhaulZyxel\":[{\"name\":\"site-h8rbd\",\"orderId\":1}]},\"dscp\":{\"dscpType\":\"preserve\",\"dscpValue\":-1},\"priority\":\"high\"}},{\"ruleName\":\"P_to-100\",\"enabled\":true,\"orderId\":3,\"source\":{\"selectType\":\"any\"},\"userGroup\":{\"name\":\"Any\"},\"destination\":{\"selectType\":\"gateway\",\"selectIds\":[{}]},\"service\":{\"serviceType\":\"any\"},\"dscp\":{\"dscpType\":\"any\",\"dscpValue\":-1},\"action\":{\"networkService\":{\"serviceType\":\"backhaul\",\"pathSelection\":[{\"orderId\":1,\"path\":\"VPN\"},{\"orderId\":2,\"path\":\"MPLS\"}],\"nat\":true,\"backhaulType\":\"zyxel\",\"backhaulZyxel\":[{\"name\":\"site-h8rbd\",\"orderId\":1}]},\"dscp\":{\"dscpType\":\"preserve\",\"dscpValue\":-1},\"priority\":\"high\"}},{\"ruleName\":\"P_via-hub\",\"enabled\":true,\"orderId\":4,\"source\":{\"selectType\":\"siteScope\",\"selectIds\":[{\"siteName\":\"site-h8rbd\",\"zoneName\":\"zone-941pr\"},{\"siteName\":\"site-dlt89\",\"zoneName\":\"zone-941pr\"}]},\"userGroup\":{\"name\":\"Any\"},\"destination\":{\"selectType\":\"any\"},\"service\":{\"serviceType\":\"any\"},\"dscp\":{\"dscpType\":\"any\",\"dscpValue\":-1},\"action\":{\"networkService\":{\"serviceType\":\"backhaul\",\"pathSelection\":[{\"orderId\":1,\"path\":\"VPN\"},{\"orderId\":2,\"path\":\"MPLS\"}],\"nat\":true,\"backhaulType\":\"zyxel\",\"backhaulZyxel\":[{\"name\":\"site-h8rbd\",\"orderId\":1}]},\"dscp\":{\"dscpType\":\"preserve\",\"dscpValue\":-1},\"priority\":\"high\"}}]},\"status\":{}}"
		val := gjson.Get(s, "spec")
		fmt.Println(val)*/
	/*
		m := make(map[string][]string)
		m["cat"] = append(m["cat"], "1")
		m["cat"] = append(m["cat"], "2")
		m["cat"] = append(m["cat"], "3")

		fmt.Println(sliceToString(m["cat"]))*/

}
