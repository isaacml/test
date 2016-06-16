package main

import (
    "fmt"
    "github.com/todostreaming/syncmap"
)

func main() {
	array := []string {"Juan", "Pedro"}
    m := syncmap.New()
    m.Set("one", 1)
    v, ok := m.Get("one")
    fmt.Println(v, ok)  // 1, true

    v, ok = m.Get("not_exist")
    fmt.Println(v, ok)  // nil, false

    m.Set("two", 2)
    m.Set("three", "three")
    m.Set("cuatro", "4")
    m.Set("cinco", 5)
    m.Set("seis",6)
    for _,v := range array{
   		 fmt.Println(v)
    }

    for item := range m.IterItems() {
        fmt.Println("key:", item.Key, "value:", item.Value)
    }
}

