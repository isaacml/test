package main

import (
	"github.com/isaacml/segplay"
	"fmt"
	"time"
	"runtime"
)

func main(){
	settings := map[string]string{
		"fvideo"		:	"h264", // 3 parametros fijos de codificacion audio/video
		"faudio"		:	"heaacv1",
		"abitrate"		:	"128",

		"overscan"		:		"0",
		"x0"			:		"0",
		"y0"			:		"0",
		"x1"			:		"719",
		"y1"			:		"575",
		"vol"			:		"0",

		"aspect_ratio"	:	"16:9",
		"mac"			:	"d4ae52d3ea66", // la recogemos con bmdinfo()

		"tv_id"			:	"2", // nos lo dará en su momento el server de gestion
		"ip_download"		:	"192.168.4.22:9999", // nos lo dará en su momento el server de gestion
	}
	
	fmt.Println("...")
	seg := segplay.SegmentPlayer("","/var/segments/",settings)
	seg.Run() // 3 gorutinas
	for {
		runtime.Gosched()
		time.Sleep(1 * time.Second)
	}
}

