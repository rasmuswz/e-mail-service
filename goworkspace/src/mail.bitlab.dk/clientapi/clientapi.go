package clientapi


import (
	"net/http"
	"io/ioutil"
	"os"
	"strings"
	"mail.bitlab.dk/mtacontainer"
	"strconv"
	"log"
)

type ClientAPI struct {
	docRoot string
	mtacontainer.HealthService;
	events chan mtacontainer.Event;
	port int;
}


func (a *ClientAPI) GetEvent() chan mtacontainer.Event {
	return a.events;
}


// --------------------------------------------------------
//
// Create a Server for the website and ClientApi
//
// --------------------------------------------------------
func NewServer(docRoot string, port int) *ClientAPI {
	var result = new(ClientAPI);
	result.docRoot = docRoot;
	result.events = make(chan mtacontainer.Event);
	result.port = port;
	go result.serve();
	return result;
}

func (a *ClientAPI) serve() {

	var mux = http.NewServeMux();
	mux.HandleFunc("/api.go",a.clientApiHandler);
	mux.HandleFunc("/",a.viewHandler);

	var addr = ":"+strconv.Itoa(a.port);
	a.events <- mtacontainer.NewEvent(mtacontainer.EK_OK,error.Error("Serving on port: "+addr));

	err := http.ListenAndServeTLS(addr,"cert.pem","key.pem",mux);

	if (err != nil){
		log.Fatalln("[ClientApi, Error] "+err.Error());
	}
}

func (a *ClientAPI) clientApiHandler(w http.ResponseWriter, r *http.Request) {

}

func loadFile(filename string) ([]byte, error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return body, nil;
}

func (s *ClientAPI) viewHandler(w http.ResponseWriter, r *http.Request) {
	var extMap map[string]string = map[string]string{"html": "text/html", "css": "text/css", "dart": "application/dart"};

	var path = r.URL.Path;
	if ("" == path || "/" == path) {
		path = "/index.html";
	}


	filename := s.docRoot + path;
	var fi, erro = os.Stat(filename);

	if (erro != nil) {
		println("[Error] " + erro.Error());
		return;
	}

	if (fi.IsDir() == false) {

		println("Serving file: " + filename)
		p, e := loadFile(filename)
		if (e != nil) {
			println("[Error] " + e.Error());
			return;
		}

		var mimeType = "text/plain";
		if strings.Contains(filename, ".") {
			var ext = filename[strings.LastIndex(filename, ".") + 1:len(filename)];
			var v, ok = extMap[ext];
			if ok {
				mimeType = v;
			}
		}

		w.Header().Add("Content-Type", mimeType);

		w.Write(p);

	} else {
		w.Write([]byte("Directory listings not supported."))
	}
}