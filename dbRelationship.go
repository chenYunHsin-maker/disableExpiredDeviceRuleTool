Mysql:
	site_firewall :
		enabled string//to decide if this rule is enabled or not
		beName string//apiserver url argument ex:http://127.0.0.1:8080/apis/businesspolicy/v1alpha1/namespaces/default/bprulesets/{beName}
	site_policy:
		siteId string//site id
		beName string //apiserver url argument ex:http://127.0.0.1:8080/apis/businesspolicy/v1alpha1/namespaces/default/bprulesets/{beName}
	profile_policy_rule:
		source struct {
			siteId
		}

