import 'dart:async';
import 'dart:html';
import 'dart:convert';
import 'mailmodel.dart';


String GetQuery(String path, String data, Map<String, String> headers) {

  // open Ajax
  HttpRequest req = new HttpRequest();
  req.open("GET", path, async: false);

  // Set custom headers
  if (headers != null) {
    headers.forEach((key, value) {
      req.setRequestHeader(key, value);
    });
  }

  // send request synchronously
  try {
    req.send(data);
  } catch (exception) {
    return null;
  }

  // check if everything is fine.
  if (req.status == 200) {
    return req.responseText;
  } else {
    // not fine, write error to browser JS-console.
    print(req.statusText);
    return null;
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


  ClientAPI(String path) {
    this._path = path;
    _previousAlive = false;
    new Timer.periodic(new Duration(seconds: 2), _check);
    this.stateListeners = new List<ConnectionListener>();
  }

  void SetAuthorization(String basicAuth) {
    this.authorization = basicAuth;
  }

  List<GeoList> getGeoLists(String sessionId) {
    List<GeoList> result = [];
    String response = GetQuery("/geolist", "", {"SessionId": sessionId});
    if (response == null) {
      return null;
    }

    List<Map<String, String>> decoded = JSON.decode(response);

    decoded.forEach((geo) {
      result.add(GeoItem.NewFromMap(geo));
    });

    return result;

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
  String doLogin(String basicAuth, String location) {
    String q = "?location="+location;
    String response = GetQuery(_path + "/login"+q, "", {"Authorization": "Basic " + basicAuth});
    return response;
  }

  void doLogout(String sessionId) {
    String response = GetQuery(_path + "/logout?session=" + sessionId, "", null);
    print("logout response: " + response);
  }

  //
  // Check that the connection to the server is alive.
  //
  void _check(Timer t) {
    String resp = GetQuery(_path + "/alive", "", null);
    _alive = resp != null;
    if (_alive != _previousAlive) {
      print("Notifying " + stateListeners.length.toString() + " listeners");
      stateListeners.forEach((ConnectionListener f) {
        f(_alive);
      });
      _previousAlive = _alive;
    }
  }


  /**
   * Send an e-mail for delivery
   */
  bool SendAnEmail(Email email, String sessionId) {
    Sring jsonString = JSON.encode(email.toMap());
    String response = GetQuery(_path + "/sendmail", jsonString, {"sessionID": sessionId});
    return response != null;
  }

  get Alive => _alive;

}
