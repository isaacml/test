package main

import ("fmt";"strings")

func main() {
	nom_tv := "Isaac ML"
	nom_distro := "Santa Cruz"
	
	tv := strings.NewReplacer(" ", "_")
	tv.Replace(nom_tv)
	
	salida := fmt.Sprintf("%s-%s", tv.Replace(nom_tv), tv.Replace(nom_distro))
	
	fmt.Println(salida)
	
}

