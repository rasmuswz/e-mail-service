/**
 *
 *  MailModel.
 *  ----------
 *
 *  The client side model of our Geo-Mail data-model.
 *
 *
 * Author: Rasmus Winter Zakarias
 */
import 'dart:html';
String GetQuery(String url) {

  HttpRequest req = new HttpRequest();
  req.open("GET","/login?username=",async: false);
  req.send();

}

/**
 * Represent an open connection to the Server implementing the Client API component.
 *
 *
 *
 */
class GeoMailConnection {

  String token;

  GeoMailConnection(String path) {

  }


  bool login(String username, String password){
    HttpRequest loginRequest = new HttpRequest();
    //loginRequest.open("GET","/go.login",async: false,user: username, password: password);

    return true;
//    if (loginRequest.status == 200) {
//      return true;
//    } else {
//      return false;
//    }

  }

}

class Email {
  String _from;
  String _to;
  String _subject;
  String _location;
  String _content;

  Email.fromMap(Map<String,String> json) {
    this._from = json['from'];
    this._to = json['to'];
    this._subject = json['subject'];
    this._location = json['location'];
  }

  Email(String from, String subject) {
    this._from = from;
    this._subject = subject;

  }

  get From => _from;
  get To => _to;
  get Subject => _subject;
  get Location => _location;
  get Content => _content;
  void setContent(String content) {
    this._content = content;
  }
}

class GeoMailDataModel {

  GeoMailDataModel() {

  }

  // LogIn to the api
  bool login(String username, String password)  {
    return true;
  }


  List<Email> loadEmailList(int offset, int count) {
    List<Email> mails = [];
    mails.add(new Email("rasmuswz@gmail.com","Testing Geo Mail for the first time."));
    mails.add(new Email("steffen@uber.com","Don't sped too much time on UX/UI."));
    mails.add(new Email("szakarias@gmail.com","Hi Congrats"));
    mails[0].setContent("Hi GeoMail,\nThis is a test\nSincerely Rasmus");
    mails[1].setContent("Show me some cool Go-code");
    mails[2].setContent("Flot site du har her !.");
    return mails;
  }
}

