/**
 *
 * Entry point of the at the browser in index.html pointing here.
 *
 * We implement the Model-View-Controller pattern. The view is written
 * in HTML found in index.html. The Model is implemented in mailmodel.dart.
 * This file contains the controllers. We have a controller class for each
 * business-logical component presented by the view. These are:
 *
 * - LoginWindow
 * - System Message
 * - ComposeEmail Window
 * - NoServiceFullScreenErrorMessage
 *
 * which are all collected in one ViewController that takes care of
 * Business logic using its sub-controllers.
 *
 * Here we use the Dart language to provide separation between view
 * and controller by letting HTML
 *
 * Author: Rasmus Winther Zakarias
 */

import 'dart:html';
import 'dart:async';
import 'mailmodel.dart';
import 'bitmailconnection.dart';

/**
 * Checks that {input} contains only characters from {validCodePoints}
 */
bool checkValidInput(String input, List<int> validCodePoints) {
  for (int i =0; i < input.length;++i) {
    int c = input.codeUnitAt(i);
    if (validCodePoints.contains(c) == false) {
      return false;
    }
  }
}

bool checkValidLogin(String username) {
  return checkValidInput(username,"abcdefghijklmnopqrstuwxyzABCDEFGHIJKLMNOPQRSTUWXYZ".codeUnits);
}

bool checkValidEmailTo(String mail) {
  if (mail.codeUnits.contains('@'.codeUnits[0]) == false) {
    print("No @ sign in e-mail address: ${mail}");
    return false;
  }

  var parts = mail.split("@");
  if (parts.length != 2) {
    print("Parts length not 2 ${parts.length}: ${mail}");
    return false;
  }

  return true;
}

bool checkToHeader(String to) {
  checkValidInput(to,"abcdefghijklmnopqrstuwxyzABCDEFGHIJKLMNOPQRSTUWXYZ@0123456789,.".codeUnits);

  List<String> a = to.split(",");
  a.forEach( (mailAddr) {
    if (checkValidEmail(mailAddr) ==false) {
      return false;
    }
  });

  return true;
}


class SystemMessageController {
  Element msg;
  Timer current;

  SystemMessageController() {
    this.msg = querySelector("#system-message");
  }

  void display(String message) {
    msg.innerHtml = message;
    msg.style.display = 'block';
    Timer current = this.current;
    if (current != null) {
      current.cancel();
    }
    current = new Timer(new Duration(seconds: 20), () {
      this.hide();
      this.current = null;
    });
  }

  void hide() {
    msg.style.display = 'none';
  }
}

class ComposeEmailWindowController {

  DivElement _view;
  InputElement _recipients;
  InputElement _subject;
  ButtonElement _send;
  ButtonElement _cancel;
  TextAreaElement _content;
  BitMailModel _model;
  ViewController _viewController;

  ComposeEmailWindowController(this._model, this._viewController) {
    _view = querySelector("#compose-email-window");
    _recipients = querySelector("#compose-email-window-recipients");
    _subject = querySelector("#compose-email-window-subject");
    _send = querySelector("#compose-email-window-send");
    _cancel = querySelector("#compose-email-window-cancel");
    _content = querySelector("#compose-email-window-content");

    _cancel.onClick.listen((e) => this._handleCancelClick());
    _send.onClick.listen((e) => this._handleSendClick());
  }

  void _reset() {
    _recipients.value = "";
    _subject.value = "";
    _content.value = "";
  }

  void _handleCancelClick() {
    this.hide();
    this._reset();
    _viewController.loginWindow.displayWindow();
    _model.logOut();
  }

  void _handleSendClick() {
    Email mail = new Email.WithModel(_model, this.Recipients, this.Subject);
    mail.setContent(this._content.value);

    if (checkValidEmailTo(this.Recipients) == false) {
      _viewController.setSystemMessage("Invalid To. Only [a-zA-Z0-9] are allowed in before and after the @<br/>" +
      "Separate by comma.");
      return;
    }

    if (_model.sendEmail(mail) == true) {
      _viewController.setSystemMessage("Sending email ok");
    } else {
      _viewController.setSystemMessage("Failed to send message");
    }
    this.hide();
    this._reset();
    _viewController.display();
  }

  void hide() {
    _view.style.display = 'none';
  }

  void display() {
    _view.style.display = 'block';
  }

  get Content => this._content.value;

  get Recipients => this._recipients.value;

  get Subject => this._subject.value;

}


class LoginWindowController {
  DivElement view;
  ButtonElement signInbutton;
  InputElement username;
  PasswordInputElement password;
  BitMailModel model;
  ViewController viewController;

  LoginWindowController(this.model, this.viewController) {
    view = querySelector("#login-window");
    signInbutton = querySelector("#login-window-sign-in-button");
    username = querySelector("#login-window-username");
    password = querySelector("#login-window-password");
    view.style.display = 'block';

    signInbutton.onClick.listen((e) {

      if (checkValidLogin(username.value) == false) {
        viewController.setSystemMessage("Invalid username can only contain upper and lower case letters [a-zA-Z].");
        return;
      }

      var ok = model.logIn(username.value, password.value);
      if (ok) {
        this.hideWindow();
        viewController.composeEmail();
      } else {
        viewController.setSystemMessage("Login failed");
      }
    });
  }

  void displayWindow() {
    view.style.display = 'block';
  }

  void hideWindow() {
    view.style.display = 'none';
  }
}


class NoServiceFullScreenErrorMessageController {
  DivElement view;
  HeadingElement msg;

  NoServiceFullScreenErrorMessageController() {
    view = querySelector("#complete-error-message");
    msg = querySelector("#error-message");
  }

  void display(String message) {
    msg.innerHtml = message;
  }

  void hide() {
    view.style.display = 'none';
  }
}

class GeoMailingListItem {
  String name;
  AnchorElement item;
  BitMailModel model;
  int count;

  GeoMailingListItem(this.name, this.model) {
    item = new AnchorElement();
    item.onClick.listen(() => this.clicked());
  }

  void buildItem() {
    item.style.display = 'none';
    item.children.clear();
    item.innerHtml = "<b>${this.name}</b>(${this.count})";
  }

  void display() {
    item.style.display = 'block';
  }

  void hide() {
    item.style.display = 'none';
  }

  void clicked() {
  }
}

class ViewController {
  LoginWindowController loginWindow;
  NoServiceFullScreenErrorMessageController completeErrorMessage;
  SystemMessageController systemMessages;
  ComposeEmailWindowController composerWindow;

  BitMailModel model;

  ViewController(this.model) {
    this.loginWindow = new LoginWindowController(model, this);
    this.completeErrorMessage = new NoServiceFullScreenErrorMessageController();
    this.systemMessages = new SystemMessageController();
    this.composerWindow = new ComposeEmailWindowController(model, this);
    model.setView(this);
  }

  void composeEmail() {
    composerWindow.display();
  }

  void connectionDown() {
    this.hide();
    completeErrorMessage.display("We have lost connection with the server.");
  }

  void connectionUp() {
    this.display();
    completeErrorMessage.hide();
  }

  void display() {
    if (model.IsLoggedIn) {
      composeEmail();
    } else {
      loginWindow.displayWindow();
    }
  }

  void hide() {
    loginWindow.hideWindow();
    completeErrorMessage.hide();
  }

  void setSystemMessage(String message) {
    this.systemMessages.display(message);
  }
}

void displayVersionString(BitMailModel model) {
  querySelector("#version").innerHtml =
  "You are watching Bit Mail version <font color=\"red\">" +
  model.getVersion() +
  "</font>";
}

/**
 *
 * Main entry point. We initialize the connection, start a GeoMailModel
 * and take control of the Html View.
 *
 */
main() {
  ClientAPI conn = new ClientAPI("/go.api");
  BitMailModel model = new BitMailModel(conn);
  ViewController view = new ViewController(model);
  conn.SetPinger( (msg) => view.setSystemMessage(msg));
  displayVersionString(model);
  view.display();
}
