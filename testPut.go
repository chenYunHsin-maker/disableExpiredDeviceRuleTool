package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
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
	fmt.Println("replace ", i, "strings")
	return str
}
func main() {
	targetUrl := "http://127.0.0.1:8080/apis/businesspolicy/v1alpha1/namespaces/default/bprulesets/bp-6dn14"
	//targetUrl := "http://127.0.0.1:8080/apis/firewall/v1alpha1/namespaces/default/fwpolicies/fw-4qt7z"
	//fw-4qt7z
	//bp-6dn14
	resp, err := http.Get(targetUrl)
	checkErr(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	checkErr(err)
	str := string(body[:])
	str = updateJson(str)
	putRequest(targetUrl, strings.NewReader(str))
}
