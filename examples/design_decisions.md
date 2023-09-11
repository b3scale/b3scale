
# Decisions Log

## 2021-01-18

Who: annika

 - Reverse proxy functionality should be handled by a
   mature reverse proxy e.g. nginx.

 - The relevant routing information (host) is passed to
   the nginx front as a queryParameter during join.
   To prevent an open proxy, this information should
   not just be used as the proxy target.

 - There is currently no reason for handling these
   requests e.g. for modifying the content.

Alternatives:

 - Implement full reverse proxy with websocket support.


## 2020-11-19
Who: waschtl & anni

 - We subscribe to the BBB / Akka internal communication
   to extract BBB events.
 - The we update the state based on this events.
   (MeetingEnded, MeetingDestroyed, UserJoined, UserLeft)

Alternatives:

 - Trying to poll the meetings. But it's impractical because
   of BBB api performance


