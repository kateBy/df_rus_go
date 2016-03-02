package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"debug/elf"
	"bytes"
)

func loadString(fileName string) map[string]string {
	/*Функция загружает строки из po-файла в виде словаря
	Возможны глюки, т.к. все довольно линейно и топорно*/

	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println("Ошибка при открытии файла", fileName)
	}

	lines := strings.Split(string(buf), "\n")
	max_lines := len(lines) - 1

	var i int
	var pos int
	var msgid string
	var msgstr string
	result := make(map[string]string)

	for {
		if i == max_lines {
			break
		}

		if strings.HasPrefix(lines[i], "msgid") {
			pos = strings.Index(lines[i], "\"") + 1
			msgid = lines[i][pos : len(lines[i])-1]
			i++
			pos = strings.Index(lines[i], "\"") + 1
			msgstr = lines[i][pos : len(lines[i])-1]

			result[msgid] = msgstr
		}
		i++
	}

	return result
}

func checkString(byte_string []byte) string {
	for i := 0; i < len(byte_string); i++ {
		if byte_string[i] < 32 || byte_string[i] > 126 {
			return ""
		}
	}
	return string(byte_string)
}

/* Извлечение hardcoded строк из исполняемого файла,
   все, что похоже на строки */
func extractStrings(fileName string) map[string]uint64 { 

	elfFile, err := elf.Open(fileName)
	if err != nil {
		fmt.Println("Ошибка при открытии", fileName)
	}
	defer elfFile.Close()

	rodata := elfFile.Section(".rodata")
	vaddr := rodata.Addr //Виртуальный адрес начала секции

	data, err := rodata.Data()
	data_length := uint64(len(data))
	var next_zero int = 0
	var checked string
	var i uint64
	result := make(map[string]uint64) //Словарь {слово:виртуальный_адрес}
	
	for i = 0; i < data_length; {
		if data[i] > 31 && data[i] < 127 {
			next_zero = bytes.IndexByte(data[i:], 0)
			if next_zero != -1 && next_zero > 2 {
				checked = checkString(data[i:i+uint64(next_zero)])
				if checked != "" {
					result[checked] = i + vaddr
				}
				i = uint64(next_zero) + i //Т.к. next_zero - относится к срезу data[i:]
			} 
		}
		i++
	}

	return result

}

/* Функция поиска строк-близнецов путём отрезания начала слов и сравнивания со словарём перевода*/
func findGemini(hStrings map[string]uint64, transStrings map[string]string) map[string]uint64 {
	result := make(map[string]uint64)

	var i uint64

	for hStr := range hStrings {
		max_len := uint64(len(hStr) - 1)
		for i = 1; i < max_len; i++ {
			if v, ok := transStrings[hStr[i:]]; ok { //Проверяем, есть ли уменьшенная строка в переводах
			    _ = v //UNUSED
				if v, ok := hStrings[hStr[i:]]; !ok { //Проверяем, нет ли уже такой строки в hardcoded строках
					_ = v //UNUSED
					result[hStr[i:]] = hStrings[hStr] + i
				}
			}
		}
	}

	return result
}