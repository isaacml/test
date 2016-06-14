package main

import (
	"strings"
	"fmt"
)

func main(){
	namefile  := "/var/segments/live/isaac-stream.m3u8"
	agent 	  := "Mozilla/5.0 (X11; Linux x86_64; rv:38.0) Gecko/20100101 Firefox/38.0 Iceweasel/38.6.1"
	forwarded := "79.109.178.183"
	remoteip  := getip("46.105.196.10:13089")
	createStats(namefile, agent, forwarded, remoteip)
}

func createStats(namefile, agent, forwarded, remoteip string){
	userAgent := map[string]string {"win":"Windows", "mac":"Mac OS X", "and":"Android", "lin":"Linux"}
	var existe bool
	var stream, ipcliente, ipproxy string
	//operaciones sobre el namefile
	fmt.Sscanf(namefile, "/var/segments/live/%s", &stream)
	nom := strings.Split(stream, ".")
	streamname := nom[0]
	username   := strings.Split(streamname, "-")
	//operaciones para el user agent
	for key, value := range userAgent{
		if strings.Contains(agent, value){
			fmt.Printf("SO: %s\n", key)
			existe = true
		}
	}
	//Agent User not find
	if !existe{			
		fmt.Printf("SO: other\n")
	}
	//Cuando el forwarded está vacio
	if forwarded == "" {
		ipcliente = remoteip
		ipproxy = ""
	}else{
		ipcliente = forwarded
		ipproxy = remoteip
	}
	fmt.Printf("Stream: %s\n", streamname)		//Nombre del stream
	fmt.Printf("User: %s\n", username[0])		//Nombre del usuario
	fmt.Printf("ClienteIP: %s\n", ipcliente)	//IP Cliente
	fmt.Printf("ProxyIP: %s\n", ipproxy)		//IP Proxy
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
