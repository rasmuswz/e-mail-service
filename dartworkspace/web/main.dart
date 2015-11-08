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
 * - MailWindow
 *    [] list of email on the left
 *    [] main mail-reading window
 * - System Message
 * - SignOut
 * - NoServiceFullScreenErrorMessage
 * - LocationMailingList
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
 * Controls the MailBox dropdown menu that selects which
 * folder is presently selected.
 */
class MailBoxSelectController {
  UListElement view;

  MailBoxSelectController() {
    view = querySelector("#mail-box-selection");
  }

  void setOptions(List<String> options) {
    this.view.children.clear();

    options.forEach((o) {
      LIElement item = new LIElement();
      AnchorElement opt = new AnchorElement();
      opt.href = "#";
      opt.text = o;
      item.children.add(opt);
      this.view.children.add(item);
    });
    return;
  }
}

class EmailViewItem {
  Email mail;

  EmailViewItem(this.mail);

  AnchorElement display() {
    AnchorElement item = new AnchorElement();
    item.className = "list-group-item";

    // set email from
    HeadingElement from = new HeadingElement.h4();
    from.text = mail.From;
    from.className = "list-group-item-heading";
    item.children.add(from);

    // set subject
    ParagraphElement subject = new ParagraphElement();
    subject.className = "list-group-item-text";
    subject.text = mail.Subject;
    item.children.add(subject);

    return item;
  }
}

class MailWindowController {
  DivElement view;
  DivElement listOfEmails;
  BitMailModel model;
  AnchorElement selected;
  DivElement emailContent;
  ButtonElement compose;
  ViewController _viewController;

  MailWindowController(this.model, this._viewController) {
    view = querySelector("#mail-window");
    listOfEmails = querySelector("#mail-window-list-of-emails");
    emailContent = querySelector("#email-content");
    view.style.display = 'none';
    compose = querySelector("#mail-window-compose");
    compose.onClick.listen((e) => _viewController.composeEmail());
  }

  void displayWindow() {
    view.style.display = 'block';
    listOfEmails.children.clear();

    new Timer(new Duration(seconds: 2), () {
      List<Email> emails = model.loadEmailList(0, 10);
      emails.forEach((mail) {
        AnchorElement m = new EmailViewItem(mail).display();
        listOfEmails.children.add(m);
        m.onClick.listen((e) {
          if (selected != null) {
            selected.className = "list-group-item";
          }
          m.className = "list-group-item active";
          selected = m;
          emailContent.innerHtml = "${mail.Content}";
        });
      });
    });

  }

  void hideWindow() {
    view.style.display = 'none';
  }
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
    _viewController.signOut.signOut();
  }

  void _handleSendClick() {
    Email mail = new Email.WithModel(_model, this.Recipients, this.Subject);
    mail.setContent(this._content.value);
    if (_model.sendEmail(mail) == true) {
      _viewController.setSystemMessage("Sending email ok");
    } else {
      _viewController.setSystemMessage("Failed to sent message");
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

class SignOutController {
  ButtonElement view;
  MailWindowController mailView;
  BitMailModel model;
  LoginWindowController loginView;

  SignOutController(this.model, this.mailView, this.loginView) {
    view = querySelector("#logout");
    view.onClick.listen((e) {
      this.signOut();
    });
  }

  void signOut() {
    mailView.hideWindow();
    loginView.displayWindow();
    model.logOut();
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

class LocationMalingListController {
  DivElement mailingList;
  BitMailModel model;

  LocationMalingListController(this.model) {
    this.mailingList = querySelector("#geo-mailing-listts");
  }

  void populateList() {
  }
}

class ViewController {
  MailWindowController mailWindow;
  LoginWindowController loginWindow;
  MailBoxSelectController mboxController;
  SignOutController signOut;
  NoServiceFullScreenErrorMessageController completeErrorMessage;
  SystemMessageController systemMessages;
  ComposeEmailWindowController composerWindow;

  BitMailModel model;

  ViewController(this.model) {
    this.mailWindow = new MailWindowController(model, this);
    this.loginWindow = new LoginWindowController(model, this);
    this.mboxController = new MailBoxSelectController();
    this.signOut = new SignOutController(model, mailWindow, loginWindow);
    this.completeErrorMessage = new NoServiceFullScreenErrorMessageController();
    this.systemMessages = new SystemMessageController();
    this.composerWindow = new ComposeEmailWindowController(model, this);
    model.setView(this);
  }

  void browseEmails() {
    mailWindow.displayWindow();
    composerWindow.hide();
  }

  void composeEmail() {
    mailWindow.hideWindow();
    composerWindow.display();
  }

  void setMailBoxes(List<String> mboxNames) {
    mboxController.setOptions(mboxNames);
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
    mailWindow.hideWindow();
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
  displayVersionString(model);
  view.display();
}
