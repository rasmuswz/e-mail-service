package model
import (
	"strings"
	"net/mail"
	"io/ioutil"
	"log"
	"fmt"
	"encoding/base64"
	"math"
	"bytes"
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

func encodeRFC2047(String string) string {
	// use mail's rfc2047 to encode any string
	addr := mail.Address{String, ""}
	return strings.Trim(addr.String(), " <>")
}

func (em *EmailImpl) pp() []byte {
	var parser mail.AddressParser = mail.AddressParser{};
	from, _ := parser.Parse(em.headers[EML_HDR_FROM][0]);
	to, _ := parser.Parse(em.headers[EML_HDR_TO][0]);
	body := em.content;

	header := make(map[string]string)
	header[EML_HDR_FROM] = from.String()
	header[EML_HDR_TO] = to.String()
	header[EML_HDR_SUBJECT] = em.headers[EML_HDR_SUBJECT][0];
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\n", k, v)
	}
	message += "\n" + base64.StdEncoding.EncodeToString([]byte(body))

	log.Println("[GetRaw]:\n"+message);

	var src []byte = []byte(body);
	n := len(src);
	var dst []byte = make([]byte,  (int(math.Floor(  float64(n / 3))) + 1) * 4 + 1);
	base64.StdEncoding.Encode(dst,src);

	return dst;
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
	m, err := mail.ReadMessage(r);

	if err != nil {
		return nil;
	}

	content, err := ioutil.ReadAll(m.Body);
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

	for k, v := range headers {
		log.Println("setting header [" + k + "]=" + v);
		mail.headers[k] = []string{v};
	}

	return mail;

}