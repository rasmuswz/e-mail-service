import 'dart:async';
import 'dart:html';
import 'dart:convert';
import 'mailmodel.dart';

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
      print("setting headers["+key+"]="+value);
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
    print("ResponseText:"+req.responseText);
    return new QueryResponse.Ok(req.responseText);
  } else {
    // not fine, write error to browser JS-console.
    print("StatusText:"+req.statusText);
    return new QueryResponse.Fail(req.statusText);
  }


}

QueryResponse GetQuery(String path, String data, Map<String, String> headers) {

  // open Ajax
  HttpRequest req = new HttpRequest();
  req.open("GET", path, async: false);

  // Set custom headers
  if (headers != null) {
    headers.forEach((key, value) {
      print("setting headers["+key+"]="+value);
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
    print("ResponseText:"+req.responseText);
    return new QueryResponse.Ok(req.responseText);
  } else {
    // not fine, write error to browser JS-console.
    print("StatusText:"+req.statusText);
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
    QueryResponse response = GetQuery("/geolist", "", {"SessionId": sessionId});
    if (response.OK == false) {
      return null;
    }

    List<Map<String, String>> decoded = JSON.decode(response.Text);

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
    QueryResponse response = GetQuery(_path + "/login"+q, "", {"Authorization": "Basic " + basicAuth});
    if (response.OK) {
      return response.Text;
    }
    return null;
  }

  /**
   * Logout un-registering session id
   */
  void doLogout(String sessionId) {
    QueryResponse response = GetQuery(_path + "/logout?session=" + sessionId, "", null);
    print("logout response: " + response.Text);
  }

  //
  // Check that the connection to the server is alive.
  //
  void _check(Timer t) {
    QueryResponse  resp = GetQuery(_path + "/alive", "", null);
    _alive = resp.OK;
    if (_alive != _previousAlive) {
      print("Notifying " + stateListeners.length.toString() + " listeners");
      stateListeners.forEach((ConnectionListener f) {
        f(_alive);
      });
      _previousAlive = _alive;
    }
  }


  List<Email> queryForMail(int offset, int count, String sessionId ) {

    String jsRequest = "{index: ${offset}, length: ${count}}";

    QueryResponse response = GetQuery(_path+"/getmail",jsRequest,{"SessionId": sessionId});

    if (response.OK) {
    print("Response to getmail \""+response.Text+"\"");
      List<Map<String,String>> mails = JSON.decode(response.Text);
      return [];
    } else {
      return [];
    }

  }

  /**
   * Send an e-mail for delivery
   */
  QueryResponse SendAnEmail(Email email, String sessionId) {
    String jsonString = email.toJson();
    print("Sending data: "+jsonString);
    QueryResponse response = PostQuery(_path + "/sendmail", jsonString,
          {"sessionID": sessionId}  );
    return response;
  }

  get Alive => _alive;

}
