package main

import (
	"strings"
	"fmt"
)

func main(){
	namefile  := "/var/segments/live/stream.m3u8"
	agent 	  := "Mozilla/5.0 (X11; Linux x86_64; rv:38.0) Gecko/20100101 Firefox/38.0 Iceweasel/38.6.1"
	forwarded := "79.109.178.183"
	remoteip  := getip("46.105.196.10:13089")
	
	createStats(namefile, agent, forwarded, remoteip)
}

func createStats(namefile, agent, forwarded, remoteip string){
	userAgent := map[string]string {"win":"Windows", "mac":"Mac OS X", "and":"Android", "lin":"Linux"}
	var existe bool
	username  := "isaac"
	var stream string
	//operaciones sobre el namefile
	fmt.Sscanf(namefile, "/var/segments/live/%s", &stream)
	streamname := strings.Split(stream, ".")
	namefile = username+"-"+streamname[0]
	//operaciones para el user agent
	for key, value := range userAgent{
		if strings.Contains(agent, value){
			fmt.Printf("%s\n", key)
			existe = true
		}else{
			existe = false
		}
	}
	//Cuando el forwarded está vacio
	if forwarded == "" {
		forwarded = remoteip
	}
	fmt.Printf("%s\n", namefile)	//Nombre del fichero
	fmt.Printf("%s\n", forwarded)	//IP forwarded
	fmt.Printf("%s\n", remoteip)	//IP remota
	if existe == false {			//Agent User no find
		fmt.Printf("other\n")
	}
}

func getip(pseudoip string) string {
	var res string
	if strings.Contains(pseudoip, "]:") {
	  part := strings.Split(pseudoip, "]:")
	  res = part[0]
	  res = res[1:]
	}else{
	  part := strings.Split(pseudoip, ":")
	  res = part[0]
	} 
	  return res
}
