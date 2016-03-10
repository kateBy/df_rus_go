package main

import (
	"bufio"
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
)

const (
	SIZE_UINT32      = 4
	NOT_FOUND        = -1
	CACHE_TXT        = "cache.txt"
	GEMINI_CACHE_TXT = "gemini_cache.txt"
	SECTION_RODATA   = ".rodata"
	SECTION_TEXT     = ".text"
)

//Байты, которые должны идти перед указателем на строку, все остальные, кроме mov esp - мусор
var GOOD_BITS = []byte{0xb8, //mov eax, offset
	                   0xb9, //mov ecx, offset
	                   0xba, //mov edx, offset
	                   0xbb, //mov ebx, offset
	                   0xbd, //mov ebp, offset
	                   0xbe, //mov esi, offset
	                   0xbf} //mov edi, offset

/* Функция загружает строки из po-файла в виде словаря
   Возможны глюки, т.к. все довольно линейно и топорно*/
func loadString(fileName string) map[string]string {

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

/* Проверка строк на валидные символы*/
func checkString(byte_string []byte) string {
	const SQUARE_COMMONS = 91 //Код символа [
	for i := 0; i < len(byte_string); i++ {
		if byte_string[i] < 32 || byte_string[i] > 126 || byte_string[0] == SQUARE_COMMONS {
			return ""
		}
	}
	return string(byte_string)
}

/* Извлечение hardcoded строк из исполняемого файла,
   все, что похоже на строки */
func extractStrings(fileName string) map[string]uint32 {

	elfFile, err := elf.Open(fileName)
	if err != nil {
		fmt.Println("Ошибка при открытии", fileName)
	}
	defer elfFile.Close()

	rodata := elfFile.Section(SECTION_RODATA)
	vaddr := uint32(rodata.Addr) //Виртуальный адрес начала секции

	data, err := rodata.Data()
	data_length := uint32(len(data))
	var next_zero int = 0
	var checked string
	var i uint32
	result := make(map[string]uint32) //Словарь {слово:виртуальный_адрес}

	for i = 0; i < data_length; {
		if data[i] > 31 && data[i] < 127 {
			next_zero = bytes.IndexByte(data[i:], 0)
			if next_zero != NOT_FOUND && next_zero > 2 {
				checked = checkString(data[i : i+uint32(next_zero)])
				if checked != "" {
					result[checked] = i + vaddr
				}
				i = uint32(next_zero) + i //Т.к. next_zero - относится к срезу data[i:]
			}
		}
		i++
	}

	return result

}

/* Функция поиска строк-близнецов путём отрезания начала слов и сравнивания со словарём перевода*/
func findGemini(hStrings *map[string]uint32, transStrings *map[string]string) map[string]uint32 {
	result := make(map[string]uint32)

	var i uint32

	for hStr := range *hStrings {
		max_len := uint32(len(hStr) - 2)
		for i = 1; i < max_len; i++ {
			if _, ok := (*transStrings)[hStr[i:]]; ok { //Проверяем, есть ли уменьшенная строка в переводах
				if _, ok := (*hStrings)[hStr[i:]]; !ok { //Проверяем, нет ли уже такой строки в hardcoded строках
					result[hStr[i:]] = (*hStrings)[hStr] + i
				}
			}
		}
	}

	return result
}

func goCheck(job chan string, out chan map[string]uint32, gemini *map[string]uint32, buf *[]byte) {
	var link_byte int
	result := make(map[string]uint32)
	bs := make([]byte, SIZE_UINT32)

	for {
		if j, more := <-job; more {
			binary.LittleEndian.PutUint32(bs, (*gemini)[j])
			link_byte = bytes.Index(*buf, bs)

			if link_byte != NOT_FOUND {
				result[j] = binary.LittleEndian.Uint32((*buf)[link_byte : link_byte+SIZE_UINT32])
			}

		} else {
			out <- result
			return
		}

	}

}

/* Функция проверки того, что найденные строки-"близнецы" есть в вызовах строк */
func checkGemini(df_filename string, gemini map[string]uint32) map[string]uint32 {
	result := make(map[string]uint32)
	runtime.GOMAXPROCS(runtime.NumCPU())

	elfFile, err := elf.Open(df_filename)
	if err != nil {
		fmt.Println("Ошибка при открытии", df_filename)
	}
	defer elfFile.Close()

	code := elfFile.Section(SECTION_TEXT)
	buf, _ := code.Data()

	procs := runtime.NumCPU()

	job := make(chan string)            //Канал передачи заданий на обработку
	out := make(chan map[string]uint32) //Канал получения результатов

	for i := 0; i < procs; i++ {
		go goCheck(job, out, &gemini, &buf)
	}

	for j := range gemini {
		job <- j
	}

	close(job) //Обязательно!, иначе DEAD LOCK

	var results []map[string]uint32

	for i := 0; i < procs; i++ {
		results = append(results, <-out)
	}

	file, _ := os.Create(GEMINI_CACHE_TXT)
	writer := bufio.NewWriter(file)

	for i := 0; i < len(results); i++ {
		for res := range results[i] {
			writer.WriteString(fmt.Sprintf("%s[*|*]%d\n", res, results[i][res]))
			result[res] = results[i][res] //Объединяем результаты работы потоков
		}
	}

	writer.Flush()
	file.Close()

	return result
}

/* Асинхронный поиск всех включений указателей на строки в исполняемом файле*/
func goFindXRef(job chan string, out chan map[string][]uint32, buf *[]byte, words *map[string]uint32, section_offset uint32) {
	result := make(map[string][]uint32)
	target := make([]byte, SIZE_UINT32)
	var found int
	var lastFound int
	var res []uint32

	for {
		if j, more := <-job; more {
			lastFound = 0
			found = 0
			res = nil

			binary.LittleEndian.PutUint32(target, (*words)[j])
			//Цикл поиска в коде
			for found != NOT_FOUND {
				found = bytes.Index((*buf)[lastFound:], target)
				if found != NOT_FOUND {
					res = append(res, uint32(found+lastFound)) //+vaddr)
					lastFound = found + lastFound + SIZE_UINT32
				}
			} //--Цикл поиска в коде

			if len(res) != 0 {
				result[j] = res
			}

		} else {
			out <- result
			return
		}
	}
}

/* Поиск перекрестных ссылок */
func findXRef(df_filename string, words *map[string]uint32) map[string][]uint32 {

	elfFile, _ := elf.Open(df_filename)

	section_text := elfFile.Section(SECTION_TEXT)
	//vaddr := uint32(rodata.Addr) //Виртуальный адрес начала секции
	section_offset := uint32(section_text.Offset)

	buf, _ := section_text.Data()
	elfFile.Close()

	job := make(chan string)              //Задания
	out := make(chan map[string][]uint32) //Результаты работы потоков

	procs := runtime.NumCPU()
	runtime.GOMAXPROCS(procs)

	for i := 0; i < procs; i++ {
		go goFindXRef(job, out, &buf, words, section_offset)
	}

	for j := range *words {
		job <- j
	}

	close(job) //Важно! DEADLOCK

	var results []map[string][]uint32

	//Собираем результаты
	for i := 0; i < procs; i++ {
		results = append(results, <-out)
	}

	file, _ := os.Create(CACHE_TXT)
	writer := bufio.NewWriter(file)
	result := make(map[string][]uint32)

	for i := 0; i < len(results); i++ {
		for res := range results[i] {
			writer.WriteString(fmt.Sprintf("%s<%v>\n", res, results[i][res]))
			result[res] = results[i][res] //Объединяем результаты работы потоков
		}
	}

	writer.Flush()
	file.Close()

	const ESP byte = 0xc7

	bug_addr := make(map[uint32]bool)
	var bit byte

	fmt.Println(len(result))

	for word := range result {
		for _, offset := range result[word] {
			bit = buf[offset]
			if bytes.IndexByte(GOOD_BITS, bit) != NOT_FOUND {
				if buf[offset-4] != ESP { //Проверяем может быть это mov dword ptr esp
					if buf[offset-3] != ESP { //Проверяем может быть это mov  word ptr esp
						//bug_addr = append(bug_addr, offset)
						bug_addr[offset] = true
					}
				}
			}
		}
	}

	//fmt.Println("BUGS:", len(bug_addr))

	return result
}
