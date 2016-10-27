package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"io/ioutil"
)

func main(){
		type Client struct {
			Ip			string  `xml:"address"`
			Time 		string  `xml:"time"`
			Publish 	int  	`xml:"publishing"`
		}
		type Stream struct {
			Nombre 		string `xml:"name"`
			Bw_in  		string `xml:"bw_in"`
			Width  		string `xml:"meta>video>width"`
			Height 		string `xml:"meta>video>height"`
			Frame  		string `xml:"meta>video>frame_rate"`
			Vcodec 		string `xml:"meta>video>codec"`
			Acodec 		string `xml:"meta>audio>codec"`
			ClientList  []Client `xml:"client"`
		}
		type Result struct {
			Stream		[]Stream  `xml:"server>application>live>stream"`
			
		}
		resp, err := http.Get("http://panel.cdnstreamserver.com:8080/stats")
		if err != nil {
			fmt.Println(err)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		var r Result
		err = xml.Unmarshal(body, &r)
		if err != nil {
			fmt.Printf("xml read error: %s", err) 
			return
		}
		for _, v := range r.Stream {
			for _, v2 := range v.ClientList {
				if v2.Publish == 1 {
					fmt.Printf("%s-%s-%s-%s-%s-%s-%s-%s-%s\n", v.Nombre, v.Bw_in, v.Width, v.Height, v.Frame, v.Vcodec, v.Acodec, v2.Ip, v2.Time)
				}
			}
		}
}