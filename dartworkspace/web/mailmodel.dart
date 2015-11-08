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
import 'bitmailconnection.dart';
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
  BitMailModel model;

  Email.fromMap(Map<String, String> json) {
    this._from = json['from'];
    this._to = json['to'];
    this._subject = json['subject'];
    this._location = json['location'];
  }

  Map<String, String> toMap() {
    Map<String, String> json = new Map<String, String>();
    String h = JSON.encode({"From": this._from, "To": this._to, "Subject": this._subject, "Location": this._location});
    print("headers: " + h);
    json["Headers"] = h;
    json["Content"] = this._content;
    return json;
  }

  String toJson() {
    return """{"Headers":{"From":"${this._from}","Subject":"${this._subject}","To":"${this._to}"},"Content":"${this._content}"}""";
    //return "{\"Headers\": {\"To\":\""+_to+"\",\"From\":\""+_from+"\":"
  }

  Email(this._from, this._subject) {

  }

  Email.WithModel(this.model, this._to, this._subject) {
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
 * The Geo Mail Model implementing the core-business logic
 */
class BitMailModel {

  String basicAuth;
  ClientAPI connection;
  List<Future<String>> stateListeners;
  String session;
  ViewController view;
  String _username;


  /**
   * With a connection to the ClientAPI-server we
   * create BitMailModel.
   */
  BitMailModel(ClientAPI connection) {
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
  bool logIn(String username, String password) {
    this.basicAuth = null;
    String auth = window.btoa(username + ":" + password);

    view.setSystemMessage("We are processing your location data and verifys your identity...");

    String sessionId = connection.doLogin(auth);
    if (sessionId != null) {
      this.basicAuth = auth;
      this.session = sessionId;
      this._username = username;
      return true;
    } else {
      return false;
    }
  }


  /** Are we logged in or not */
  get IsLoggedIn => basicAuth != null;

  /**
   * Populate the model with inbox mails
   */
  List<Email> loadEmailList(int offset, int count) {
    List<Email> mails = [];

    mails = connection.queryForMail(offset, count, session);

    return mails;
  }


  bool sendEmail(Email mail) {
    return connection.SendAnEmail(mail, this.session).OK;
  }

  String getVersion() {
    return connection.getVersion();
  }
}

