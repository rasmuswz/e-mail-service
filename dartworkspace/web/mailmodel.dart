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
import 'dart:convert';
import 'geoconnection.dart';
import 'main.dart';

/**
 * The Email Model element represents the client side view of an e-mail.
 *
 * We can create it from a JSON map
 *
 * We can Serialise an mail to a JSON map
 *
 * We can acquire its components
 */
class Email {
  String _from;
  String _to;
  String _subject;
  String _location;
  String _content;
  GeoMailModel model;

  Email.fromMap(Map<String, String> json) {
    this._from = json['from'];
    this._to = json['to'];
    this._subject = json['subject'];
    this._location = json['location'];
  }

  Map<String, String> toMap() {
    Map<String, String> json = new Map<String, String>();
    String h = JSON.encode({"From": this._from,"To": this._to, "Subject": this._subject, "Location": this._location});
    print("headers: "+h);
    json["Headers"] = h;
    json["Content"] = this._content;
    return json;
  }

  String toJson() {
    return """{"Headers":{"From":"${this._from}","Subject":"${this._subject}","To":"${this._to}"},"Content":"${this._content}"}""";
    //return "{\"Headers\": {\"To\":\""+_to+"\",\"From\":\""+_from+"\":"
  }

  Email(this._from,this._subject) {

  }

  Email.WithModel(this.model,this._to, this._subject) {
    this._from = model.Username;
    this._location = "here";
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
const int CHECK_INTERVAL_S = 2;
/**
 * A Geo list is represented by its name.
 */
class GeoList {
  String _listName;
  int _friendsCount;
  ClientAPI connection;

  GeoList(this._listName,this.connection) {
    new Timer(new Duration(seconds: CHECK_INTERVAL_S),() => this._updateThisGeoList());
  }

  GeoList.NewFromMap(Map<String, String> map, this.connection) {
    this._listName = map["name"];
    new Timer(new Duration(seconds: 2),() => this._updateThisGeoList());
  }

  /**
   * Name
   */
  get Name => _listName;

  /**
   * Every
   */
  void _updateThisGeoList() {
    count = connection.updateGeoList(_listName, this.sessionId);
  }

  /**
   * Get the number of friends logged in with Geographical Locations
   * putting them on this list.
   */
  get FriendsOnList => _friendsCount;

}

/**
 * The Geo Mail Model implementing the core-business logic
 */
class GeoMailModel {

  String basicAuth;
  ClientAPI connection;
  List<Future<String>> stateListeners;
  String session;
  ViewController view;
  String _username;


  /**
   * With a connection to the ClientAPI-server we
   * create a GeoMailModel.
   */
  GeoMailModel(ClientAPI connection) {
    this.connection = connection;
    this.stateListeners = [];
    _username = "";
  }

  get Username => _username;

  /**
   * Allow the view controllers to get notified
   * if the server connection is down.
   */
  void _listenForState() {
    ViewController view = this.view;
    this.connection.listenForState((b) {
      if (b) {
        view.connectionUp();
      } else {
        view.connectionDown();
      }
    });
  }

  /**
   * Logout the user out forgetting his credentials and session id.
   */
  void logOut() {
    this.connection.doLogout(this.session);

  }

  /**
   * This model may need to take action on the view if
   * the server goes down or new users logs in elsewhere.
   */
  void setView(ViewController view) {
    this.view = view;
    this._listenForState();
  }

  /**
   * Perform login using RFC 1945 Basic Authorization.
   */
  Future<bool> logIn(String username, String password) {
    this.basicAuth = null;
    String auth = window.btoa(username + ":" + password);
    Completer<bool> c = new Completer<bool>();
    view.setSystemMessage("We are processing your location data and verifys your identity...");
    var f = window.navigator.geolocation.getCurrentPosition(enableHighAccuracy: false,
    timeout: new Duration(minutes: 2))..then( (position) {
      print("position acquired");
      String location = position.toString();
      String sessionId = connection.doLogin(auth,location);
      if (sessionId != null) {
        this.basicAuth = auth;
        this.session = sessionId;
        this._username = username;
        c.complete(true);
      } else {
        view.setSystemMessage("Login failed");
        c.complete(false);
      }
    });


  return c.future;
  }


  /** Are we logged in or not */
  get IsLoggedIn => basicAuth != null;

  /**
   * Populate the model with inbox mails
   */
  List<Email> loadEmailList(int offset, int count) {
    List<Email> mails = [];
    /*
    mails.add(new Email("geomail@mail.bitlab.dk", "Hi and welcome to Geo-mail"));
    mails[0].setContent("Hi,\nTry out the geo-mailing lists in the bottom of the "+
    "page, it will allow you to send e-mails to everyone logged into GeoMail in your"+
    " local area.\n\nThanks\nThe GeoMail Team");
*/

    mails = connection.queryForMail(offset,count, session);

    return mails;
  }


  bool sendEmail(Email mail) {
    return connection.SendAnEmail(mail,this.session).OK;
  }

  /**
   * Get the current Geo Lists.
   */
  List<GeoList> getGeoListsForUser() {
    List<GeoList> result = connection.getGeoLists();
    if (response == null) {
      return [];
    }
    return result;
  }

  String getVersion() {
    QueryResponse resp = GetQuery("version.txt", "", null).Text;
    if (resp.OK == false) {
      return " no version ";
    }
    return resp.Text;
  }
}

