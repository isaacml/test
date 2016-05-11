package main

import (
	"github.com/isaacml/test/segplay"
	"fmt"
	"time"
	"runtime"
)

func main(){
	settings := map[string]string{
		"overscan"		:		"0",
		"x0"			:		"0",
		"y0"			:		"0",
		"x1"			:		"719",
		"y1"			:		"575",
		"vol"			:		"0",	
	}
	
	fmt.Println("...")
	seg := segplay.SegmentPlayer("","/root/",settings)
	seg.Run() // 3 gorutinas
	for {
		runtime.Gosched()
		time.Sleep(1 * time.Second)
	}
}

