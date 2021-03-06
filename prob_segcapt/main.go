package main

import (
	"github.com/isaacml/segcapt"
	"time"
	"fmt"
)

func main(){
	settings := map[string]string{
		"fvideo"		:	"h264", // 3 parametros fijos de codificacion audio/video
		"faudio"		:	"heaacv1",
		"abitrate"		:	"128",

		"v_mode"		:	"2",
		"v_input"		:	"3",
		"a_input"		:	"2",
		"v_output"		:	"1", // SDTV
		"a_level"		:	"0", // 0 dB
		"aspect_ratio"	:	"16:9",
		"mac"			:	"d4ae52d3ea66", // la recogemos con bmdinfo()

		"tv_id"			:	"2", // nos lo dará en su momento el server de gestion
		"ip_upload"		:	"192.168.4.22:9999", // nos lo dará en su momento el server de gestion
	}
	
	seg := segcapt.SegmentCapturer("mac_","/var/segments/",settings)
	seg.Run(false) // prog
	fmt.Println("================== (PROGRAMA) =================")
	time.Sleep(55 * time.Second) 
	fmt.Println("================== (PUBLICIDAD) =================")
	seg.CutSegment(true) // pub
	time.Sleep(55 * time.Second)
	seg.CutSegment(false) // prog
	fmt.Println("================== (PROGRAMA) =================")
	time.Sleep(15 * time.Second)
	seg.Stop()
	
//	for {} // infinito
	
}
