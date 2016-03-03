package main

import (
	"fmt"
	_ "time"
)

const DF_FILENAME = "Dwarf_Fortress"

func main() {
	fmt.Print("Загрузка строк перевода, trans.po ... ")
	translation := loadString("trans.po")
	fmt.Println(len(translation), "пар строк загружено")

	fmt.Print("Извлечение строк из исполняемого файла ... ")
	hardcodedStrings := extractStrings(DF_FILENAME)
	fmt.Println(len(hardcodedStrings), "строк загружено")

	fmt.Print("Поиск строк-близнецов ... ")
	gemini := findGemini(hardcodedStrings, translation)
	fmt.Println(len(gemini), "строк-близнецов найдено")
	
	/*fmt.Print("Отсев строк-близнецов ... ")
	start := time.Now()
	chk := checkGemini(DF_FILENAME, gemini)
	finish := time.Now()	
	fmt.Println(len(chk), "за", finish.Sub(start))*/
	
	//FIXME Объединить chk и hardcodedStrings
	fmt.Println("Поиск перекрестных ссылок в коде ... ")
	xref := findXRef(DF_FILENAME, hardcodedStrings)
	fmt.Println(len(xref))
	

}


