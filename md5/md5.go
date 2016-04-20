package main

import ("fmt";"io";"io/ioutil";"crypto/md5")

func main(){
	fmt.Println(md5sum("/home/isaac/sample.bin"))
}

func md5sum(filename string) string {
  buf,_ := ioutil.ReadFile(filename)  ///-P
  h := md5.New()
  io.WriteString(h, string(buf))
  out := fmt.Sprintf("%x", h.Sum(nil)) // out es un string con el md5sum -b /etc/init.d/streamixserver
  return out
}