package main

import (
	"fmt"
)

func main() {
	//strings := loadString("trans.po")
	//fmt.Println("Hello")
	res := extractStrings("Dwarf_Fortress")
	fmt.Print(res)
}
