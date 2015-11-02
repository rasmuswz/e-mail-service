package model
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
	GetHeaders() map[string]string;
	GetContent() string;
}

const (
	EML_HDR_FROM = "From";
	EML_HDR_TO = "To";
	EML_HDR_SUBJECT = "Subject";
	EML_HDR_CONTENT_TYPE = "Content-Type";
	EML_HDR_CONTENT_LENGTH = "Content-Length";
)

type EmailImpl struct {
	header map[string]string;
	content string;
}

func (em *EmailImpl) GetHeaders() map[string]string {
	return em.header;
}

func (em *EmailImpl) GetContent() string {
	return em.content;
}

 // Quick way for generating an {Email} struct from strings.
 //
 // Example: var email = CreateEmail("Hi, this is an email.",
 //                                  "subject","My First Geo-Mail",
 //                                  "to","rwl@geomail.dk");
 //
func NewEmail(content string, headers ... string) Email {
	var result = new(EmailImpl);
	result.content = content;
	result.header = make(map[string]string);
	for i := 0; i < len(headers) / 2; i += 1 {
		var key = headers[i];
		var val = headers[i + 1];
		result.header[key] = val;
	}
	return result;
}