package main

import ("fmt";"os/exec";"bufio";"strings")

func main(){
	fmt.Println(lista_segments("/home/isaac/SEGMENTS/testing*"))
}

func lista_segments(archivo string) ([]string){
	var arreglo = []string {}
	out:= exec.Command("/bin/sh","-c","ls -1 -p "+ archivo)
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
	return arreglo
}