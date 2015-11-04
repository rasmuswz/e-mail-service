/**
 *
 * Entry point of the at the browser in index.html pointing here.
 *
 * We setup the model view controller here.
 *
 * Author: Rasmus Winther Zakarias
 */

import 'dart:html';
import 'dart:async';
import 'mailmodel.dart';

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


/**
 * Controls the list of mail-meta-data items in the list of mail.
 */
class MailListController {

  DivElement view;

  MailBoxMailListController() {
    view = querySelector("#list-of-emails");
  }

  void addMailitem(Email mail) {

  }


}


class EmailViewItem {

  Email mail;

  EmailViewItem(this.mail);


  AnchorElement render() {
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
  GeoMailDataModel model;
  AnchorElement selected;
  DivElement  emailContent;


  MailWindowController(this.model) {
    view = querySelector("#mail-window");
    listOfEmails = querySelector("#mail-window-list-of-emails");
    emailContent = querySelector("#email-content");
    view.style.display = 'none';
  }

  void displayWindow() {
    view.style.display = 'block';
    listOfEmails.children.clear();
    List<Email> emails = model.loadEmailList(0,10);
    emails.forEach( (mail) {
      AnchorElement m = new EmailViewItem(mail).render();
      listOfEmails.children.add(m);
      m.onClick.listen( (e) {
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


class LoginWindowController {

  DivElement view;
  ButtonElement signInbutton;
  InputElement username;
  PasswordInputElement password;
  GeoMailDataModel model;
  MailWindowController nextControl;

  LoginWindowController(this.model, this.nextControl) {
    view = querySelector("#login-window");
    signInbutton = querySelector("#login-window-sign-in-button");
    username = querySelector("#login-window-username");
    password = querySelector("#login-window-password");
    view.style.display = 'block';

    signInbutton.onClick.listen((e) {
      if (model.login(username.value, password.value)) {
        this.hideWindow();
        nextControl.displayWindow();
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
  GeoMailDataModel model;
  LoginWindowController loginView;

  SignOutController(this.model,this.mailView,this.loginView) {
    view = querySelector("#logout");
    view.onClick.listen( (e) {
      this.signOut();
    });
  }

  void signOut() {
    mailView.hideWindow();
    loginView.displayWindow();
    model.logout();

  }

}

main() {

  GeoMailConnection conn = new GeoMailConnection("/go.api");

  GeoMailDataModel model = new GeoMailDataModel(conn);

  MailWindowController mailWindow = new MailWindowController(model);
  LoginWindowController loginWindow = new LoginWindowController(model, mailWindow);

  MailBoxSelectController mboxController = new MailBoxSelectController();
  mboxController.setOptions(["inbox", "sent", "drafts", "play"]);

  SignOutController signOut = new SignOutController(model,mailWindow,loginWindow);

  model.ListenForConnectionState().then( (status) {
    if (status == "down") {
      signOut.signOut();
    }
  });


  querySelector("#version").innerHtml =
    "You are watching Geo Mail version <font color=\"red\">"+model.getVersion()+"</font>";



}