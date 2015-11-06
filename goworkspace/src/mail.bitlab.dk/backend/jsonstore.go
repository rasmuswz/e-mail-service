//
// A Store of JSon blobs. We limit this to only
// include maps from string to string.
//
// Author: Rasmus Winther Zakarias
//
package backend
import (
	"strconv"
	"encoding/json"
	"strings"
	"net/http"
)


// ------------------------------------------------------------
// Interface JSonStore supports the two operations {GetJSonBlob} and
// {PutJSonBlob}. We do not include removing blobs at this time.
//
// TODO(rwz): Add removal of blobs at some point.
// ------------------------------------------------------------
type JSonStore interface {
	GetJSonBlob(matching map[string]string) []map[string]string;
	PutJSonBlob(jsonblob map[string]string) uint64;
}


// ------------------------------------------------------------
//
// A User Blob stored in the Store.
//
// ------------------------------------------------------------
type UserBlob struct {
	Username  string;
	Password  string;
	Location  string;
	SessionId string;
}

func (ths *UserBlob) IsLoggedIn() {
	return ths.SessionId != "";
}

func (ths *UserBlob) ToJSonMap() map[string]string {
	result := make(map[string]string);
	result["Username"] = ths.Username;
	result["Password"] = ths.Password;
	result["Location"] = ths.Location;
	result["SessionId"] = ths.SessionId;
	return result;
}

func UserBlobFromJSonMap(m map[string]string) *UserBlob {
	result := new(UserBlob);
	result.Username = m["Username"];
	result.Password = m["Password"];
	result.Location = m["Location"];
	result.SessionId = m["SessionId"];
	return result;
}

func NewUserBlob(jsonBlob string) *UserBlob {
	blob := new(UserBlob);
	var decoder = json.NewDecoder(strings.NewReader(json));
	decoder.Decode(blob);
	return blob;
}


// ------------------------------------------------------------------
//
// An MailBox Blob stored in the Store
//
// ------------------------------------------------------------------
type MBoxBlob struct {
	Name string;
	Username string; // mail box owner
}

func (ths *MBoxBlob) ToJSonMap() map[string]string {
	result := make(map[string]string,2);
	result["Name"] = ths.Name;
	result["Username"] = ths.Username;
	return result;
}


func MBoxBlobFromMap(m map[string]string) *MBoxBlob {
	result := new(MBoxBlob);
	result.Name = m["Name"];
	result.Username = m["Username"];
	return result;
}

// ------------------------------------------------------------------
//
// An E-mail Blob stored in the Store
//
// ------------------------------------------------------------------
type EmailBlob struct {
	Mbox    string;
	Subject string;
	To      string;
	From    string;
	Content string;
	Uid     string;
}

func (ths *EmailBlob) ToJSonMap() map[string]string {
	result := make(map[string]string,6);
	result["MBox"] = ths.Mbox;
	result["Subject"] = ths.Subject;
	result["To"] = ths.To;
	result["From"] = ths.From;
	result["Content"] = ths.Content;
	result["Uid"] = ths.Uid;
	return result;
}


func EmailBlobFromJSonMap(m map[string]string) *EmailBlob {
	result := new(EmailBlob);
	result.Mbox = m["MBox"];
	result.Subject = m["Subject"];
	result.To = m["To"];
	result.From = m["From"];
	result.Content = m["Content"];
	result.Uid = m["Uid"];
	return result;
}


// ---------------------------------------------------------------
//
// MemoryJsonStore a JSonStore.
//
//
// Preliminary in-memory implementation. An enduring solution would
// require a Database like MySQL or High-Scale-File-Systems like
// https://wiki.apache.org/hadoop/HDFS
//
// TODO(rwz): Implement enduring JSon store
// ---------------------------------------------------------------
const MEMORY_JSON_STORE_UID_KEY = "__uid__";
type MemoryJsonStore struct {
	memory map[uint64]map[string]string;
	idpool uint64;
}

//
// We scan the whole DB in search for all blobs that has matching
// keys and values from {matching}
//
func (ths *MemoryJsonStore) GetJSonBlob(matching map[string]string) []map[string]string {
	var res = make([]map[string]string, 0);
	//var ids = make([]uint64,0);
	for id, val := range ths.memory {
		for key, v := range val {
			var allMatch = true;
			for skey, sv := range matching {
				if (key != skey && v != sv) {
					allMatch = false;
					break;
				}
			}
			if (allMatch) {
				res = append(res, ths.memory[id]);
				//	ids = append(ids,id);
			}
		}
	}
	return res;
}

//
// We fix a new unique id and insert the blob
//
func (ths *MemoryJsonStore) PutJSonBlob(jsonblob map[string]string) uint64 {
	id := ths.idpool + 1;
	ths.idpool += 1;
	ths.memory[id] = jsonblob;
	jsonblob[MEMORY_JSON_STORE_UID_KEY] = strconv.FormatUint(id, 10);
	return id;
}


func NewMemoryStore() JSonStore {
	var result *MemoryJsonStore = new(MemoryJsonStore);
	result.memory = make(map[uint64]map[string]string);
	result.idpool = 0;
	return result;
}


// ---------------------------------------------------------
//
// Proxy Store a JSonStore
//
// In our current solution there is only one physical store
// (like MemoryJsonStore above) where this are stored. With
// multiple nodes (like mail0.bitlab.dk and mail1.bitlab.dk)
// on needs to use the others store for consistent service.
//
// A proxy store takes an URL of ClientApi https entry point
// and forwards Get and Put commands on behalf of this client.
//
// ---------------------------------------------------------
type ProxyStore struct {
	clientApiEndPoint string;

}

func NewProxyStore(endpoint string) JSonStore {
	result := new(ProxyStore);
	result.clientApiEndPoint = endpoint;
	return result;
}


func (ths *ProxyStore) PutJSonBlob(jsonblob map[string]string) uint64 {

	jsonstr,jsonStrErr := json.Marshal(jsonblob);

	if jsonStrErr != nil {
		// TODO(rwz): We need to log this somewhere
	}

	strings.NewReader()
	http.NewRequest("GET",ths.clientApiEndPoint,)

	json.NewEncoder()

}


func (ths *ProxyStore) GetJSonBlob(matching map[string]string) []map[string]string {

}