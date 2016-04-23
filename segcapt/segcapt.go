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
	"os/exec"
)

var resol = []string{} // resol=append(resol,"1920x1080")
var rate = []float64{}
var interlaced = []bool{} // interlaced=append(interlaced,true)
var pal = []bool{}

type SegCapt struct {
	cmd1, cmd2	string
	exe1, exe2	*cmdline.Exec
	settings	map[string]string
	fileupload	string							// basename del fichero a subir que irá seguido de un número indice de segmento
	uploaddir	string							// directorio RAMdisk donde se guardan los ficheros capturados listos para subir
	recording	bool
	uploading	bool
	cutsegment	bool							// acaba de ocurrir un cutsegment por cambio PROGRAM <=> PUBLI (no natural)
	lastrecord, lastupload, nextrecord int		// indice entero del ultimo segmento capturado y cerrado(lastrecord), ultimo subido(lastupload) y siguiente capturandose ahora mismo(nextrecord)
	lastrecord_dur int							// duracion en segundos enteros del ultimo segmento capturado y cerrado
	lastrecord_timestamp int64					// timestamp del comienzo del ultimo segmento capturado y cerrado
	lastrecord_pub, nextrecord_pub bool			// true si es un segmento de publicidad, false si es un segmento de programa
	semaforo	string							// R(red), Y(yellow), G(green)
	mu_seg		sync.Mutex
}

func SegmentCapturer(fileupload, uploaddir string, settings map[string]string) *SegCapt {
	seg := &SegCapt{
		exe1: cmdline.Cmdline("ps ax"),
		exe2: cmdline.Cmdline("ps ax"),
//		settings: settings,
	}
	seg.mu_seg.Lock()
	defer seg.mu_seg.Unlock()
	seg.settings = settings
	seg.fileupload = fileupload
	seg.uploaddir = uploaddir
	seg.recording = false
	seg.uploading = false
	seg.lastrecord = -1 // si < 0 significa que no hay segmento aun
	seg.lastupload = -1 // si < 0 significa que no hay segmento aun
	seg.nextrecord = -1
	seg.lastrecord_pub = false
	seg.nextrecord_pub = false
	seg.cutsegment = false
	seg.lastrecord_timestamp = time.Now().Unix()
	seg.lastrecord_dur = 0
	seg.semaforo = "G" // comenzamos en verde
	seg.bmdinfo()
	
	// creamos el cmd1
	modo := toInt(seg.settings["v_mode"])
	seg.cmd1 = fmt.Sprintf("/usr/bin/capture -d 0 -m %d -V %s -A %s -v /tmp/video_fifo -a /tmp/audio_fifo", modo, seg.settings["v_input"], seg.settings["a_input"])

	// creamos el cmd2
	var yadif, rv, outs string
	var keyint int
	output_video:= toInt(seg.settings["v_output"])
	hvres := strings.Split(resol[modo],"x")
	hres  := toInt(hvres[0])
	vres  := toInt(hvres[1])
	if interlaced[modo] {
		yadif = " -vf yadif=3"
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
	seg.cmd2 = fmt.Sprintf("/usr/bin/avconv -video_size %s -framerate %.4f -pixel_format uyvy422 -f rawvideo -i /tmp/video_fifo -sample_rate 48k -channels 2 -f s16le -i /tmp/audio_fifo -pix_fmt yuv420p%s -c:v libx264 -b:v %dk -minrate:v %dk -maxrate:v %dk -bufsize:v 1835k -flags:v +cgop -profile:v high -x264-params level=4.1:keyint=%d%s -threads %d -af volume=volume=%sdB:precision=fixed -c:a libfdk_aac -profile:a aac_he -b:a 128k -s %s -aspect %s -hls_time 10 -hls_list_size 3 %s%s.m3u8",
							resol[modo],rate[modo],yadif,v_bitrate,v_bitrate,v_bitrate,keyint,rv,runtime.NumCPU(),seg.settings["a_level"],outs,seg.settings["aspect_ratio"],seg.uploaddir,seg.fileupload)
	
	return seg
}

func (s *SegCapt) Print() {
	fmt.Printf("[resol]=%v\n", resol)
	fmt.Printf("[rate]=%v\n", rate)
	fmt.Printf("[interlaced]=%v\n", interlaced)
	fmt.Printf("[pal]=%v\n", pal)
	fmt.Printf("[cmd1] %s\n\n[cmd2] %s\n",s.cmd1,s.cmd2)
}

func (s *SegCapt) bmdinfo() {
	var name, modes, card bool
	card = false
	s.settings["cardname"] = ""
	
	cmd := exec.Command("/usr/bin/bmdinfo")
	stdoutRead, _ := cmd.StdoutPipe()
	reader := bufio.NewReader(stdoutRead)
	cmd.Start()
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimRight(line, "\n")

		if line == "" {
			continue
		}
		if strings.Contains(line, "#ERROR") {
			if !card {
				s.settings["cardname"] = "NO_CARD"
				break
			}
		} else if strings.Contains(line, "#NAME") {
			name = true
			modes = false
			card = true
		} else if strings.Contains(line, "#MODES input") {
			name = false
			modes = true
		} else if strings.Contains(line, "#MODES output") {
			name = false
			modes = false
		} else if strings.Contains(line, "#INPUT") {
			name = false
			modes = false
		} else if strings.Contains(line, "#OUTPUT") {
			name = false
			modes = false
		} else { // linea a interpretar
			if name {
				s.settings["cardname"] = line
			} else if modes { // <option value="0" {{ if eq .mode "0" }}selected{{else}}{{end}}>NTSC</option>
				item := strings.Split(line, ":")
				resol = append(resol, item[2])
				r, _ := strconv.ParseFloat(item[3], 64)
				rate = append(rate, r)
				// pal=pal(pal,true)
				if strings.Contains(item[1], "PAL") {
					pal = append(pal, true)
				} else if strings.Contains(item[1], "NTSC") {
					pal = append(pal, false)
				} else if strings.Contains(item[1], "50") || strings.Contains(item[1], "25") {
					pal = append(pal, true)
				} else {
					pal = append(pal, false)
				}
				// item[1] say i or p
				if strings.Contains(item[1], "Progre") || strings.Contains(item[1], "p") { // progresivo
					interlaced = append(interlaced, false)
				} else { // entrelazado
					interlaced = append(interlaced, true)
				}
			}
		}
	}
	cmd.Wait()
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
	go s.upload()
	//go s.contactServer()
	
	return err
}

func (s *SegCapt) command1(ch chan int){ // capture
	
	fmt.Println("[cmd1]",s.cmd1)
	
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
			//fmt.Printf("[cmd1] %s\n",line)
		}

		s.exe1.Stop()
		<- ch
	}
}

func (s *SegCapt) command2(ch chan int){ // avconv
	fmt.Println("[cmd2]",s.cmd2)
	var tiempo int
	var cmd2run bool
	
	for {
		cmd2run = false
		s.exe2 = cmdline.Cmdline(s.cmd2)
		lectura,errL := s.exe2.StderrPipe()
		if errL != nil{
			fmt.Println(errL)
		}
		mReader := bufio.NewReader(lectura)
		tiempo = time.Now().Second()
		go func() {
			for {
				if (time.Now().Second()-tiempo) > 2 {
					s.exe2.Stop() // SIGKILL
					break
				}
				time.Sleep(1 * time.Second)
			}
		}()
		s.exe2.Start()
		s.mu_seg.Lock()
		s.recording = true
		s.mu_seg.Unlock()
		startofseg := true

		for{ // bucle de reproduccion normal
			tiempo = time.Now().Second()//; time.Sleep(5 * time.Second)
			if startofseg {
				s.mu_seg.Lock()
				s.lastrecord_timestamp = time.Now().Unix() // seconds from 1970-01-01 UTC
				s.mu_seg.Unlock()
				startofseg = false
			}
			line,err := mReader.ReadString('\n') // bloquea
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
				fmt.Sscanf(line,fmt.Sprintf("EXT-X-SEGMENTFILE:%s%%d.ts",s.fileupload),&s.lastrecord)
				if s.cutsegment {
					s.nextrecord = 0
					s.cutsegment = false
				} else {
					s.nextrecord = s.lastrecord + 1
				}
				s.mu_seg.Unlock()
				fmt.Printf("[cmd2] %s\n",line)
			}
			if strings.Contains(line, "EXTINF:") { // EXTINF:10   (int seconds)
				dur:=0
				fmt.Sscanf(line,"EXTINF:%d",&dur)
				s.mu_seg.Lock()
				s.lastrecord_dur = dur
				s.mu_seg.Unlock()
				fmt.Printf("[cmd2] %s\n",line)
			}
			//fmt.Printf("[cmd2] %s\n",line)
		}

		s.exe2.Stop()
		s.mu_seg.Lock()
		s.recording = false
		s.mu_seg.Unlock()
		ch <- 1
	}
}

func (s *SegCapt) CutSegment(pub bool) error { // pub=true si entramos en publicidad, pub=false si salimos de la publicidad con este corte de segmento
	s.mu_seg.Lock()
	s.cutsegment = true // ha ocurrido un corte de segmento (no dice el tipo del corte)
	s.lastrecord_pub = !pub
	s.nextrecord_pub = pub
	s.mu_seg.Unlock()
	return s.exe2.SigInt() // envio un Ctrl-C al avconv cmd2
}

func (s *SegCapt) contactServer() error {
	var err error
	
	return err
}

// equivalent to md5sum -b filename
func (s *SegCapt) md5sum(filename string) string {
	out,_ := exec.Command("/bin/sh","-c","md5sum -b " + filename + " | awk '{print $1}'").CombinedOutput()
	
	return strings.TrimSpace(string(out))
}

func (s *SegCapt) upload() {
	var lastupload int
	var uploadedok bool
	
	for{
		s.mu_seg.Lock()
		// vamos a decidir si hay un segmento nuevo para subir, si no, hacemos continue
		if (s.lastrecord >= 0) && (s.lastrecord != s.lastupload) { // podemos subir lastrecord
			lastupload = s.lastrecord
		}else{
			s.mu_seg.Unlock()
			time.Sleep(1000 * time.Millisecond) // hacemos el continue si no hay nada nuevo 100 ms)
			continue
		}
		////////////////////////////////////////////////////////////////////////////
		filetoupload := s.uploaddir + s.fileupload + fmt.Sprintf("%d", lastupload) + ".ts"
		fileinfo,err := os.Stat(filetoupload) // fileinfo.Size()
		if err != nil {
			fmt.Println(err)
		}
		filesize := fileinfo.Size()
		
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
		b := s.fileupload + fmt.Sprintf("%d", lastupload)
		c := s.md5sum(filetoupload)
		lineacomandos := fmt.Sprintf("/usr/bin/curl -F segment=@%s -F tv_id=%s -F filename=%s -F bytes=%d -F md5sum=%s -F fvideo=%s -F faudio=%s -F hres=%s -F vres=%s -F numfps=%d -F denfps=%d -F vbitrate=%d -F abitrate=%s -F block=%s -F next=%s -F duration=%d -F timestamp=%d -F mac=%s -F semaforo=%s http://%s/upload.cgi",
										filetoupload ,s.settings["tv_id"] ,b ,filesize, c, s.settings["fvideo"], s.settings["faudio"],hres,vres,numfps,denfps,v_bitrate, s.settings["abitrate"],block,next,s.lastrecord_dur,s.lastrecord_timestamp,s.settings["mac"],s.semaforo,s.settings["ip_upload"])

		s.mu_seg.Unlock()
		fmt.Printf("[curl] %s\n",lineacomandos)
		/////////////////////////////////////////////////////////////////////////////////////////////
		exe := cmdline.Cmdline(lineacomandos)
		lectura,errL := exe.StdoutPipe()
		if errL != nil{
			fmt.Println(errL)
		}
		mReader := bufio.NewReader(lectura)
		time_semaforo := time.Now()
		exe.Start()
		for{ // bucle de reproduccion normal
			line,err := mReader.ReadString('\n')
			if err != nil {
				fmt.Println("Fin del curl !!!")
				break;
			}
			line = strings.TrimRight(line,"\n")
			if line == "OK" {
				fmt.Println("Uploaded OK")
				uploadedok = true
			} else {
				fmt.Println("Uploaded BAD")
				uploadedok = false
			}
			fmt.Printf("[curl] %s\n",line)
		}
		exe.Stop()
		dur_semaforo := time.Since(time_semaforo).Seconds()

		s.mu_seg.Lock()
		// decidir el color del semaforo
		var color float64
		color = float64(dur_semaforo)/float64(s.lastrecord_dur)
		switch {
			case color > 1.2 :
				s.semaforo = "R"
			case color < 0.8 :
				s.semaforo = "G"
			default :
				s.semaforo = "Y"
		}
		if !uploadedok {
			s.semaforo = "R"
			s.mu_seg.Unlock()
			time.Sleep(1 * time.Second) // fail on upload, wait for 1 second until next attempt
			continue 
		}
		// el fichero ha subido bien, y nos metemos en el post-proceso normal
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
				delete(ficheros,fmt.Sprintf("%s.m3u8",s.fileupload)) // que no borre el .m3u8
				for k,_:=range ficheros {
					fmt.Println("[delete file]",k)
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

// ls -1 -p filter
// ejemplo: filter = /var/segments/testing*.ts
func listfiles(filter string) []string {
  var arreglo = []string {}
  out:= exec.Command("/bin/sh","-c","ls -1 -p "+ filter)
  leer, err := out.StdoutPipe()
  if err != nil{
    fmt.Println(err)
  }
  mReader := bufio.NewReader(leer)
  out.Start()
  for{
    line,err := mReader.ReadString('\n')
    if err != nil{
      break
    }
    if line == "\n" {
      break
    }
    line = strings.TrimRight(line,"\n")  
    arreglo = append(arreglo, line)
  }
  out.Wait()
  
  return arreglo
}

