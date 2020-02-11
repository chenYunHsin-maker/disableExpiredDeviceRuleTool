package main

import (
	"fmt"

	"github.com/robfig/cron"
)

func main() {
	c := cron.New()
	//"*/3 * * * *"
	c.AddFunc("0 * * * *", func() {
		fmt.Println("Hi! every 1 hour executing")
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
