package main

import (
	"net/http"
	"net/url"
	"io/ioutil"
	"log"
	"fmt"
)

func main(){
	v := url.Values{}
	v.Set("query", "SELECT ipv4 FROM server where id = 2")
	resp, err := http.PostForm("http://localhost:9999/data.cgi", v)
	if err!=nil{
		log.Fatal(err)
	}
	defer resp.Body.Close()
  	body, err := ioutil.ReadAll(resp.Body)
  	if err!=nil{
		log.Fatal(err)
	}
	fmt.Printf("Hola: %s\n", string(body))
}

