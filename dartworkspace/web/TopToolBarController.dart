/**
 *
 * Controller for the tool-bar in the top of the page.
 *
 */

import 'dart:html';

/**
 * +=============================+
 * | MailBoxes        Compose    |
 * +=============================+
 *
 * Controls the "top-tool-bar" div on the page.
 */
class TopToolBarController {
  DivElement view = $['top-tool-bar'];
  MailModel model;
  TopToolBarController(this.model);
}
