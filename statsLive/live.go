package main

import (
	"strings"
	"fmt"
)
func main(){
	userAgent := map[string]string {"win":"Windows", "mac":"Mac OS X", "and":"Android", "lin":"Linux"}
	existe 	  := false
	username  := "isaac"
	namefile  := "/var/segments/live/stream.m3u8"
	agent 	  := "Mozilla/5.0 (X11; Linux x86_64; rv:38.0) Gecko/20100101 Firefox/38.0 Iceweasel/38.6.1"
	forwarded := "79.109.178.183"
	remoteip  := "46.105.196.10:43025"
	
	//operaciones para el nombre de fichero
	archivo 	:= strings.Split(namefile, "/")
	streamname  := strings.Split(archivo[4], ".")
	namefile = username+"-"+streamname[0]
	//operaciones para la ip remota
	ip 			:= strings.Split(remoteip, ":")
	//operaciones para el user agent
	for key, value := range userAgent{
		if strings.Contains(agent, value){
			fmt.Printf("%s\n", key)
			existe = true
		}else{
			existe = false
		}
	}
	fmt.Printf("%s\n", namefile)	//Nombre del fichero
	fmt.Printf("%s\n", forwarded)	//IP forwarded
	fmt.Printf("%s\n", ip[0])		//IP remota
	if existe == false{				//Agent User no find
		fmt.Printf("other\n")
	}
}
