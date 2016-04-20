package main

import ("os";"fmt")

var ficheros map[int]string = make(map[int]string)

func main(){
	
	records(3, 4)
}

func records(lastrecord int, nextrecord int){
	
	fileupload:="testing"
	ficheros := make(map[string]int)
    file, err := os.Open("/home/isaac/SEGMENTS/")
    if err != nil {
      fmt.Println(err)
    }
    elements, err := file.Readdirnames(0)
    if err != nil {
      fmt.Println(err)
    }
    for k,v:=range elements {
      ficheros[v] = k
    }
    
    delete(ficheros,fmt.Sprintf("%s%d.ts",fileupload,lastrecord))
    delete(ficheros,fmt.Sprintf("%s%d.ts",fileupload,nextrecord))

    for k,_:=range ficheros {
      os.Remove(fmt.Sprintf("%s%s","/home/isaac/SEGMENTS/",k))
    }
    file.Close()
}

