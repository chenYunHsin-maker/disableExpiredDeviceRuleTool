package main

import (
	"fmt"

	"github.com/robfig/cron"
)

func main() {
	c := cron.New()
	//"*/3 * * * *"
	//at minutes 0 (0 * * * *)
	//at min0,1hour do 1 time (0 */1 * * *)
	c.AddFunc("*/3 * * * * *", func() {
		fmt.Println("Hi! every 3 sec executing")
	})

	go c.Start()
	defer c.Stop()

	select {
	/*
		case <-time.After(time.Second * 10):
			return
	*/
	}
}
