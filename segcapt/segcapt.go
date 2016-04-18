package segcapt

import (
	"github.com/isaacml/cmdline"
	"sync"
	"fmt"
	"time"
	"bufio"
	"strings"
	"strconv"
	"os"
)

type SegCapt struct {
	cmd1, cmd2	string
	exe1, exe2	*cmdline.Exec
	fileupload	string
	uploaddir	string
	recording	bool
	uploading	bool
	lastrecord, lastupload int
	mu_seg		sync.Mutex
}

func SegmentCapturer(cmd1, cmd2, fileupload, uploaddir string) *SegCapt {
	seg := &SegCapt{
		exe1: cmdline.Cmdline(""),
		exe2: cmdline.Cmdline(""),
	}
	seg.mu_seg.Lock()
	defer seg.mu_seg.Unlock()
	seg.cmd1 = cmd1
	seg.cmd2 = cmd2
	seg.fileupload = fileupload
	seg.uploaddir = uploaddir
	seg.recording = false
	seg.uploading = false
	seg.lastrecord = -1 // si < 0 significa que no hay segmento aun
	seg.lastupload = -1 // si < 0 significa que no hay segmento aun
	
	return seg
}

//Function to know the state of the record at any moment
func (s *SegCapt) IsRecording() bool {
	s.mu_seg.Lock()
	defer s.mu_seg.Unlock()
		
	return s.recording  
}

//Function to know the state of the upload at any moment
func (s *SegCapt) IsUploading() bool {
	s.mu_seg.Lock()
	defer s.mu_seg.Unlock()
	
	return s.uploading  
}

func (s *SegCapt) Run() error {
	var err error
	ch := make(chan int)
	
	go s.command1(ch)
	go s.command2(ch)
	
	
	return err
}

func (s *SegCapt) command1(ch chan int){ // capture
	
	fmt.Println(s.cmd1)
	
	for {
		s.exe1 = cmdline.Cmdline(s.cmd1)
		lectura,errL := s.exe1.StderrPipe()
		if errL != nil{
			fmt.Println(errL)
		}
		mReader := bufio.NewReader(lectura)
		<- ch
		s.exe1.Start()
		for{ // bucle de reproduccion normal
			line,err := mReader.ReadString('\n')
			if err != nil {
				fmt.Println("Fin del cmd1 !!!")
				break;
			}
			line = strings.TrimRight(line,"\n")	
			fmt.Printf("[cmd1] %s\n",line)
		}
		s.exe1.Stop()
		<- ch
	}
}

func (s *SegCapt) command2(ch chan int){ // avconv
	fmt.Println(s.cmd2)
	var tiempo time.Time
	var cmd2run bool
	
	for {
		cmd2run = false
		s.exe2 = cmdline.Cmdline(s.cmd2)
		lectura,errL := s.exe2.StderrPipe()
		if errL != nil{
			fmt.Println(errL)
		}
		mReader := bufio.NewReader(lectura)
		tiempo = time.Now()
		go func() {
			for {
				if time.Since(tiempo).Seconds() > 2.0 {
					s.exe2.Stop()
					break
				}
			}
		}()
		s.exe2.Start()
		s.mu_seg.Lock()
		s.recording = true
		s.mu_seg.Unlock()
		for{ // bucle de reproduccion normal
			tiempo = time.Now()
			line,err := mReader.ReadString('\n')
			if err != nil {
				fmt.Println("Fin del cmd2 !!!")
				break;
			}
			line = strings.TrimRight(line,"\n")	
			if strings.Contains(line, "built on"){
				if !cmd2run {
					//time.Sleep(3*time.Second)
					ch <- 1
					cmd2run = true
				}
			}
			if strings.Contains(line, "EXT-X-SEGMENTFILE:") { // EXT-X-SEGMENTFILE:testing654757575.ts (fileupload = testing)
				s.mu_seg.Lock()
				s.lastrecord = s.extractsegmentid(line)
				s.mu_seg.Unlock()
			}
			fmt.Printf("[cmd2] %s\n",line)
		}
		s.exe2.Stop()
		s.mu_seg.Lock()
		s.recording = false
		s.mu_seg.Unlock()
		ch <- 1
	}
}

func (s *SegCapt) extractsegmentid(linea string) int {
	var ret int

	archivo  := strings.Split(linea, ":") // Separo por los dos puntos
	unext    := strings.Split(archivo[1], ".") // Quito la extension
	segmento := strings.Trim(unext[0], s.fileupload) //Quitamos el nombre del fichero
	ret,_ = strconv.Atoi(segmento)
	
	return ret	
}

func (s *SegCapt) CutSegment() error {
	return s.exe2.Stop()
}

// /usr/bin/curl -F segment=@mac_0.ts -F tv_id=2 -F filename=mac_0 -F bytes=16874 -F md5sum=eed91981eafe1106fe90c48148b250fb -F fvideo=h264 -F faudio=heaacv1 -F hres=1920 -F vres=1080 -F numfps=25 
// -F denfps=0 -F vbitrate=2300 -F abitrate=128 -F block=prog -F next=pub -F duration=3500 -F timestamp=1430765872 -F mac=d4ae52d3ea66 -F semaforo=G "http://localhost/segments/upload.php"
func (s *SegCapt) upload() {

	for{
		s.mu_seg.Lock()
		lastupload := s.lastrecord
		s.mu_seg.Unlock()
		filetoupload := s.uploaddir + s.fileupload + fmt.Sprintf("%d", lastupload) + ".ts"
		lineacomandos := ""+filetoupload // curl chorizo
		exe := cmdline.Cmdline(lineacomandos)

		lectura,errL := exe.StdoutPipe()
		if errL != nil{
			fmt.Println(errL)
		}
		mReader := bufio.NewReader(lectura)
		exe.Start()
		for{ // bucle de reproduccion normal
			line,err := mReader.ReadString('\n')
			if err != nil {
				fmt.Println("Fin del curl !!!")
				break;
			}
			line = strings.TrimRight(line,"\n")	
			fmt.Printf("[curl] %s\n",line)
		}
		exe.Stop()
		s.mu_seg.Lock()
		s.lastupload = lastupload
		// borrar segmentos pasados ¿cuales son?
		if (s.lastupload == s.lastrecord) || (s.lastrecord == (s.lastupload+1)) {
			 s.deletefile(s.lastupload)
		}else{
			for i := s.lastupload; i < s.lastrecord; i++ {
				s.deletefile(i)
			}
		}
		s.mu_seg.Unlock()
		
	}
}

func (s *SegCapt) deletefile(index int) error {
	return os.Remove(fmt.Sprintf("%s%s%d.ts",s.uploaddir,s.fileupload,index))	
}



