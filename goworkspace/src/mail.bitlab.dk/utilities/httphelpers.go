package utilities
import (
	"net/http"
	"io/ioutil"
	"encoding/base64"
	"encoding/json"
)

//
// RFC 1945 11.1
//
func BasicAuthorization(user string, passphrase string) string {
	var res = "";
	var toEncode = user+":"+passphrase;
	res = " Basic "+base64.StdEncoding.EncodeToString([]byte(toEncode));
	return res;
}

//
// Talk HTTP-GET to url with http headers
//
func GetQuery(url string, http_headers map[string]string) string {

	request, requestErr := http.NewRequest("GET", url, nil);
	if requestErr != nil {
		return "";
	}

	for key, val := range http_headers {
		request.Header.Add(key,val);
	}

	client := http.Client{};
	var response,responseErr = client.Do(request);
	if (responseErr != nil) {
		return "";
	}

	var bytes, bytesErr = ioutil.ReadAll(response.Body);
	if bytesErr != nil {
		return "";
	}

	return string(bytes);
}

func ReadJSonBody(request *http.Request, thing interface{}) error {
	b,e := ioutil.ReadAll(request.Body);
	if e != nil {
		return e
	}

	err := json.Unmarshal(b,thing);
	if err != nil {
		return err
	}

	return nil;
}


func GetString(url string) string {
	return GetQuery(url,make(map[string]string));
}