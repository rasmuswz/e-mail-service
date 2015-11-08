package model
import (
	"bytes"
	"strings"
	"net/mail"
	"io/ioutil"
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


 // Quick way for generating an {Email} struct from strings.
 //
 // Example: var email = CreateEmail("Hi, this is an email.",
 //                                  "subject","My First Geo-Mail",
 //                                  "to","rwl@geomail.dk");
 //
func NewEmailFlattenHeaders(content string, headers ... string) Email {
	var result = new(EmailImpl);
	result.content = content;
	result.headers = make(map[string][]string);
	for i := 0; i < len(headers) / 2; i += 1 {
		var key = headers[i];
		var val = headers[i + 1];
		result.headers[key] = []string{val};
	}
	return result;
}

func NewMail(content string, headers map[string][]string) Email {
	var result = new(EmailImpl);
	result.content = content;
	result.headers = headers;
	return result;
}