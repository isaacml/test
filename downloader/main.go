package main

import (
	"fmt"
	"github.com/isaacml/cmdline"
	"bufio"
	"time"
	"strings"
	"runtime"
	"sync"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"os/exec"
)

const (
	// variables de configuracion del servidor HTTP
	DirDB     = "/usr/local/bin/download.db"
	rootdir   = "/var/segments/"
)

var settings = make(map[string]string)
// var internas de SegPlay
var semaforo, lastdownload string
var downloaded, downloadedok bool
var	db      *sql.DB
var	db_mu	sync.Mutex

var g_bytes, g_hres, g_vres, g_numfps, g_denfps, g_vbitrate, g_abitrate, g_duration, g_timestamp int
var g_filename, g_md5sum, g_fvideo, g_faudio, g_block, g_next string


func init(){
	var err_db error
	db, err_db = sql.Open("sqlite3", DirDB)
	if err_db != nil {
		log.Fatalln(err_db)
	}
}

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

// /usr/bin/wget -S -O download.ts --post-data 'tv_id=2&mac=d4ae52d3ea66&semaforo=G&downloaded=mac_2.ts&bytes=285008&md5sum=7cce5be2e91ed0eed5c1452538a2a290' http://192.168.4.22:9999/download.cgi
// X-Frame-Options: bytes=285008 filename=mac_2 md5sum=7cce5be2e91ed0eed5c1452538a2a2c6 fvideo=h264 faudio=heaacv1 hres=720 vres=576 numfps=25 
//					denfps=0 vbitrate=2000 abitrate=128 block=prog next=prog duration=10 timestamp=1462803329
func download(){
	var lineacomandos string
	connected := false // si ha conectado con el servidor
	// consultamos la BD para ver todos los datos de la ultima bajada
	query, err := db.Query("SELECT * FROM `SEGMENTOS` WHERE SEGMENTOS.timestamp = (SELECT MAX(SEGMENTOS.timestamp) FROM SEGMENTOS)")
	if err != nil {
		log.Println(err)
	}
	for query.Next(){
		var bytes, hres, vres, numfps, denfps, vbitrate, abitrate, duration, timestamp, last_connect, tv_id int
		var filename, md5sum, fvideo, faudio, block, next, semaforo, mac string
		err = query.Scan(&filename, &bytes, &md5sum, &fvideo, &faudio, &hres, &vres, &numfps, &denfps, &vbitrate, &abitrate, &block, &next, &semaforo, &duration, &timestamp, &mac, &last_connect, &tv_id)
		if err != nil {
			log.Println(err)
		}
		lastdownload = filename+".ts"
		lineacomandos = fmt.Sprintf("/usr/bin/wget -S -O %sdownload.ts --post-data 'tv_id=%s&mac=%s&semaforo=%s&downloaded=%s&bytes=%d&md5sum=%s' http://192.168.4.22:9999/download.cgi",
										rootdir,settings["tv_id"],settings["mac"],semaforo,lastdownload,bytes,md5sum)
	}	
	// construimos la linea de comandos
	exe := cmdline.Cmdline(lineacomandos)
	lectura, err:= exe.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}
	mReader := bufio.NewReader(lectura)
	downloaded, downloadedok = false, false
	time_semaforo := time.Now()
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
			downloaded = true
		}
		if strings.Contains(line, "X-Frame-Options:") {
			connected = true
			if downloaded {
				line = strings.Trim(line, " ")
				fmt.Sscanf(line, "X-Frame-Options: bytes=%d filename=%s md5sum=%s fvideo=%s faudio=%s hres=%d vres=%d numfps=%d denfps=%d vbitrate=%d abitrate=%d block=%s next=%s duration=%d timestamp=%d", 
								&g_bytes, &g_filename, &g_md5sum, &g_fvideo, &g_faudio, &g_hres, &g_vres, &g_numfps, &g_denfps, &g_vbitrate, &g_abitrate, &g_block, &g_next, &g_duration, &g_timestamp)
			}else{ // X-Frame-Options: already downloaded ; X-Frame-Options: access not granted
				
			}
		}
		fmt.Printf("[wget] %s\n", line)
		runtime.Gosched()
	}
	exe.Stop()
	dur_semaforo := time.Since(time_semaforo).Seconds()
	if downloaded {
		// comprobar que el fichero se ha bajado correctamente
		fileinfo, err := os.Stat(rootdir+"download.ts") // fileinfo.Size()
		if err != nil {
			downloadedok = false
			fmt.Println(err)
		}
		filesize := fileinfo.Size()
		md5:= md5sumfunc(rootdir+"download.ts")
		if filesize == int64(g_bytes) && md5 == g_md5sum {
			downloadedok = true	
		}else{
			downloadedok = false
		}
	}
	// decidir el color del semaforo
	if downloadedok {
		var color float64
		color = float64(dur_semaforo) / float64(g_duration) // duration=10
		switch {
			case color > 1.2:
				semaforo = "R"
			case color < 0.8:
				semaforo = "G"
			default:
				semaforo = "Y"
		}
	}
	if !connected {
		semaforo = "R"
	}
	// grabamos los datos del nuevo fichero downloaded en la BD
	if downloadedok{
		err = exec.Command("/bin/sh","-c","mv -f "+rootdir+"download.ts"+" "+rootdir+lastdownload).Run()
		if err != nil {
			log.Println(err)
		}
		db_mu.Lock()
		_, err = db.Exec("INSERT INTO segmentos (filename,bytes,md5sum,fvideo,faudio,hres,vres,num_fps,den_fps,vbitrate,abitrate,block,next,duration,timestamp,mac,last_connect,semaforo,tv_id) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", g_filename, g_bytes, g_md5sum,
			g_fvideo, g_faudio, g_hres, g_vres, g_numfps, g_denfps, g_vbitrate, g_abitrate, g_block, g_next, g_duration, g_timestamp, "-", g_timestamp, semaforo, 0)
		db_mu.Unlock()
	}
		
}

// equivalent to md5sum -b filename
func md5sumfunc(filename string) string {
	out, _ := exec.Command("/bin/sh", "-c", "md5sum -b "+filename+" | awk '{print $1}'").CombinedOutput()

	return strings.TrimSpace(string(out))
}

/*
  HTTP/1.1 404 Not Found
  Content-Type: text/plain; charset=utf-8
  X-Content-Type-Options: nosniff
  X-Frame-Options: Already downloaded
  X-Frame-Options: access not granted  
  Date: Mon, 09 May 2016 14:34:46 GMT
  Content-Length: 19

  HTTP/1.1 200 OK
  X-Frame-Options: bytes=285008 filename=mac_2 md5sum=7cce5be2e91ed0eed5c1452538a2a2c6 fvideo=h264 faudio=heaacv1 hres=720 vres=576 numfps=25 denfps=0 vbitrate=2000 abitrate=128 block=prog next=prog duration=10 timestamp=1462803329
  Date: Mon, 09 May 2016 14:59:17 GMT
  Content-Type: application/octet-stream
  Transfer-Encoding: chunked

*/