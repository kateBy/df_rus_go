package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main(){
	strings := loadString("trans.po")
	fmt.Println(strings)
}

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
			pos = strings.Index(lines[i],"\"") + 1
			msgid = lines[i][pos:len(lines[i])-1]
			i++
			pos = strings.Index(lines[i],"\"") + 1
			msgstr = lines[i][pos:len(lines[i])-1]
			
			result[msgid] = msgstr
			
		}
		
		i++
	}
		
	return result
}