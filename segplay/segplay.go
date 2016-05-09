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
)

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
	go s.director() // envia segmentos a /tmp/fifo1 cuando s.playing && s.restamping

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

func (s *SegPlay) director() { // director = secuenciador + downloader + director_pub
	for {
		if s.playing && s.restamping {
			fmt.Println("Preparado para recibir segmentos por el /tmp/fifo1")
			break
		}
		runtime.Gosched()
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
