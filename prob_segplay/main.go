package main

import (
	"github.com/isaacml/cmdline"
	"bufio"
	"strings"
	"fmt"
	"sync"
)

type SegPlay struct {
	cmdomx								string
	exe									*cmdline.Exec
	exe2								*cmdline.Exec
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


func main(){
		s := &SegPlay{
			exe: cmdline.Cmdline("ps ax"),
			exe2: cmdline.Cmdline("ps ax"),
		}
		s.exe2 = cmdline.Cmdline("/usr/bin/ffmpeg -y -f mpegts -re -i /tmp/fifo1 -f mpegts -acodec copy -vcodec copy /tmp/fifo2")
		fmt.Println("[cmd2] - 1")
		lectura,err := s.exe2.StderrPipe()
		fmt.Println("[cmd2] - 11")
		if err != nil{
			fmt.Println(err)
		}
		fmt.Println("[cmd2] - 12")
		mReader := bufio.NewReader(lectura)
		fmt.Println("[cmd2] - 2")
	
		s.exe2.Start()
		fmt.Println("[cmd2] - 3")

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
				fmt.Println("Ready.....")
			}	
			if strings.Contains(line,"frame=") {
				fmt.Printf("[cmd2] %s\n",line)
			}
		}
		s.exe2.Stop()
		s.mu_seg.Lock()
		s.restamping = false
//		if !s.running { break }
		s.mu_seg.Unlock()
		fmt.Println("[cmd2] - 4")
}

/*
import (
	"github.com/isaacml/test/segplay"
	"fmt"
	"time"
)

func main(){
	settings := map[string]string{
		"overscan"		:		"1",
		"x0"			:		"0",
		"y0"			:		"0",
		"x1"			:		"719",
		"y1"			:		"575",
		"vol"			:		"1",	
	}
	
	fmt.Println("...")
	seg := segplay.SegmentPlayer("","/root/",settings)
	seg.Run()
	time.Sleep(30 * time.Second)
	seg.Stop()
	for {}
}
*/