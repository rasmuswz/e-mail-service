/**
 *
 * Entry point of the at the browser in index.html pointing here.
 *
 * We setup the model view controller here.
 *
 * Author: Rasmus Winther Zakarias
 */

import 'package:bitlab_email_client/Common.dart';
import 'package:bitlab_email_client/MailModel.dart';




main() {
  List<String> mailBoxes = [];
  MailModel model = new MailModel(null );
  MyError e = model.ListMailBoxes(mailBoxes);
  if (e == null) {
    print("No MailBoxes available");
  }
}