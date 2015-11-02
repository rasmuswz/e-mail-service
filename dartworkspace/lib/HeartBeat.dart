/**
 *
 * In this file we implement a HeartBeat protocol
 * to check APIs are online.
 *
 * Author: Rasmus Winther Zakarias
 *
 */


/**
 *  A class implementing HeartBeat has a way of probing
 *  its services for health information.
 */
abstract class HeartBeat {
  String url;

  HeartBeat(this.url);

  bool isOnline();
}

