package main

import (
	"fmt"
)

func main() {
	fmt.Print("Загрузка строк перевода, trans.po ... ")
	translation := loadString("trans.po")
	fmt.Println(len(translation), "пар строк загружено")
	
	fmt.Print("Извлечение строк из исполняемого файла ... ")
	hardcodedStrings := extractStrings("Dwarf_Fortress")
	fmt.Println(len(hardcodedStrings), "строк загружено")
	
    fmt.Print("Поиск строк-близнецов ... ")
	gemini := findGemini(hardcodedStrings, translation)
	fmt.Println(len(gemini), "строк-близнецов найдено")
	

}


