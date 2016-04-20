package main

import (
	"github.com/isaacml/test/segcapt"
)

func main(){
	settings := map[string]string{
		"v_mode"	:	"2",
		"v_input"	:	"3",
		"a_input"	:	"2",
		"v_output"	:	"1", // SDTV
		"a_level"	:	"0", // 0 dB
		"aspect_ratio"	:	"16:9",
		"mac"		:	"d4ae52d3ea66",
		"ip_upload"	:	"192.168.4.22:9999",
	}
	
	seg := segcapt.SegmentCapturer("mac_","/var/segments/",settings)
	seg.Run(false)  
	
	for {} // infinito
	
}
