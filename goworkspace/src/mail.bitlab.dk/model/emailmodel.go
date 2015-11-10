package model
import (
	"bytes"
	"strings"
	"net/mail"
	"io/ioutil"
	"log"
	"encoding/base64"
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
}

// support the common case of reading the first (and typically only) value
func (em *EmailImpl) GetHeader(name string) string {
	v,ok := em.headers[name];
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
	var stringContent = base64.StdEncoding.D
	buffer.WriteString(em.GetContent());

	return buffer.Bytes();
}

func NewEmailFromBytes(data []byte) Email {
	result := EmailImpl{}

	r := strings.NewReader(string(data));
	m,err := mail.ReadMessage(r);

	if err != nil {
		return nil;
	}

	content,err := ioutil.ReadAll(m.Body);
	result.content = string(content);
	result.headers = m.Header;

	return &result;
}

func NewMail(content string, headers map[string][]string) Email {
	var result = new(EmailImpl);
	result.content = content;
	result.headers = headers;
	return result;
}

func NewMailS(content string, headers map[string]string) Email {

	mail := new(EmailImpl);
	mail.headers = make(map[string][]string);
	mail.content = content;

	for k,v := range headers {
		log.Println("setting header ["+k+"]="+v);
		mail.headers[k] = []string{v};
	}

	return mail;

}