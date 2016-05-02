package segplay

import (
	"github.com/isaacml/cmdline"
	"sync"
	"fmt"
	"strconv"
	"bufio"
	"strings"
)

type SegPlay struct {
	cmdomx								string
	exe									*cmdline.Exec
	mediawriter							*bufio.Writer					// por aqui puedo enviar caracteres al omxplayer
	settings							map[string]string				// read-only map
	downloaddir							string							// directorio RAMdisk donde se guardan los ficheros bajados del server y listos para reproducir
	pubdir								string							// directorio del HD donde se guardan los ficheros de publicidad locales
	playing								bool							// omxplayer esta reproduciendo
	restamping							bool							// ffmpeg esta reestampando
	downloading							bool							// esta bajando segmentos
	running								bool							// proceso completo funcionando
	lastplay, nextplay, lastdownload	string							// nombre completo de los ficheros 		
	lastplay_pub, nextplay_pub 			bool
	lastplay_timestamp 					int64
	semaforo							string							// R(red), Y(yellow), G(green) download speed	
	mu_seg								sync.Mutex			
}

func SegmentPlayer(pubdir, downloaddir string, settings map[string]string) *SegPlay {
	seg := &SegPlay{
		exe: cmdline.Cmdline("ps ax"),
	}
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
	
	var overscan string
	if seg.settings["overscan"] == "1" {
		fmt.Sprintf(" --win %s,%s,%s,%s",seg.settings["x0"],seg.settings["y0"],seg.settings["x1"],seg.settings["y1"])
	}
	vol := toInt(seg.settings["vol"])
	// creamos el cmdomx
	// /usr/bin/omxplayer -s -o both --vol 100 --hw --win '0 0 719 575' --no-osd -b /tmp/fifo2
	seg.cmdomx = fmt.Sprintf("/usr/bin/omxplayer -s -o both --vol %d --hw%s --layer 1 --no-osd -b /tmp/fifo2", 100 * vol, overscan)
	
	return seg
}

func (s *SegPlay) Run() error {
	var err error
	
	s.mu_seg.Lock()
	if s.running { // ya esta corriendo
		s.mu_seg.Unlock()
		return fmt.Errorf("segplay: ALREADY_RUNNING_ERROR")
	}
	s.running = true // comienza a correr
	s.mu_seg.Unlock()

	go s.command1()
	go s.command2()
	go s.director() // envia segmentos a /tmp/fifo1 cuando s.playing && s.restamping 
	
	return err
}

func (s *SegPlay) command1(){ // omxplayer
	for {
		s.exe = cmdline.Cmdline(s.cmdomx)
		lectura,err := s.exe.StderrPipe()
		if err != nil{
			fmt.Println(err)
		}
		mReader := bufio.NewReader(lectura)

		stdinWrite,err := s.exe.StdinPipe()
		if err != nil{
			fmt.Println(err)
		}
		s.mediawriter = bufio.NewWriter(stdinWrite)	
		
		s.exe.Start()

		for{ // bucle de reproduccion normal
			line,err := mReader.ReadString('\n')
			if err != nil {
				fmt.Println("Fin del omxplayer !!!")
				break;
			}
			line = strings.TrimRight(line,"\n")	
			if strings.Contains(line,"Comenzando...") {
				s.mu_seg.Lock()
				s.playing = true
				s.mu_seg.Unlock()
			}
			if strings.Contains(line,"M:") {
				fmt.Printf("[cmd1] %s\n",line)
			}
		}
		s.exe.Stop()
		s.mu_seg.Lock()
		s.playing = false
		if !s.running { break }
		s.mu_seg.Unlock()
	}
}

func (s *SegPlay) command2(){ // ffmpeg
	for {
		s.exe = cmdline.Cmdline("/usr/bin/ffmpeg -y -f mpegts -re -i /tmp/fifo1 -f mpegts -acodec copy -vcodec copy /tmp/fifo2")
		lectura,err := s.exe.StderrPipe()
		if err != nil{
			fmt.Println(err)
		}
		mReader := bufio.NewReader(lectura)
	
		s.exe.Start()

		for{ // bucle de reproduccion normal
			line,err := mReader.ReadString('\n')
			if err != nil {
				fmt.Println("Fin del ffmpeg !!!")
				break;
			}
			line = strings.TrimRight(line,"\n")
			if strings.Contains(line, "libswresample") {
				s.mu_seg.Lock()
				s.restamping = true
				s.mu_seg.Unlock()
			}	
			if strings.Contains(line,"frame=") {
				fmt.Printf("[cmd2] %s\n",line)
			}
		}
		s.exe.Stop()
		s.mu_seg.Lock()
		s.restamping = false
		if !s.running { break }
		s.mu_seg.Unlock()
	}
}

func (s *SegPlay) director(){ // director = secuenciador + downloader + director_pub
	for {
		if 	s.playing && s.restamping {
			fmt.Println("Preparado para recibir segmentos por el /tmp/fifo1")
		}
	}
	
}

// convierte un string numérico en un entero int
func toInt(cant string) (res int) {
	res, _ = strconv.Atoi(cant)
	return
}
