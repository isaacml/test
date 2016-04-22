package main

import ("fmt";"os/exec";"strings")

func main(){
	fmt.Printf("-%s-\n",md5sum("/home/isaac/sample.bin"))
}

// equivalent to md5sum -b filename
func md5sum(filename string) string {
  out,_ := exec.Command("/bin/sh","-c","md5sum -b "+ filename +" | awk '{print $1}'").CombinedOutput()
  return strings.TrimSpace(string(out))
}