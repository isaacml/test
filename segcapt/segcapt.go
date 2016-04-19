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
	"runtime"
	"io/ioutil"
	"crypto/md5"
	"io"
)

type SegCapt struct {
	cmd1, cmd2	string
	exe1, exe2	*cmdline.Exec
	settings	map[string]string
	fileupload	string
	uploaddir	string
	recording	bool
	uploading	bool
	cutsegment	bool
	lastrecord, lastupload, nextrecord int
	lastrecord_dur int
	lastrecord_timestamp int64
	lastrecord_pub, nextrecord_pub bool
	mu_seg		sync.Mutex
}

func SegmentCapturer(fileupload, uploaddir string, settings map[string]string) *SegCapt {
	seg := &SegCapt{
		exe1: cmdline.Cmdline(""),
		exe2: cmdline.Cmdline(""),
	}
	seg.mu_seg.Lock()
	defer seg.mu_seg.Unlock()
	seg.settings = settings
	seg.fileupload = fileupload
	seg.uploaddir = uploaddir
	
	// creamos el cmd1
	modo := toInt(seg.settings["v_mode"])
	seg.cmd1 = fmt.Sprintf("/usr/bin/capture -d 0 -m %d -V %s -A %d -v /tmp/video_fifo -a /tmp/audio_fifo", modo, seg.settings["v_input"], seg.settings["a_input"])

	// creamos el cmd2
	var yadif, rv, outs string
	var resol = []string{} // resol=append(resol,"1920x1080")
	var rate = []float64{}
	var interlaced = []bool{} // interlaced=append(interlaced,true)
	var keyint int
	output_video:= toInt(seg.settings["v_output"])
	hvres := strings.Split(resol[modo],"x")
	hres  := toInt(hvres[0])
	vres  := toInt(hvres[1])
	if interlaced[modo] {
		yadif = " -vf 'yadif=3'"
		rv = fmt.Sprintf(" -r:v %.4f", 2.0*rate[modo])
		keyint = int(4.0 * rate[modo])
		outs = fmt.Sprintf("%dx%d", hres, vres/2)
	}else{
		keyint = int(2.0 * rate[modo])
		outs = fmt.Sprintf("%dx%d", hres, vres)
	}
	var v_bitrate int
	switch output_video {
		case 0: 
			v_bitrate = 1000
		case 1:
			v_bitrate = 2000
		case 2,3:
			v_bitrate = 3000
	}
	seg.cmd2 = fmt.Sprintf("/usr/bin/avconv -video_size %s -framerate %.4f -pixel_format uyvy422 -f rawvideo -i /tmp/video_fifo -sample_rate 48k -channels 2 -f s16le -i /tmp/audio_fifo -pix_fmt yuv420p%s -c:v libx264 -b:v %dk -minrate:v %dk -maxrate:v %dk -bufsize:v 1835k -flags:v +cgop -profile:v high -x264-params level=4.1:keyint=%d%s -threads %d -af 'volume=volume=%sdB:precision=fixed' -c:a libfdk_aac -profile:a aac_he -b:a 128k -s %s -aspect %s -hls_time 10 -hls_list_size 3 %s/%s.m3u8",
							resol[modo],rate[modo],yadif,v_bitrate,v_bitrate,v_bitrate,keyint,rv,runtime.NumCPU(),seg.settings["a_level"],outs,seg.settings["aspect_ratio"],seg.uploaddir,seg.fileupload)

	seg.recording = false
	seg.uploading = false
	seg.lastrecord = -1 // si < 0 significa que no hay segmento aun
	seg.lastupload = -1 // si < 0 significa que no hay segmento aun
	seg.nextrecord = -1
	seg.lastrecord_pub = false
	seg.nextrecord_pub = false
	seg.cutsegment = false
	
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

func (s *SegCapt) Run(pub bool) error {
	var err error
	ch := make(chan int)
	
	s.mu_seg.Lock()
	s.lastrecord_pub = pub
	s.nextrecord_pub = pub
	s.mu_seg.Unlock()

	go s.command1(ch)
	go s.command2(ch)
	go s.contactServer()
	
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
		startofseg := true
		for{ // bucle de reproduccion normal
			tiempo = time.Now()
			if startofseg {
				s.lastrecord_timestamp = tiempo.Unix()
				startofseg = false
			}
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
				startofseg = true
				s.mu_seg.Lock()
				s.lastrecord = s.extractsegmentid(line)
				if s.cutsegment { 
					s.nextrecord = 0
					s.cutsegment = false
				} else {
					s.nextrecord = s.lastrecord + 1
				}
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

func (s *SegCapt) CutSegment(pub bool) error {
	s.mu_seg.Lock()
	s.cutsegment = true // ha ocurrido un corte de segmento (no dice el tipo del corte)
	s.lastrecord_pub = !pub
	s.nextrecord_pub = pub
	s.mu_seg.Unlock()
	return s.exe2.Stop()
}

func (s *SegCapt) contactServer() error {
	var err error
	
	return err
}

// equivalent to md5sum -b filename
func (s *SegCapt) md5sum(filename string) string {
	buf,_ := ioutil.ReadFile(filename)
	h := md5.New()
	io.WriteString(h, string(buf))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// /usr/bin/curl -F segment=@mac_0.ts -F tv_id=2 -F filename=mac_0 -F bytes=16874 -F md5sum=eed91981eafe1106fe90c48148b250fb -F fvideo=h264 -F faudio=heaacv1 -F hres=1920 -F vres=1080 -F numfps=25 
// -F denfps=0 -F vbitrate=2300 -F abitrate=128 -F block=prog -F next=pub -F duration=3500 -F timestamp=1430765872 -F mac=d4ae52d3ea66 -F semaforo=G "http://localhost/segments/upload.php"
func (s *SegCapt) upload() {

	for{
		s.mu_seg.Lock()
		lastupload := s.lastrecord
		s.mu_seg.Unlock()
		filetoupload := s.uploaddir + s.fileupload + fmt.Sprintf("%d", lastupload) + ".ts"
		fileinfo,err := os.Stat(filetoupload)
		if err != nil {
			fmt.Println(err)
		}


	var resol = []string{} // resol=append(resol,"1920x1080")
	var rate = []float64{}
	var pal = []bool{}
	output_video:= toInt(s.settings["v_output"])
	modo := toInt(s.settings["v_mode"])
	hvres := strings.Split(resol[modo],"x")
	hres  := hvres[0]
	vres  := hvres[1]
	var v_bitrate int
	switch output_video {
		case 0: 
			v_bitrate = 1000
		case 1:
			v_bitrate = 2000
		case 2,3:
			v_bitrate = 3000
	}
	numfps := int(rate[modo])
	denfps := 0
	if !pal[modo] { denfps = 1 }
	block := "prog"
	if s.lastrecord_pub { block = "pub" }
	next := "prog"
	if s.nextrecord_pub { next = "pub" }
		/////////////////////////////////////////////////////////////////////////////////////////////
		// fileupload, uploaddir
		lineacomandos := fmt.Sprintf("/usr/bin/curl -F segment=@%s -F tv_id=2 -F filename=%s -F bytes=%d -F md5sum=%s -F fvideo=h264 -F faudio=heaacv1 -F hres=%s -F vres=%s -F numfps=%d -F denfps=%d -F vbitrate=%d -F abitrate=128 -F block=%s -F next=%s -F duration=3500 -F timestamp=1430765872 -F mac=d4ae52d3ea66 -F semaforo=G http://localhost/segments/upload.php",
										filetoupload, s.fileupload + fmt.Sprintf("%d", lastupload),fileinfo.Size(), s.md5sum(filetoupload),hres,vres,numfps,denfps,v_bitrate,block,next)






		/////////////////////////////////////////////////////////////////////////////////////////////
		exe := cmdline.Cmdline(lineacomandos)
		/////////////////////////////////////////////////////////////////////////////////////////////
		// fileupload, uploaddir







		/////////////////////////////////////////////////////////////////////////////////////////////
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
		// borramos siempre el lastupload xq ya lo hemos subido
		s.deletefile(s.lastupload)
		// borramos el resto de ficheros, que no sean: lastrecord, nextrecord
		ficheros := make(map[string]int)
 		file, err := os.Open(s.uploaddir)
		if err != nil {
			fmt.Println(err)
		} else {
			elements, err := file.Readdirnames(0)
			if err != nil {
				fmt.Println(err)
			} else {
				for k,v:=range elements {
					ficheros[v] = k
				}
				delete(ficheros,fmt.Sprintf("%s%d.ts",s.fileupload,s.lastrecord))
				delete(ficheros,fmt.Sprintf("%s%d.ts",s.fileupload,s.nextrecord))
				for k,_:=range ficheros {
					os.Remove(fmt.Sprintf("%s%s",s.uploaddir,k))
				}
			}
			file.Close()
		}			
		s.mu_seg.Unlock()
		
	}
}

func (s *SegCapt) deletefile(index int) error {
	return os.Remove(fmt.Sprintf("%s%s%d.ts",s.uploaddir,s.fileupload,index))
}

// convierte un string numérico en un entero int
func toInt(cant string) (res int) {
	res, _ = strconv.Atoi(cant)
	return
}


