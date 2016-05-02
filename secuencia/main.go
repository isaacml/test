package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

// secuencia /tmp/fifo
// secuencia /tmp/outputmix.ts
func main() {
	segments := []string{
		"play1.ts", "play2.ts", "play3.ts", "play4.ts",
	}

	if len(os.Args) != 2 {
		fmt.Printf("Usage:\n\n\t%s /tmp/outputmix.ts\n\n", os.Args[0])
		os.Exit(1)
	}
	fmt.Println("Comenzamos ...")
	fw, err := os.OpenFile(os.Args[1], os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	defer fw.Close()
	for _, v := range segments {
		fr, err := os.Open(v) // read-only
		if err != nil {
			log.Fatalln(err)
		}
		if n, err := io.Copy(fw, fr); err == nil {
			fmt.Printf("Copiados %d bytes\n", n)
		} else {
			log.Println(err) // no salimos en caso de error de copia
		}
		fr.Close()
	}
	fmt.Println("Finalizamos")
}
