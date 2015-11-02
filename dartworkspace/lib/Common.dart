/**
 * Common things that are used as helpers
 * everywhere is goes here.
 *
 * Author: Rasmus Winther Zakarias
 *
 */

/**
 * An error carries a message for the user.
 */
class MyError {
  String _userMessage;
  MyError(this._userMessage);
  get UserMessage => _userMessage;
}
