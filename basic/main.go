package main

import (
	"fmt"
	"os/exec"
	"bufio"
	"strings"
)

func main(){
/*	settings := map[string]string{
		"overscan"		:		"1",
		"x0"			:		"0",
		"y0"			:		"0",
		"x1"			:		"719",
		"y1"			:		"575",
		"vol"			:		"1",	
	}
*/	fmt.Println("...")
	decoder_exe := exec.Command("/bin/sh","-c","/usr/bin/omxplayer -s -o both --layer 1 --no-osd -b play.ts")
	stderrRead,_ := decoder_exe.StderrPipe()
	mediareader := bufio.NewReader(stderrRead)
	decoder_exe.Start()
	for{
		line,err := mediareader.ReadString('\n')
		if err != nil {
			fmt.Println("Salgo de omxplayer")	
			break
		}
		line=strings.TrimRight(line,"\n")
		fmt.Println("[cmd]",line)
	}	
	decoder_exe.Wait()
}

