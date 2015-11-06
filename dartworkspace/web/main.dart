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
import 'geoconnection.dart';

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
  GeoMailModel model;
  AnchorElement selected;
  DivElement emailContent;

  MailWindowController(this.model) {
    view = querySelector("#mail-window");
    listOfEmails = querySelector("#mail-window-list-of-emails");
    emailContent = querySelector("#email-content");
    view.style.display = 'none';
  }

  void displayWindow() {
    view.style.display = 'block';
    listOfEmails.children.clear();
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
  }

  void hideWindow() {
    view.style.display = 'none';
  }
}

class SystemMessageController {
  Element msg;

  SystemMessageController() {
    this.msg = querySelector("#system-message");
  }

  void display(String message) {
    msg.innerHtml = message;
    msg.style.display = 'block';
    new Timer(new Duration(seconds: 2), () {
      this.hide();
    });
  }

  void hide() {
    msg.style.display = 'none';
  }
}

class LoginWindowController {
  DivElement view;
  ButtonElement signInbutton;
  InputElement username;
  PasswordInputElement password;
  GeoMailModel model;
  MailWindowController nextControl;

  LoginWindowController(this.model, this.nextControl) {
    view = querySelector("#login-window");
    signInbutton = querySelector("#login-window-sign-in-button");
    username = querySelector("#login-window-username");
    password = querySelector("#login-window-password");
    view.style.display = 'block';

    signInbutton.onClick.listen((e) {
      model.logIn(username.value, password.value).then((ok) {
        if (ok) {
          this.hideWindow();
          nextControl.displayWindow();
        } else {
        }
      });
    });
  }

  void displayWindow() {
    view.style.display = 'block';
  }

  void hideWindow() {
    view.style.display = 'none';
  }
}

class

SignOutController {
  ButtonElement view;
  MailWindowController mailView;
  GeoMailModel model;
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
  GeoMailModel model;
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
  GeoMailModel model;

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

  GeoMailModel model;

  ViewController(this.model) {
    this.mailWindow = new MailWindowController(model);
    this.loginWindow = new LoginWindowController(model, mailWindow);
    this.mboxController = new MailBoxSelectController();
    this.signOut = new SignOutController(model, mailWindow, loginWindow);
    this.completeErrorMessage = new NoServiceFullScreenErrorMessageController();
    this.systemMessages = new SystemMessageController();
    model.setView(this);
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
      mailWindow.displayWindow();
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

void displayVersionString() {
  querySelector("#version").innerHtml =
  "You are watching Geo Mail version <font color=\"red\">" +
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
  GeoMailModel model = new GeoMailModel(conn);
  ViewController view = new ViewController(model);
  view.display();
}
