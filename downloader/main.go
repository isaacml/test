package main

import (
	"fmt"
	"github.com/isaacml/cmdline"
	"bufio"
	"time"
	"strings"
	"runtime"
)

var settings = make(map[string]string)
// var internas de SegPlay
var semaforo, lastdownload string
var downloadedok bool

func main() {
	settings = map[string]string{
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
	semaforo = "G"
	lastdownload = "mac_2.ts"
	download()
}

// /usr/bin/wget -S -O test.ts --post-data "tv_id=2&mac=d4ae52d3ea66&semaforo=G&downloaded=mac_2.ts&bytes=285008&md5sum=7cce5be2e91ed0eed5c1452538a2a290" http://192.168.4.22:9999/download.cgi
// X-Frame-Options: bytes=285008 filename=mac_2 md5sum=7cce5be2e91ed0eed5c1452538a2a2c6 fvideo=h264 faudio=heaacv1 hres=720 vres=576 numfps=25 
//					denfps=0 vbitrate=2000 abitrate=128 block=prog next=prog duration=10 timestamp=1462803329
func download(){
	lineacomandos := fmt.Sprintf("/usr/bin/wget -S -O download.ts --post-data 'tv_id=%s&mac=%s&semaforo=%s&downloaded=%s&bytes=285008&md5sum=7cce5be2e91ed0eed5c1452538a2a290' http://192.168.4.22:9999/download.cgi",
									settings["tv_id"],settings["mac"],semaforo,lastdownload)
	exe := cmdline.Cmdline(lineacomandos)
	lectura, errL := exe.StdoutPipe()
	if errL != nil {
		fmt.Println(errL)
	}
	mReader := bufio.NewReader(lectura)
	time_semaforo := time.Now()
	downloadedok = false
	exe.Start()
	for { // bucle de reproduccion normal
		line, err := mReader.ReadString('\n')
		if err != nil {
			fmt.Println("Fin del wget !!!")
			break
		}
		line = strings.TrimRight(line, "\n")
		if strings.Contains(line, "HTTP/1.1 200 OK") {
			fmt.Println("Uploaded OK")
			downloadedok = true
		}
		if strings.Contains(line, "X-Frame-Options:") && downloadedok {
			analyze(line)
		}
		fmt.Printf("[curl] %s\n", line)
		runtime.Gosched()
	}
	exe.Stop()
	dur_semaforo := time.Since(time_semaforo).Seconds()
	// decidir el color del semaforo
	var color float64
	color = float64(dur_semaforo) / float64(10) // duration=10
	switch {
		case color > 1.2:
			semaforo = "R"
		case color < 0.8:
			semaforo = "G"
		default:
			semaforo = "Y"
	}
	
		
}

func analyze(line string){
	
}

/*
  HTTP/1.1 404 Not Found
  Content-Type: text/plain; charset=utf-8
  X-Content-Type-Options: nosniff
  X-Frame-Options: Already downloaded
  Date: Mon, 09 May 2016 14:34:46 GMT
  Content-Length: 19

  HTTP/1.1 200 OK
  X-Frame-Options: bytes=285008 filename=mac_2 md5sum=7cce5be2e91ed0eed5c1452538a2a2c6 fvideo=h264 faudio=heaacv1 hres=720 vres=576 numfps=25 denfps=0 vbitrate=2000 abitrate=128 block=prog next=prog duration=10 timestamp=1462803329
  Date: Mon, 09 May 2016 14:59:17 GMT
  Content-Type: application/octet-stream
  Transfer-Encoding: chunked

*/