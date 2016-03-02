package main

import (
	"debug/elf"
	"fmt"
)

func main() {
	//strings := loadString("trans.po")
	//fmt.Println("Hello")
	extractStrings("Dwarf_Fortress")

}

func extractStrings(fileName string) { //[]string {

	elfFile, err := elf.Open(fileName)
	if err != nil {
		fmt.Println("Ошибка при открытии", fileName)
	}
	defer elfFile.Close()

	rodata := elfFile.Section(".rodata")
	if rodata != nil {
		data, err := rodata.Data()
		fmt.Println(len(data))
	}

}
