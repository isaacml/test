package main

import (
	"fmt"
	"strings"
)

func main(){

line := "  X-Frame-Options: bytes=285008 filename=mac_2 md5sum=7cce5be2e91ed0eed5c1452538a2a2c6 fvideo=h264 faudio=heaacv1 hres=720 vres=576 numfps=25 denfps=0 vbitrate=2000 abitrate=128 block=prog next=prog duration=10 timestamp=1462803329  "

analyze(line)

}

func analyze(line string){
	var bytes, hres, vres, numfps, denfps, vbitrate, abitrate, duration, timestamp int
	var filename, md5sum, fvideo, faudio, block, next string
	ln := strings.Trim(line, " ")
	fmt.Sscanf(ln, "X-Frame-Options: bytes=%d filename=%s md5sum=%s fvideo=%s faudio=%s hres=%d vres=%d numfps=%d denfps=%d vbitrate=%d abitrate=%d block=%s next=%s duration=%d timestamp=%d", 
			   &bytes, &filename, &md5sum, &fvideo, &faudio, &hres, &vres, &numfps, &denfps, &vbitrate, &abitrate, &block, &next, &duration, &timestamp)
}