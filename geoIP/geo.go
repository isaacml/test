package main

import (
    "fmt"
    "github.com/todostreaming/geoip2-golang"
    "log"
    "net"
)

func main() {
  var region string
    db, err := geoip2.Open("/home/isaac/DBs/GeoIP2-City.mmdb")
    if err != nil {
            log.Fatal(err)
    }
    defer db.Close()
    // If you are using strings that may be invalid, check that ip is not nil
    ip := net.ParseIP("2001:1388:10d:8704:c0e3:1eba:b814:f938")
    record, err := db.City(ip)
    if err != nil {
            log.Fatal(err)
    }
    fmt.Printf("City name: %v\n", record.City.Names["en"])
    if len(record.Subdivisions) > 0 {
      region = record.Subdivisions[0].Names["en"]
    } 
    fmt.Printf("Region name: %v\n", region)
    fmt.Printf("Country name: %v\n", record.Country.Names["en"])
    fmt.Printf("Country code: %v\n", record.Country.IsoCode)
    fmt.Printf("Time zone: %v\n", record.Location.TimeZone)
    fmt.Printf("Coordinates: %v, %v\n", record.Location.Latitude, record.Location.Longitude)
    // Output:
    // Portuguese (BR) city name: Londres
    // English subdivision name: England
    // Russian country name: Великобритания
    // ISO country code: GB
    // Time zone: Europe/London
    // Coordinates: 51.5142, -0.0931
}