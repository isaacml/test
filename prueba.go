package main

import ("fmt";"github.com/isaacml/cmdline";"bufio";"strings")

var status_string, status string
var mReader *bufio.Reader

func main(){
	fmt.Println("Hola Hola")
	player2()
}

func player2(){
	
	decoder := cmdline.Cmdline("/usr/bin/avconv -video_size 720x576 -framerate 25.0000 -pixel_format uyvy422 -f rawvideo -i /tmp/video_fifo -sample_rate 48k -channels 2 -f s16le -i /tmp/audio_fifo -pix_fmt yuv420p -vf 'yadif=3' -c:v libx264 -b:v 2000k -minrate:v 2000k -maxrate:v 2000k -bufsize:v 1835k -flags:v +cgop -profile:v high -x264-params level=4.1:keyint=100 -r:v 50.0000 -threads 8 -af 'volume=volume=2dB:precision=fixed' -c:a libfdk_aac -profile:a aac_he -b:a 128k -s 720x288 -aspect 16:9 -hls_time 10 -hls_list_size 3 -hls_wrap 5 /home/isaac/capture/Isaac_ML2-Santa_Cruz2.m3u8")
	
	lectura,errL := decoder.StderrPipe()
	
	if errL != nil{
		fmt.Println(errL)
	}
	
	mReader = bufio.NewReader(lectura)
	
	decoder.Start()
	
		for{ // bucle de reproduccion normal
			line,err := mReader.ReadString('\r')
			fmt.Println("%s",line)
			if err != nil {
				fmt.Println("Fin de la reproducción !!!")
				fmt.Println(err)
				break;
			}
			line=strings.TrimRight(line,"\r")
			fmt.Printf("%s",line)
			if strings.Contains(line,"A-V:") {
				fmt.Printf("%s\r",line)
			}
		}
	
	decoder.Stop()
}