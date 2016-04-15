package main

import ("fmt";"github.com/isaacml/cmdline";"bufio";"sync";"strings")

var cmd2run bool
var wg sync.WaitGroup

func main(){

	canal := make(chan int)
	
	wg.Add(2)
	
	go cmd2("/usr/bin/avconv -video_size 720x576 -framerate 25.0000 -pixel_format uyvy422 -f rawvideo -i /tmp/video_fifo -sample_rate 48k -channels 2 -f s16le -i /tmp/audio_fifo -pix_fmt yuv420p -vf yadif=3 -c:v libx264 -b:v 2000k -minrate:v 2000k -maxrate:v 2000k -bufsize:v 1835k -flags:v +cgop -profile:v high -x264-params level=4.1:keyint=100 -r:v 50.0000 -threads 8 -af volume=volume=2dB:precision=fixed -c:a libfdk_aac -profile:a aac_he -b:a 128k -s 720x288 -aspect 16:9 -hls_time 10 -hls_list_size 3 -hls_wrap 5 /usr/local/bin/capturas/Isaac_ML2-Santa_Cruz2.m3u8", canal)
	
	go cmd1("/usr/bin/capture -d 0 -m 2 -V 3 -A 2 -v /tmp/video_fifo -a /tmp/audio_fifo", canal)
	
	wg.Wait()
}

func cmd1(comando1 string, ch chan int){
	
	fmt.Println(comando1)
	defer wg.Done()
	
	for {
		fmt.Println("cmd1-1")
		decoder := cmdline.Cmdline(comando1)
		lectura,errL := decoder.StderrPipe()
		if errL != nil{
			fmt.Println(errL)
		}
		mReader := bufio.NewReader(lectura)
		fmt.Println("cmd1-2")
		<- ch
		fmt.Println("cmd1-3")
		decoder.Start()
		for{ // bucle de reproduccion normal
			line,err := mReader.ReadString('\n')
			if err != nil {
				fmt.Println("Fin del cmd1 !!!")
				break;
			}
			fmt.Printf("[cmd1] %s",line)
		}
		decoder.Stop()
		fmt.Println("cmd1-4")
		<- ch
		fmt.Println("cmd1-5")
	}
}

func cmd2(comando2 string, ch chan int){
	fmt.Println(comando2)
	defer wg.Done()
	
	for {
		fmt.Println("cmd2-1")
		cmd2run = false
		decoder := cmdline.Cmdline(comando2)
		lectura,errL := decoder.StderrPipe()
		if errL != nil{
			fmt.Println(errL)
		}
		mReader := bufio.NewReader(lectura)
		decoder.Start()
		for{ // bucle de reproduccion normal
			fmt.Println("cmd2-22")
			line,err := mReader.ReadString('\n')
			fmt.Println("cmd2-23")
			if err != nil {
				fmt.Println("Fin del cmd2 !!!")
				break;
			}
				
			if strings.Contains(line, "built on"){
				if !cmd2run {
					//time.Sleep(3*time.Second)
					fmt.Println("cmd2-2")
					ch <- 1
					fmt.Println("cmd2-3")
					cmd2run = true
				}
			}
			fmt.Printf("[cmd2] %s",line)
		}
		decoder.Stop()
		fmt.Println("cmd2-4")
		ch <- 1
		fmt.Println("cmd2-5")
	}
}