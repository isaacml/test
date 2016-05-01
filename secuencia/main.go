package main

import (
	"os"
	"fmt"
	"log"
	"io"
)

// secuencia /tmp/fifo
// secuencia /tmp/outputmix.ts
func main(){
	segments := []string {
		"play1.ts", "play2.ts", "play3.ts", "play4.ts",
	}
	fmt.Println("Comenzamos ...")
	
	if len(os.Args) != 2 {
		os.Exit(1)
	}
	fw,err := os.OpenFile(os.Args[1], os.O_WRONLY | os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	defer fw.Close()
	for _,v := range segments {
		fr,err := os.Open(v)
		if err != nil {
			log.Fatalln(err)
		}
		if n,err := io.Copy(fw,fr); err == nil {
			fmt.Printf("Copiados %d bytes\n",n)
		}else{
			log.Println(err) // no salimos en caso de error de copia
		}
		fr.Close()
	}
	fmt.Println("Finalizamos")
}
