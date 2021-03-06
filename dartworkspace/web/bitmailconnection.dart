/**
 *
 * Browser side of the ClientAPI, handling serialization, string and bytes, while
 * the model in mailmodel.dart works with objects.
 *
 * Author: Rasmus Winther Zakarias
 *
 */

import 'dart:async';
import 'dart:html';
import 'mailmodel.dart';

typedef void PingMessageDisplayer(String pingMessage);

/**
 *
 * Sending helpers
 *
 */
class QueryResponse {
  String Text;
  bool OK;

  QueryResponse.Ok(this.Text) {
    OK = true;
  }

  QueryResponse.Fail(this.Text) {
    OK = false;
  }
}

QueryResponse PostQuery(String path, String data, Map<String, String> headers) {
  // open Ajax
  HttpRequest req = new HttpRequest();
  req.open("POST", path, async: false);

  // Set custom headers
  if (headers != null) {
    headers.forEach((key, value) {
      print("setting headers[" + key + "]=" + value);
      req.setRequestHeader(key, value);
    });
  }

  // send request synchronously
  try {
    print("Sending data " + data);
    req.send(data);
  } catch (exception) {
    return new QueryResponse.Fail("Exception: ${exception.toString()}");
  }

  // check if everything is fine.
  if (req.status == 200) {
    print("ResponseText:" + req.responseText);
    return new QueryResponse.Ok(req.responseText);
  } else {
    // not fine, write error to browser JS-console.
    print("StatusText:" + req.statusText);
    return new QueryResponse.Fail(req.statusText);
  }


}

QueryResponse _GetQuery(String path, String data, Map<String, String> headers) {

  // open Ajax
  HttpRequest req = new HttpRequest();
  req.open("GET", path, async: false);

  // Set custom headers
  if (headers != null) {
    headers.forEach((key, value) {
      print("setting headers[" + key + "]=" + value);
      req.setRequestHeader(key, value);
    });
  }

  // send request synchronously
  try {
    print("Sending data" + data);
    req.send(data);
  } catch (exception) {
    return new QueryResponse.Fail("Exception: ${exception.toString()}");
  }

  // check if everything is fine.
  if (req.status == 200) {
    print("ResponseText:" + req.responseText);
    return new QueryResponse.Ok(req.responseText);
  } else {
    // not fine, write error to browser JS-console.
    print("StatusText:" + req.statusText + " " + req.responseText);
    return new QueryResponse.Fail(req.statusText);
  }

}

typedef ConnectionListener(bool s);

/**
 * Represent an open connection to the Server implementing the Client API component.
 */
class ClientAPI {

  String _path;
  bool _alive;
  bool _previousAlive;
  List<ConnectionListener> stateListeners;
  String authorization;
  PingMessageDisplayer pingMsgDisp;


  ClientAPI(String path) {
    this._path = path;
    _previousAlive = false;
    new Timer.periodic(new Duration(seconds: 5), _check);
    this.stateListeners = new List<ConnectionListener>();
  }

  void SetPinger(PingMessageDisplayer arg) {
    this.pingMsgDisp = arg;
  }

  void message(String m){
    if (this.pingMsgDisp != null) {
      pingMsgDisp(m);
    }
  }

  void SetAuthorization(String basicAuth) {
    this.authorization = basicAuth;
  }

  /**
   * Register {connectionListener} to be invoked upon
   * connection state change.
   */
  void listenForState(ConnectionListener connectionListener) {
    stateListeners.add(connectionListener);
  }

  /**
   * Check whether given credentials will work
   */
  String doLogin(String basicAuth) {
    QueryResponse response = _GetQuery(_path + "/login", "", {"Authorization": "Basic " + basicAuth});
    if (response.OK) {
      return response.Text;
    } else {
      message(response.Text);
    }
    return null;
  }

  /**
   * Logout un-registering session id
   */
  void doLogout(String sessionId) {
    QueryResponse response = _GetQuery(_path + "/logout?SessionId=" + sessionId, "", null);
    print("logout response: " + response.Text);
  }

  //
  // Check that the connection to the server is alive.
  //
  void _check(Timer t) {
    QueryResponse resp = _GetQuery(_path + "/alive?state", "", null);
    _alive = resp.OK;
    if (_alive != _previousAlive) {
      print("Notifying " + stateListeners.length.toString() + " listeners");
      stateListeners.forEach((ConnectionListener f) {
        f(_alive);
      });
      _previousAlive = _alive;
    }
    if (resp.Text.startsWith("!")) {
      if (this.pingMsgDisp != null) {
        this.pingMsgDisp(resp.Text.substring(1));
      }
    }
  }


  String getVersion() {
    QueryResponse resp = _GetQuery("version.txt", "", null);
    if (resp.OK) {
      return resp.Text;
    } else {
      message(resp.Text);
      return "no version";
    }
  }


  /**
   * Send an e-mail for delivery
   */
  QueryResponse SendAnEmail(Email email, String sessionId) {
    String jsonString = email.toWireMail();
    print("Sending data: " + jsonString);
    QueryResponse response = PostQuery(_path + "/sendmail", jsonString,
    {"SessionId": sessionId});

    if  (response.OK == false ) {
      message(response.Text);
    }
    return response;
  }


  get Alive => _alive;

}
