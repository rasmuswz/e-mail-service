//
// A Store of JSon blobs. We limit this to only
// include maps from string to string.
//
// Author: Rasmus Winther Zakarias
//
package backend
import (
	"strconv"
	jsonPackage "encoding/json"
	"strings"
//	"net/http"
	"encoding/base64"
	"mail.bitlab.dk/model"
)


// ------------------------------------------------------------
// Interface JSonStore supports the two operations {GetJSonBlob} and
// {PutJSonBlob}. We do not include removing blobs at this time.
//
// TODO(rwz): Add removal of blobs at some point.
// ------------------------------------------------------------
type JSonStore interface {
	GetJSonBlobs(matching map[string]string) []map[string]string;

	//
	// Insert {blob} into storage and get back an unique
	// identifier for it
	//
	PutJSonBlob(blob map[string]string) uint64;

	//
	// Update an existing item, returns true if items was
	// found and updated. Otherwise we return false.
	//
	UpdJSonBlob(matching map[string]string, blob map[string]string) (uint64,bool);

	//
	// Get by id, having the id lookups happen much quicker
	//
	GetById(id uint64) (map[string]string,bool);
}


// ------------------------------------------------------------
//
// A User Blob stored in the Store.
//
// ------------------------------------------------------------
type UserBlob struct {
	Username  string;
	Password  string;
	SessionId string;
}

func (ths *UserBlob) IsLoggedIn() bool {
	return ths.SessionId != "";
}


func UserBlobNewFull(username , password,  sessionid string) *UserBlob {
	result := new(UserBlob);
	result.Username = username;
	result.Password = password;
	result.SessionId = sessionid;
	return result;
}

func UserBlobNew(username , password string) *UserBlob{
	result := new(UserBlob);
	result.Username = username;
	result.Password = password;
	result.SessionId = "";
	return result;
}

func (ths *UserBlob) ToJSonMap() map[string]string {
	result := make(map[string]string);
	result["Username"] = ths.Username;
	result["Password"] = ths.Password;
	result["SessionId"] = ths.SessionId;
	return result;
}

func UserBlobFromJSonMap(m map[string]string) *UserBlob {
	result := new(UserBlob);
	result.Username = m["Username"];
	result.Password = m["Password"];
	result.SessionId = m["SessionId"];
	return result;
}

func UserBlobNewFromJSonStr(jsonBlob string) *UserBlob {
	blob := new(UserBlob);
	var decoder = jsonPackage.NewDecoder(strings.NewReader(jsonBlob));
	decoder.Decode(blob);
	return blob;
}


// ------------------------------------------------------------------
//
// An MailBox Blob stored in the Store
//
// ------------------------------------------------------------------
type MBoxBlob struct {
	UniqueID string;
	Name     string;
	Username string; // mail box owner
}

func MBoxUniqueName(username, boxname string) string {
	return base64.StdEncoding.EncodeToString([]byte(username+":"+boxname));
}

func NewMBox(username string, boxname string) *MBoxBlob{
	result := new(MBoxBlob);
	result.UniqueID = MBoxUniqueName(username,boxname);
	return result;
}

func (ths *MBoxBlob) ToJSonMap() map[string]string {
	result := make(map[string]string, 2);
	result["Name"] = ths.Name;
	result["Username"] = ths.Username;
	result["UniqueID"] = ths.UniqueID;
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
	Mbox    string; // MBoxBlob.UniqueID.
	Subject string;
	To      string;
	From    string;
	Content string;
	Headers map[string]string;
}

func (ths *EmailBlob) ToJSonMap() map[string]string {
	result := make(map[string]string, 6);
	result["MBox"] = ths.Mbox;
	result["Subject"] = ths.Subject;
	result["To"] = ths.To;
	result["From"] = ths.From;
	result["Content"] = ths.Content;
	for k,v := range ths.Headers {
		result[k] = v;
	}
	return result;
}

func NewEmailBlobFromJSonMap(m map[string]string ) *EmailBlob{
	result := new(EmailBlob);
	result.Mbox = m["MBox"];
	result.Subject = m["Subject"];
	result.To = m["To"];
	result.From = m["From"];
	result.Content = m["Content"];
	return result;
}

func NewEmailBlobForFindingMBox(mbox string) *EmailBlob{
	return NewEmailBlob(mbox,"","","","");
}

func NewEmailBlobFromEmail(email model.Email) *EmailBlob{

	result := new(EmailBlob);
	result.Mbox = model.MBOX_NAME_INBOX;
	result.Subject = email.GetHeader(model.EML_HDR_SUBJECT);
	result.To = email.GetHeader(model.EML_HDR_TO);
	result.From = email.GetHeader(model.EML_HDR_FROM);
	result.Content = email.GetContent();

	return result;

}

func NewEmailBlob(mbox, subject, to, from,content string ) *EmailBlob {
	result := new(EmailBlob);
	result.Mbox = mbox;
	result.Subject = subject;
	result.To = to;
	result.From = from;
	result.Content = content;
	return result;
}

func EmailBlobFromJSonMap(m map[string]string) *EmailBlob {
	result := new(EmailBlob);
	result.Mbox = m["MBox"];
	result.Subject = m["Subject"];
	result.To = m["To"];
	result.From = m["From"];
	result.Content = m["Content"];
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

func checkMatch(db map[string]string, m map[string]string) bool {

	for k, v := range m {
		dbval, ok := db[k];

		if ok == false {
			return false
		};

		if (strings.Compare("",v) != 0) {
			if (strings.Compare(dbval, v) != 0) {
				return false;
			}
		}
	}

	return true;
}


func (ths *MemoryJsonStore) findBlob(j map[string]string, from uint64) (uint64,bool) {

	for id,databaseItem := range ths.memory {
		if id >= from {
			if (checkMatch(databaseItem,j)) {
				return id,true;
			}
		}
	}
	return 0,false;
}

func (ths * MemoryJsonStore) UpdJSonBlob(matching map[string]string,
										 blob map[string]string) (uint64,bool) {
	var id,ok = ths.findBlob(matching,0);
	if ok {
		for k,v := range blob {
			ths.memory[id][k] = v;
		}
		return id,true;
	}
	return 0,false;
}


//
// We scan the whole DB in search for all blobs that has matching
// keys and values from {matching}
//
func (ths *MemoryJsonStore) GetJSonBlobs(matching map[string]string) []map[string]string {
	var res = make([]map[string]string, 0);
	for _, databaseItem := range ths.memory {
		allMatchingKeysFound := checkMatch(databaseItem, matching);
		if allMatchingKeysFound == true {
			res = append(res,databaseItem);
		}
	}
	return res;
}

//
// We fix a new unique id and insert the blob
//
func (ths *MemoryJsonStore) PutJSonBlob(jsonblob map[string]string) uint64 {
	id := ths.idpool;
	ths.idpool += 1;
	ths.memory[id] = jsonblob;
	jsonblob[MEMORY_JSON_STORE_UID_KEY] = strconv.FormatUint(id, 10);
	return id;
}


func (ths *MemoryJsonStore) GetById(id uint64) (map[string]string,bool) {
	v,ok :=  ths.memory[id];
	return v,ok;
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

	// TODO(rwz): Marshal {jsonblob}

	// TODO(rwz): Send serialised PubJSonStore command to {ths.clientApiEndPoint}

	// TODO(rwz): Deserialise uint64

	return 0;
}


func (ths *ProxyStore) GetJSonBlobs(matching map[string]string) []map[string]string {

	// TODO(rwz): Marshal {matching} and send {GetJSonBlobs} to end point

	//TODO(rwz): Deserialise a list of entries that match

	return make([]map[string]string, 1);
}

func (ths *ProxyStore) UpdJSonBlob(matching map[string]string, blob map[string]string) (uint64,bool) {

	// TODO(rwz): Marshal {matching} and {blob}

	//TODO(rwz): Deserialise response from server

	return 0,false;
}

func (ths *ProxyStore) GetById(id uint64) (map[string]string,bool) {

	// TODO(rwz): Serialise {id} to the server

	//TODO(rwz): Deserialise the resulting map

	return nil,false;
}