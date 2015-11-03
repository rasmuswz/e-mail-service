package main


import (
	"net/http"
	"io/ioutil"
	"os"
	"strings"
)

type WebServer struct {
	docRoot string
}

func loadFile(filename string) ([]byte, error) {

	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return body,nil;
}

func NewServer(docRoot string) *WebServer {
	var result = new(WebServer);
	result.docRoot = docRoot;
	return result;
}

func (s *WebServer) viewHandler(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path;
	if ("" == path) {
		path = "index.html";
	}

	if (path == "/main.dart") {
		path = "/main.dart.js";
	}

	filename := s.docRoot+path;


	println("Serving file: "+filename)
	p, e := loadFile(filename)
	if (e != nil) {
		println("[Error] "+e.Error());
		return;
	}

	if (strings.Contains(path,".css")) {
		w.Header().Add("Content-Type","text/css");
	}

	w.Write(p);
}

func main() {

	if (len(os.Args) != 2) {
		println("webserver <doc root path>");
		os.Exit(-1);
	}

	println("Serving from: "+os.Args[1]);
	http.HandleFunc("/", NewServer(os.Args[1]).viewHandler);
	http.ListenAndServe(":80", nil)
}