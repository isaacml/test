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
	var stream, ipcliente, ipproxy, so, user, streamname string
	//operaciones sobre el namefile
	fmt.Sscanf(namefile, "/var/segments/live/%s", &stream)
	nom := strings.Split(stream, ".")
	username := strings.Split(nom[0], "-")
	user = username[0]
	streamname = username[1]
	//operaciones para el user agent
	for key, value := range userAgent{
		if strings.Contains(agent, value){
			so = key
			fmt.Printf("SO: %s\n", so)
			existe = true
		}
	}
	//Agent User not find
	if !existe{
		so = "other"			
		fmt.Println(so)
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
	fmt.Printf("User: %s\n", user)				//Nombre del usuario
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
