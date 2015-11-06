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

  Email.fromMap(Map<String, String> json) {
    this._from = json['from'];
    this._to = json['to'];
    this._subject = json['subject'];
    this._location = json['location'];
  }

  Map<String, String> toMap() {
    Map<String, String> json = new Map<String, String>();
    json['from'] = this._from;
    json['to'] = this._to;
    json['subject'] = this._subject;
    json['location'] = this._location;
    return json;
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

  /**
   * With a connection to the ClientAPI-server we
   * create a GeoMailModel.
   */
  GeoMailModel(ClientAPI connection) {
    this.connection = connection;
    this.stateListeners = [];
  }

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
  bool logIn(String username, String password) {
    this.basicAuth = null;
    String auth = window.btoa(username + ":" + password);

    Geolocation location = Geolocation.getCurrentPosition(enableHighAccuracy: false,
    timeout: new Duration(minutes: 2));
    locationString = location.toString();
    String sessionId = connection.doLogin(auth);
    if (sessionId != null) {
      this.basicAuth = auth;
      this.session = sessionId;
    }
    return this.basicAuth != null;
  }


  /** Are we logged in or not */
  get IsLoggedIn => basicAuth != null;

  /**
   * Populate the model with inbox mails
   */
  List<Email> loadEmailList(int offset, int count) {
    List<Email> mails = [];
    mails.add(new Email("geomail@mail.bitlab.dk", "Hi and welcome to Geo-mail"));
    mails[0].setContent("Hi,\nTry out the geo-mailing lists in the bottom of the "+
    "page, it will allow you to send e-mails to everyone logged into GeoMail in your"+
    " local area.\n\nThanks\nThe GeoMail Team");


    return mails;
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
    String verstr = GetQuery("version.txt", "", null);
    if (verstr == null) {
      return " no version ";
    }
    return verstr;
  }
}

