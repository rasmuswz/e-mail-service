//
//
// This file describes how we represent {Email}s in memory
// and on the Wire.
//
// Author: Rasmus Winther Zakarias.
//

package model
import (
	"io/ioutil"
	"log"
	"encoding/base64"
	"bytes"
	"io"
	"encoding/json"
	"encoding/hex"
)
/**
 *
 * Application wide definitions of business logic objects. In this case
 * the primary model object is an e-mail.
 *
 * Author: Rasmus Winther Zakarias
 */

/*
 * we adopt where a simple model for e-mails, an e-mail is
 * represented by its headers and a single string {content}
 * which is its contents.
 */
type Email interface {
	GetHeaders() map[string][]string;
	// Get single header
	GetHeader(key string) string;
	GetContent() string;
	GetRaw() []byte;
	// session id of user sending mail
	GetSessionId() string;

}

// ----------------------------------------------------
//
// WireEmail is the serialized form of an email in this
// application. We use json.Marshal and json.Unmarshal
// to put and get WireEmail objects on/of the wire.
//
// On the wire content is encoded as base64
//
// ----------------------------------------------------
type WireEmail struct {
	Headers map[string]string;
	Content string;
}

func (ths *WireEmail) To() string {
	return ths.Headers[EML_HDR_TO];
}

func (ths *WireEmail) SetTo(to string) {
	ths.Headers[EML_HDR_TO] = to;
}

//
// Convert an {WireEmail} to {Email}. Data is copied
// {ths} remains intact.
//
func (ths * WireEmail) ToEmail(sessionId ... string) Email {
	result := new(EmailImpl);

	str,strErr := base64.StdEncoding.DecodeString(ths.Content);
	if strErr != nil {
		return nil;
	}
	result.content = string(str);
	result.headers = make(map[string][]string);
	for k,v := range ths.Headers {
		result.headers[k] = []string{v};
	}

	if len(sessionId) == 1 {
		result.sessionId = sessionId[0];
	}

	return result;
}

//
// Vice versa
//
func NewWireEmail(mail Email) *WireEmail{
	result := new(WireEmail);
	result.Headers = make(map[string]string);
	for k,v := range mail.GetHeaders() {
		if len(result.Headers[k]) > 0 {
			result.Headers[k] = v[0];
		}
	}
	result.Content = base64.StdEncoding.EncodeToString([]byte(mail.GetContent()));
	return result;
}

//
// Read an Email of the wire.
//
func NewWireEMailFromReader(reader io.Reader) *WireEmail {
	result := new(WireEmail);

	data,dataErr := ioutil.ReadAll(reader);
	if dataErr != nil {
		log.Println("[WireEmail] Reading data stream failed. "+dataErr.Error());
		return nil;
	}

	resultErr := json.Unmarshal(data,result);
	if resultErr != nil {
		log.Println("[WireEmail] "+resultErr.Error()+"\nOffending data:\n"+hex.Dump(data));
		return nil;
	}

	return result;
}

// ---------------------------------------------------
//
//
// Email Implementation
//
// We provide several way of getting an Email object as
// it has been needed at some point.
//
//
// ---------------------------------------------------
const (
	EML_HDR_FROM = "From";
	EML_HDR_TO = "To";
	EML_HDR_SUBJECT = "Subject";
	EML_HDR_CONTENT_TYPE = "Content-Type";
	EML_HDR_CONTENT_LENGTH = "Content-Length";
// ---------------------------------------
	MBOX_NAME_INBOX = "INBOX";
	MBOX_NAME_SENT = "Sent";
)

type EmailImpl struct {
	headers map[string][]string;
	content string;
	sessionId string;
}

func (ths *EmailImpl) SetSessionId(id string) {
	ths.sessionId = id;
}

func (ths *EmailImpl) GetSessionId() string {
	return ths.sessionId;
}

// support the common case of reading the first (and typically only) value
func (em *EmailImpl) GetHeader(name string) string {
	v, ok := em.headers[name];
	if ok == false {
		return "";
	}
	return v[0];
}


func (em *EmailImpl) GetHeaders() map[string][]string {
	return em.headers;
}

func (em *EmailImpl) GetContent() string {
	return em.content;
}

func (em *EmailImpl) GetRaw() []byte {
	buffer := bytes.NewBuffer(nil);
	for k, v := range em.GetHeaders() {
		buffer.WriteString(k + ":" + v[0] + "\n");
	}
	buffer.WriteString("\n");
	buffer.WriteString(em.GetContent());

	return buffer.Bytes();
}

func NewMail(content string, headers map[string][]string) Email {
	var result = new(EmailImpl);
	result.content = content;
	result.headers = headers;
	return result;
}

func NewMailSimpleHeaders(content string, headers map[string]string) Email {

	mail := new(EmailImpl);
	mail.headers = make(map[string][]string);
	mail.content = content;

	for k, v := range headers {
		log.Println("setting header [" + k + "]=" + v);
		mail.headers[k] = []string{v};
	}

	return mail;
}