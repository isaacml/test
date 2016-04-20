package main

import ("fmt")

var fileupload string

func main(){
	
	fileupload = "testing"
	
	duration("EXT-X-SEGMENTFILE:testing0.ts",fileupload)
	duration("EXT-X-SEGMENTFILE:testing22.ts",fileupload)
	duration("EXT-X-SEGMENTFILE:testing344.ts",fileupload)
}

func duration(linea string, file string) { // EXT-X-SEGMENTFILE:testing654757575.ts (fileupload = testing)
  var ret int
  fmt.Sscanf(linea,fmt.Sprintf("EXT-X-SEGMENTFILE:%s%%d.ts",file),&ret)
  fmt.Println(ret)
/*
  archivo  := strings.Split(linea, ":") // Separo por los dos puntos
  unext    := strings.Split(archivo[1], ".") // Quito la extension
  segmento := strings.Trim(unext[0], s.fileupload) //Quitamos el nombre del fichero
  ret,_ = strconv.Atoi(segmento)
*/  
}
