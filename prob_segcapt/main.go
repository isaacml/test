package main

import (
	"github.com/isaacml/test/segcapt"
	"time"
	"fmt"
)

func main(){
	settings := map[string]string{
		"fvideo"		:	"h264",
		"faudio"		:	"heaacv1",
		"abitrate"		:	"128",
		"tv_id"			:	"2",
		"v_mode"		:	"2",
		"v_input"		:	"3",
		"a_input"		:	"2",
		"v_output"		:	"1", // SDTV
		"a_level"		:	"0", // 0 dB
		"aspect_ratio"	:	"16:9",
		"mac"			:	"d4ae52d3ea66",
		"ip_upload"		:	"192.168.4.22:9999",
	}
	
	seg := segcapt.SegmentCapturer("mac_","/var/segments/",settings)
	seg.Run(false) // prog
	fmt.Printf("================== (PROGRAMA) =================")
	time.Sleep(95 * time.Second) 
	fmt.Printf("================== (PUBLICIDAD) =================")
	seg.CutSegment(true) // pub
	time.Sleep(95 * time.Second)
	seg.CutSegment(false) // prog
	fmt.Printf("================== (PROGRAMA) =================")
	
	for {} // infinito
	
}
