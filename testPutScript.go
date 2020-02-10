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
	for strings.Index(str, "\"enabled\":false,") != -1 {
		i++
		str = strings.Replace(str, "\"enabled\":false,", "\"enabled\":true,", 1)
	}
	fmt.Println("replace ", i, "strings")
	return str
}
func main() {
	targetUrl := "http://127.0.0.1:8080/apis/businesspolicy/v1alpha1/namespaces/default/bprulesets/bp-6dn14"
	targetUrl2 := "http://127.0.0.1:8080/apis/firewall/v1alpha1/namespaces/default/fwpolicies/fw-4qt7z"
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

	resp, err = http.Get(targetUrl2)
	checkErr(err)
	defer resp.Body.Close()
	body2, err := ioutil.ReadAll(resp.Body)
	checkErr(err)
	str2 := string(body2[:])
	str2 = updateJson(str2)
	putRequest(targetUrl2, strings.NewReader(str2))
}
