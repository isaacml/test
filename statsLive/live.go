package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"io/ioutil"
)

func main() {

type Result struct {
	Nombre  []string `xml:"server>application>live>stream>name"`
	Time  	[]string `xml:"server>application>live>stream>time"`
	Bw_in  	[]string `xml:"server>application>live>stream>bw_in"`
	Ip  	[]string `xml:"server>application>live>stream>client>address"`
	Width  	[]string `xml:"server>application>live>stream>meta>video>width"`
	Height  []string `xml:"server>application>live>stream>meta>video>height"`
	Frame   []string `xml:"server>application>live>stream>meta>video>frame_rate"`
	Vcodec  []string `xml:"server>application>live>stream>meta>video>codec"`
	Acodec  []string `xml:"server>application>live>stream>meta>audio>codec"`
}
resp, err := http.Get("http://cdn.nulldrops.com:8080/stats")
if err != nil {
	fmt.Println(err)
}
defer resp.Body.Close()
body, err := ioutil.ReadAll(resp.Body)
v := Result{}
err = xml.Unmarshal([]byte(body), &v)
if err != nil {
	fmt.Printf("error: %v", err)
	return
}
fmt.Printf("Name: %q\n", v.Nombre)
fmt.Printf("Time: %q\n", v.Time)
fmt.Printf("Bw_in: %q\n", v.Bw_in)
fmt.Printf("IP: %q\n", v.Ip)
fmt.Printf("Width: %q\n", v.Width)
fmt.Printf("Height: %q\n", v.Height)
fmt.Printf("Frame Rate: %q\n", v.Frame)
fmt.Printf("Codec de video: %q\n", v.Vcodec)
fmt.Printf("Codec de audio: %q\n", v.Acodec)

}
