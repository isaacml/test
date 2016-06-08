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
	var i int
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
	for {
		file := fmt.Sprintf("play%d.ts",i)
		fr, err := os.Open(file) // read-only
		if err != nil {
			log.Fatalln(err)
		}
		if n, err := io.Copy(fw, fr); err == nil {
			fmt.Printf("[%s]Copiados %d bytes\n", file, n)
		} else {
			log.Println(err) // no salimos en caso de error de copia
		}
		fr.Close()
		i++
		if i > 5 {
			i = 0
		}
		// wait for an Intro key
		//r := bufio.NewReader(os.Stdin)
		//r.ReadByte()
		//time.Sleep(200 * time.Millisecond)
	}
	fmt.Println("Finalizamos")
}
