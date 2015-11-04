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
	var extMap map[string]string = map[string]string{"html": "text/html", "css": "text/css","dart": "application/dart"};

	var path = r.URL.Path;
	if ("" == path || "/" == path) {
		path = "/index.html";
	}


	filename := s.docRoot+path;
	var fi, erro = os.Stat(filename);

	if (erro != nil) {
		println("[Error] "+erro.Error());
		return;
	}

	if (fi.IsDir() == false) {

		println("Serving file: "+filename)
		p, e := loadFile(filename)
		if (e != nil) {
			println("[Error] "+e.Error());
			return;
		}

		var mimeType = "text/plain";
		if strings.Contains(filename,".") {
			var ext = filename[strings.LastIndex(filename, ".") + 1:len(filename)];
			var v,ok = extMap[ext];
			if  ok {
				mimeType = v;
			}
		}

		w.Header().Add("Content-Type",mimeType);

		w.Write(p);

	} else {
		w.Write([]byte("Directory listings not supported."))
	}



}

func main() {

	if (len(os.Args) != 2) {
		println("webserver <doc root path>");
		os.Exit(-1);
	}

	println("Serving from: "+os.Args[1]);
	http.HandleFunc("/", NewServer(os.Args[1]).viewHandler);
	err := http.ListenAndServe(":8080", nil)
	if (err != nil){
		println("[Webserver Error]: "+err.Error());
		os.Exit(-1);
	}
}