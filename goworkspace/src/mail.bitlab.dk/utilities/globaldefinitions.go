package utilities

// ------------------------------------------------------
//
// Port numbers for intra server communication (and HTTPS 443)
//
// ------------------------------------------------------
const (
	CLIENTAPI_SERVICE_PORT = ":443";
	MTA_MAILGUN_SERVICE_PORT = ":31415";
	RECEIVE_BACKEND_LISTENS_FOR_MTA = ":27182";
	SEND_BACKEND_LISTEN_FOR_CLIENT_API = ":16180";
	RECEIVE_BACKEND_LISTEN_FOR_CLIENT_API = ":10301";
	MTA_LISTENS_FOR_SEND_BACKEND = ":10501";
	CLIENT_API_LISTEN_FOR_MTA = ":9999";
)

