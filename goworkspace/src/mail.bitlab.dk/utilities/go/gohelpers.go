package goh
import "strconv"


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