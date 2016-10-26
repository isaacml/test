package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"io/ioutil"
	"strings"
)

func main(){
		type Encoder struct {
			Ip			[]string  `xml:"address"`
			Time 		[]string  `xml:"time"`
			Flash 		[]string  `xml:"flashver"`
		}
		type Result struct {
			Nombre 		[]string `xml:"server>application>live>stream>name"`
			Bw_in  		[]string `xml:"server>application>live>stream>bw_in"`
			Width  		[]string `xml:"server>application>live>stream>meta>video>width"`
			Height 		[]string `xml:"server>application>live>stream>meta>video>height"`
			Frame  		[]string `xml:"server>application>live>stream>meta>video>frame_rate"`
			Vcodec 		[]string `xml:"server>application>live>stream>meta>video>codec"`
			Acodec 		[]string `xml:"server>application>live>stream>meta>audio>codec"`
			EncoderList []Encoder `xml:"server>application>live>stream>client"`
		}
		resp, err := http.Get("http://panel.cdnstreamserver.com:8080/stats")
		if err != nil {
			fmt.Println(err)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		// ZONA XML DE SERVIDOR
		var r Result
		err = xml.Unmarshal(body, &r)
		if err != nil {
			fmt.Printf("xml read error: %s", err) 
			return
		}
		for k, _ := range r.Nombre {
			fmt.Printf("%s-%s-%s-%s-%s-%s-%s\n", r.Nombre[k], r.Bw_in[k], r.Width[k], r.Height[k], r.Frame[k], r.Vcodec[k], r.Acodec[k])
		}

		for key, val := range r.EncoderList {
			if strings.Contains(val.Flash[key], "FMLE"){
				fmt.Printf("%s\n", val.Flash)
			}
		}
		/*
		c := Encoder{}
		err2 := xml.Unmarshal([]byte(body), &c)
		if err2 != nil {
			fmt.Printf("xml read error: %s", err2) 
			return
		}
		for k, _ := range c.Ip {
			if strings.Contains(c.Flash[k], "FMLE"){
				fmt.Printf("%s-%s\n", c.Ip[k], c.Time[k])
			}
		}
		*/
}