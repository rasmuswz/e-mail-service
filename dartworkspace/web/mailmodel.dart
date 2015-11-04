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
import 'dart:async';

String GetQuery(String path) {

  HttpRequest req = new HttpRequest();
  req.open("GET",path,async: false);
  req.send();

  if (req.status== 200) {
    return req.responseText;
  } else {
    return null;
  }

}

/**
 * Represent an open connection to the Server implementing the Client API component.
 */
class GeoMailConnection {

  String _path;
  Timer _watchDog;
  bool _alive;
  bool _previousAlive;
  List<Completer<String>> stateListeners;


  GeoMailConnection(String path) {
    this._path = path;
    _previousAlive = false;
    this._watchDog = new Timer.periodic(new Duration(seconds: 2),check);

  }

  void setAlive(bool v) {
    this._alive = v;
  }

  Future<String> ListenForState() {
    Completer<String> c = new Completer<String>();
    return c.future;
  }

  void check(Timer t) {
    String resp = GetQuery(_path+"/alive");
    _alive = resp != null;
    if (!_alive) {
      print("Geo Mail Connection is down...");
    }

    if (_alive != _previousAlive) {
      stateListeners.forEach( (completer) {
          completer.complete(_alive ? "going up" : "down");
      });
    }
  }

  get Alive => _alive;

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

  String basicAuth;
  GeoMailConnection connection;
  List<Future<String>> stateListeners;

  GeoMailDataModel(GeoMailConnection connection) {
    this.connection = connection;
    this.stateListeners = [];
  }

  void logout() {
    this.basicAuth = null;
  }

  Future<String> ListenForConnectionState() {
    return this.connection.ListenForState();
  }

  // LogIn to the api
  bool login(String username, String password)  {
    this.basicAuth = window.btoa(username+":"+password);
    return true;
  }


  get IsLoggedIn => basicAuth != null;


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

  String getVersion() {
    String verstr = GetQuery("version.txt");
    if (verstr == null) {
      return " no version ";
    }
    return verstr;
  }
}

