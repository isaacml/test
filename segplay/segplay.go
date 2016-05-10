package segplay

import (
	"bufio"
	"fmt"
	"github.com/isaacml/cmdline"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"io"
)

const (
	DirDB     = "/root/download.db"
)

var	(
	db      *sql.DB
	db_mu	sync.Mutex
)

func init(){
	var err_db error
	db, err_db = sql.Open("sqlite3", DirDB)
	if err_db != nil {
		log.Fatalln(err_db)
	}
}

type SegPlay struct {
	cmdomx                           string
	exe                              *cmdline.Exec
	exe2                             *cmdline.Exec
	mediawriter                      *bufio.Writer     // por aqui puedo enviar caracteres al omxplayer
	settings                         map[string]string // read-only map
	downloaddir                      string            // directorio RAMdisk donde se guardan los ficheros bajados del server y listos para reproducir
	pubdir                           string            // directorio del HD donde se guardan los ficheros de publicidad locales
	playing                          bool              // omxplayer esta reproduciendo
	restamping                       bool              // ffmpeg esta reestampando
	downloading                      bool              // esta bajando segmentos
	running                          bool              // proceso completo funcionando
	lastplay, nextplay, lastdownload string            // nombre completo de los ficheros
	lastplay_pub, nextplay_pub       bool
	lastplay_timestamp               int64
	semaforo                         string // R(red), Y(yellow), G(green) download speed
	volume                           int    // dB
	mu_seg                           sync.Mutex
}

func SegmentPlayer(pubdir, downloaddir string, settings map[string]string) *SegPlay {
	seg := &SegPlay{}
	seg.mu_seg.Lock()
	defer seg.mu_seg.Unlock()
	seg.settings = settings
	seg.downloaddir = downloaddir
	seg.pubdir = pubdir
	seg.playing = false
	seg.restamping = false
	seg.downloading = false
	seg.running = false
	seg.lastplay = ""
	seg.nextplay = ""
	seg.lastdownload = ""
	seg.lastplay_pub = false
	seg.nextplay_pub = false
	seg.lastplay_timestamp = 0
	seg.semaforo = "G" // comenzamos en verde

	return seg
}

func (s *SegPlay) Run() error {
	var err error
	ch := make(chan int)

	s.mu_seg.Lock()
	if s.running { // ya esta corriendo
		s.mu_seg.Unlock()
		return fmt.Errorf("segplay: ALREADY_RUNNING_ERROR")
	}
	s.running = true // comienza a correr
	s.mu_seg.Unlock()

	go s.command1(ch)
	go s.command2(ch)
	go s.downloader() // bajando a su bola sin parar lo que el director le diga de donde bajarlo (tv_id, mac, ip_download)
	go s.director() // envia segmentos al secuenciador cuando s.playing && s.restamping

	return err
}

func (s *SegPlay) Stop() error {
	var err error

	s.mu_seg.Lock()
	defer s.mu_seg.Unlock()
	if !s.running {
		return fmt.Errorf("segcapt: ALREADY_STOPPED_ERROR")
	}
	s.running = false
	killall("omxplayer omxplayer.bin ffmpeg")
	err = s.exe.Stop()
	if err != nil {
		err = fmt.Errorf("segcapt: STOP_ERROR")
	}

	return err
}

func (s *SegPlay) command1(ch chan int) { // omxplayer
	var tiempo int64
	for {
		var overscan string
		s.mu_seg.Lock()
		if s.settings["overscan"] == "1" {
			overscan = fmt.Sprintf(" --win %s,%s,%s,%s", s.settings["x0"], s.settings["y0"], s.settings["x1"], s.settings["y1"])
		}
		vol := toInt(s.settings["vol"])
		// creamos el cmdomx
		// /usr/bin/omxplayer -s -o both --vol 100 --hw --win '0 0 719 575' --no-osd -b /tmp/fifo2
		s.cmdomx = fmt.Sprintf("/usr/bin/omxplayer -s -o both --vol %d --hw%s --layer 100 --no-osd -b /tmp/fifo2", 100*vol, overscan)
		s.exe = cmdline.Cmdline(s.cmdomx)
		lectura, err := s.exe.StderrPipe()
		if err != nil {
			fmt.Println(err)
		}
		mReader := bufio.NewReader(lectura)

		stdinWrite, err := s.exe.StdinPipe()
		if err != nil {
			fmt.Println(err)
		}
		s.mediawriter = bufio.NewWriter(stdinWrite)
		s.mu_seg.Unlock()
		tiempo = time.Now().Unix()
		go func() {
			for {
				if (time.Now().Unix() - tiempo) > 10 {
					s.mu_seg.Lock()
					s.restamping = false
					s.playing = false
					s.mu_seg.Unlock()
					killall("omxplayer omxplayer.bin ffmpeg")
					break
				}
				time.Sleep(1 * time.Second)
			}
		}()
		s.exe.Start()

		for { // bucle de reproduccion normal
			tiempo = time.Now().Unix() //; time.Sleep(5 * time.Second)
			line, err := mReader.ReadString('\n')
			if err != nil {
				s.mu_seg.Lock()
				s.playing = false
				s.mu_seg.Unlock()
				fmt.Println("Fin del omxplayer !!!")
				break
			}
			line = strings.TrimRight(line, "\n")
			if strings.Contains(line, "Comenzando...") {
				fmt.Println("[omx]", "Ready...")
				ch <- 1
				s.mu_seg.Lock()
				s.playing = true
				s.mu_seg.Unlock()
			}
			if strings.Contains(line, "Current Volume:") { // Current Volume: -2 => "Current Volume: %d\n"
				var currvol int
				fmt.Sscanf(line,"Current Volume: %d",&currvol)
				s.mu_seg.Lock()
				s.settings["vol"] = fmt.Sprintf("%d",currvol)
				s.volume = currvol
				s.mu_seg.Unlock()
			}
			if strings.Contains(line, "Time:") {
				fmt.Printf("[omx] %s\n", line)
			}
			runtime.Gosched()
		}
		killall("omxplayer omxplayer.bin ffmpeg")
		s.exe.Stop()
		s.mu_seg.Lock()
		if !s.running {
			s.mu_seg.Unlock()
			break
		}
		s.mu_seg.Unlock()
		ch <- 1
	}
}

func (s *SegPlay) command2(ch chan int) { // ffmpeg
	var tiempo int64
	for {
		s.exe2 = cmdline.Cmdline("/usr/bin/ffmpeg -y -f mpegts -re -i /tmp/fifo1 -f mpegts -acodec copy -vcodec copy /tmp/fifo2")
		lectura, err := s.exe2.StderrPipe()
		if err != nil {
			fmt.Println(err)
		}
		mReader := bufio.NewReader(lectura)
		tiempo = time.Now().Unix()
		go func() {
			for {
				if (time.Now().Unix() - tiempo) > 5 {
					s.mu_seg.Lock()
					s.restamping = false
					s.playing = false
					s.mu_seg.Unlock()
					killall("omxplayer omxplayer.bin ffmpeg")
					break
				}
				time.Sleep(1 * time.Second)
			}
		}()
		<-ch
		s.exe2.Start()

		for { // bucle de reproduccion normal
			tiempo = time.Now().Unix() //; time.Sleep(5 * time.Second)
			line, err := mReader.ReadString('\n')
			if err != nil {
				s.mu_seg.Lock()
				s.restamping = false
				s.mu_seg.Unlock()
				fmt.Println("Fin del ffmpeg !!!")
				break
			}
			line = strings.TrimRight(line, "\n")
			if strings.Contains(line, "libpostproc") {
				fmt.Println("[ffmpeg]", "Ready...")
				s.mu_seg.Lock()
				s.restamping = true
				s.mu_seg.Unlock()
			}
			if strings.Contains(line, "frame=") {
				fmt.Printf("[ffmpeg] %s\n", line)
			}
			runtime.Gosched()
		}
		killall("omxplayer omxplayer.bin ffmpeg")
		s.exe2.Stop()
		s.mu_seg.Lock()
		if !s.running {
			s.mu_seg.Unlock()
			break
		}
		s.mu_seg.Unlock()
		<-ch
	}
}

// esta funcion envia los ficheros a reproducir a la cola de reproducción en el FIFO1 /tmp/fifo1
// secuencia /tmp/fifo1
func (s *SegPlay) secuenciador(file string) {

	fw, err := os.OpenFile("/tmp/fifo1", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	defer fw.Close()

	fr, err := os.Open(file) // read-only
	if err != nil {
		log.Fatalln(err)
	}
	if n, err := io.Copy(fw, fr); err == nil {
		fmt.Printf("Copiados %d bytes\n", n)
	} else {
		log.Println(err) // no salimos en caso de error de copia
	}
	fr.Close()
	
}

// el director ahora mismo solo le dirá al secuenciador que ficheros enviar a la cola de reproduccion
func (s *SegPlay) director() {
	for {
		if s.playing && s.restamping {
			fmt.Println("Preparado para recibir segmentos por el /tmp/fifo1")
			break
		}
		runtime.Gosched()
	}
	
	s.secuenciador("segment2.ts") // ejemplo: bloquea hasta que acaba de reproducir el fichero
	

}

// downloader en un futuro dependerá del valor del server s.settings["ip_download"] y por tanto del servidor de gestion, además del playout que le indiqie el director bajar
// debe añadirse el código q recoge las variables lastplay_pub, nextplay_pub importantes para el director playout
func (s *SegPlay) downloader() {
	var downloaded, downloadedok bool
	var g_bytes, g_hres, g_vres, g_numfps, g_denfps, g_vbitrate, g_abitrate, g_duration, g_timestamp int
	var g_filename, g_md5sum, g_fvideo, g_faudio, g_block, g_next string
	
	s.mu_seg.Lock()
	rootdir := s.downloaddir
	s.mu_seg.Unlock()
	
	contador := 0 // internamente en for va de 1 a 12 y cicla

	for {
		contador++
		if contador > 12 { contador = 1 }
		var lineacomandos string
		connected := false // si ha conectado con el servidor
		// consultamos la BD para ver todos los datos de la ultima bajada
		query, err := db.Query("SELECT * FROM `SEGMENTOS` WHERE SEGMENTOS.timestamp = (SELECT MAX(SEGMENTOS.timestamp) FROM SEGMENTOS)")
		if err != nil {
			log.Println(err)
		}
		var bytes, hres, vres, numfps, denfps, vbitrate, abitrate, duration, timestamp, last_connect, tv_id int
		var filename, md5sum, fvideo, faudio, block, next, semaforo, mac string
		for query.Next(){
			err = query.Scan(&filename, &bytes, &md5sum, &fvideo, &faudio, &hres, &vres, &numfps, &denfps, &vbitrate, &abitrate, &block, &next, &semaforo, &duration, &timestamp, &mac, &last_connect, &tv_id)
			if err != nil {
				log.Println(err)
			}
		}	
		s.mu_seg.Lock()
		s.lastdownload = filename+".ts"
		lineacomandos = fmt.Sprintf("/usr/bin/wget --limit-rate=625k -S -O %sdownload.ts --post-data tv_id=%s&mac=%s&semaforo=%s&downloaded=%s&bytes=%d&md5sum=%s http://%s/download.cgi",
											rootdir,s.settings["tv_id"],s.settings["mac"],semaforo,s.lastdownload,bytes,md5sum,s.settings["ip_download"])
		s.mu_seg.Unlock()
		fmt.Println(lineacomandos)
		// construimos la linea de comandos
		exe := cmdline.Cmdline(lineacomandos)
		lectura, err:= exe.StderrPipe()
		if err != nil {
			fmt.Println(err)
		}
		mReader := bufio.NewReader(lectura)
		downloaded, downloadedok = false, false
		time_semaforo := time.Now()
		go func(){
			for{
				if time.Since(time_semaforo).Seconds() > 20 { // no permitir bajadas de más de 20 segundos (2 segmentos)
					exe.Stop()
					break
				}
				time.Sleep(1 * time.Second)
			}			
		}()
		exe.Start()
		for { // bucle de reproduccion normal
			line, err := mReader.ReadString('\n')
			if err != nil {
				fmt.Println("Fin del wget !!!")
				break
			}
			line = strings.TrimRight(line, "\n")
			if strings.Contains(line, "HTTP/1.1 200 OK") {
				fmt.Println("Downloaded OK")
				downloaded = true
			}
			if strings.Contains(line, "X-Frame-Options:") {
				connected = true
				if downloaded {
					line = strings.Trim(line, " ")
					fmt.Sscanf(line, "X-Frame-Options: bytes=%d filename=%s md5sum=%s fvideo=%s faudio=%s hres=%d vres=%d numfps=%d denfps=%d vbitrate=%d abitrate=%d block=%s next=%s duration=%d timestamp=%d", 
									&g_bytes, &g_filename, &g_md5sum, &g_fvideo, &g_faudio, &g_hres, &g_vres, &g_numfps, &g_denfps, &g_vbitrate, &g_abitrate, &g_block, &g_next, &g_duration, &g_timestamp)
				}else{ // X-Frame-Options: already downloaded ; X-Frame-Options: access not granted
					fmt.Println("NOT Downloaded")
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
		segmento := fmt.Sprintf("%d",contador)
		if downloadedok{
			err = exec.Command("/bin/sh","-c","mv -f "+rootdir+"download.ts"+" "+rootdir+"segment"+segmento+".ts").Run()
			if err != nil {
				log.Println(err)
			}
			last_connect := time.Now().Unix() // es el momento de la grabación del downloaded segment
			db_mu.Lock()
			_, err = db.Exec("INSERT INTO segmentos (filename,bytes,md5sum,fvideo,faudio,hres,vres,num_fps,den_fps,vbitrate,abitrate,block,next,duration,timestamp,mac,last_connect,semaforo,tv_id) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", "segment"+segmento, g_bytes, g_md5sum,
				g_fvideo, g_faudio, g_hres, g_vres, g_numfps, g_denfps, g_vbitrate, g_abitrate, g_block, g_next, g_duration, g_timestamp, "-", last_connect, semaforo, 0)
			db_mu.Unlock()
			if err != nil {
				log.Println(err)
			}
			fmt.Println("Grabado en base de datos y fichero movido")
		}else{
			os.Remove(rootdir+"download.ts")
			fmt.Println("download.ts borrado")
		}
		time.Sleep(1 * time.Second) // re-try downloads every second
	}
}

func (s *SegPlay) Volume(up bool) { // director = secuenciador + downloader + director_pub
	s.mu_seg.Lock()
	defer s.mu_seg.Unlock()
	if up {
		if s.volume < 12 {
			s.mediawriter.WriteByte('+')
			s.mediawriter.Flush()
		}
	} else {
		if s.volume > -12 {
			s.mediawriter.WriteByte('-')
			s.mediawriter.Flush()
		}
	}
}

// equivalent to md5sum -b filename
func md5sumfunc(filename string) string {
	out, _ := exec.Command("/bin/sh", "-c", "/usr/bin/md5sum -b "+filename+" | awk '{print $1}'").CombinedOutput()

	return strings.TrimSpace(string(out))
}

// killall("omxplayer omxplayer.bin ffmpeg")
func killall(list string) {
	prog := strings.Fields(list)
	for _, v := range prog {
		exec.Command("/bin/sh", "-c", "/bin/kill -9 `ps -A | /usr/bin/awk '/"+v+"/{print $1}'`").Run()
	}
}

// convierte un string numérico en un entero int
func toInt(cant string) (res int) {
	res, _ = strconv.Atoi(cant)
	return
}
