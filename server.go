package main

import (
	"net/http"

	"github.com/labstack/echo"
)

type Cronjob struct {
	//原来golang对变量是否包外可访问，是通过变量名的首字母是否大小写来决定的
	Name string `json:"name" form:"name" query:"name"`
	Freq string `json:"freq" form:"freq" query:"freq"`
	Cmd  string `json:"cmd" form:"cmd" query:"cmd"`
}

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.POST("/add", func(c echo.Context) error {

		job := new(Cronjob)
		//fmt.Println(c.Request().Body)
		if err := c.Bind(job); err != nil {
			return err
		}
		//fmt.Println(job)
		//resultStr := name + "_" + freq + "_" + cmd
		return c.JSON(http.StatusOK, job)
	})
	e.Logger.Fatal(e.Start(":3143"))
}
