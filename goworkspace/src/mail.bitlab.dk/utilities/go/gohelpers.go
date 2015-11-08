package goh
import (
	"strconv"
	"os"
	"bufio"
)


func IntToStr(i int) string {
	var s = strconv.Itoa(i);
	return s;
}

func StrToInt(s string) int {
	var i,e = strconv.Atoi(s);

	if (e != nil) {
		i = 0;
	}

	return i;
}

func ReadLine(prompt string) string {
	print(prompt);

	line,_,err := bufio.NewReader(os.Stdin).ReadLine();
	if err != nil {
		return "";
	}

	return string(line);
}