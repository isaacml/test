package main

import (
	"github.com/isaacml/test/segplay"
	"fmt"
	"time"
)

func main(){
	settings := map[string]string{
		"overscan"		:		"1",
		"x0"			:		"0",
		"y0"			:		"0",
		"x1"			:		"719",
		"y1"			:		"575",
		"vol"			:		"1",	
	}
	
	fmt.Println("...")
	seg := segplay.SegmentPlayer("","/root/",settings)
	seg.Run()
	time.Sleep(30 * time.Second)
	seg.Stop()
	for {}
}
