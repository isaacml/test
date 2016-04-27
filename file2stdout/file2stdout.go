package main

import ("fmt";"os";"io";"bufio")

var settings map[string]string = make(map[string]string)
var MB1 = 1*1024*1024

func main(){
	dumpfile("unatv.flv")
}

func dumpfile (filename string){
	fr,err := os.Open(filename)
	defer fr.Close()
	if err == nil{
//		reader := bufio.NewReader(fr)
		data := make([]byte, MB1)
		for{
			data = data[:cap(data)]
			n, err := fr.Read(data)
			fmt.Fprintf(os.Stderr, "Bytes read: %d\n", n)
	        if err != nil {
	            if err == io.EOF { 
	            	fmt.Fprintln(os.Stderr, "END OF FILE\n")
	            	break 
	           	}
	            fmt.Fprintln(os.Stderr, err)
	            return
        	}
       		data = data[:n]
       		writer := bufio.NewWriter(os.Stdout)
			writer.Write(data)
			writer.Flush()
		}
	}
	fmt.Fprintln(os.Stderr, "END OF DUMP\n")
}

func dumpdata (data []byte){
	writer := bufio.NewWriter(os.Stdout)
	writer.Write(data)
	writer.Flush()
	fmt.Fprintln(os.Stderr, "END OF DUMP")
}
