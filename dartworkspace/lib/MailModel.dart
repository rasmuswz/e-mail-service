/**
 *
 * The model of our MVC meta-pattern is implemented here.
 * Our model maintains a list of e-mail and it can do the following:
 *
 * - getContent - of an e-mail
 *
 * - listEmails - list e-mails in a range
 *
 * Author: Rasmus Winther Zakarias
 *
 */
 abstract class ClientAPI {

   Error ListEmails(int offset, int length);

   Error SendEmail(Map<String,String> header, String content);

}

/**
 * This is what makes an email on the client.
 */
class Email {
  Map<String,String> headers;
  String content;
  Email(this.headers,this.content);
}

/**
 * Get an e-mail
 */
class MailModel {

  ClientAPI api;

  MailModel(this.api);

  MyError ListMailBoxes(List<String> boxNames) {
    boxNames.add("INBOX");
    return null;
  }

  MyError ListEmails(int offset, int length, List<Email> mailsOut) {
    return new MyError("Not implemented yet");
  }

  MyError SendEmail(Email m) {
    return new MyError("Not implemented yet");
  }
}



class HttpClientApi implements ClientAPI {

  String url;

  HttpClientApi(this.url) {

  }

}
