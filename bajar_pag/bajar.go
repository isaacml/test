package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	res, err := http.Get("http://www.todostreaming.es")
	
	if err != nil {
		log.Fatal(err)
	}
	
	robots, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	
	if err != nil {
		log.Fatal(err)
	}
	
	err2 := ioutil.WriteFile("pagina.html", robots, 0666)
	
	if err2 != nil {
		log.Fatal(err2)
	}
	
	fmt.Printf("%s", robots)
}
