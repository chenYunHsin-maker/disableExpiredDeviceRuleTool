package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/syhlion/sqlwrapper"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	username, password := "root", "root"
	mysqlDomain := "127.0.0.1:3308"
	dbName := "cubs"
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+mysqlDomain+")/"+dbName+"?charset=utf8&parseTime=True")
	checkErr(err)
	command := "SELECT * FROM cubs.auth_useractivitylog WHERE userName='disable expired device cronJob';"
	rows, _ := db.Query(command)
	defer rows.Close()
	var recordId sql.NullString
	var userId sql.NullString
	var userName sql.NullString
	var requestTime sql.NullString
	var layerId sql.NullString
	var layerName sql.NullString
	var modulePage sql.NullString
	var feature sql.NullString
	var requestDest sql.NullString
	var requestBody sql.NullString
	var oldValue sql.NullString
	var newValue sql.NullString

	for rows.Next() {
		if err := rows.Scan(&recordId, &userId, &userName, &requestTime, &layerId, &layerName, &modulePage, &feature, &requestDest, &requestBody, &oldValue, &newValue); err != nil {
			fmt.Println(" err :", err)
		}
		fmt.Println("recordId: ", recordId.String)
		fmt.Println("userId: ", userId.String)
		fmt.Println("userName:", userName.String)
		fmt.Println("requestTIme:", requestTime.String)
		fmt.Println("layerId:", layerId.String)
		fmt.Println("layerName:", layerName.String)
		fmt.Println("modulePage:", modulePage.String)
		fmt.Println("feature:", feature.String)
		fmt.Println("requestDest:", requestDest.String)
		fmt.Println("requestBody:", requestBody.String)
		fmt.Println("oldValue:", oldValue.String)
		fmt.Println("newValue:", newValue.String)
	}

	get_max_cmd := "SELECT MAX(recordId) FROM cubs.auth_useractivitylog;"
	rows, _ = db.Query(get_max_cmd)
	for rows.Next() {
		var max_id sql.NullString
		if err := rows.Scan(&max_id); err != nil {
			fmt.Println(" err :", err)
		}
		fmt.Println("maxId:", max_id.String)
	}
}
