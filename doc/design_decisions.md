
# Decisions Log


## 2020-11-19
Who: waschtl & anni

 - We subscribe to the BBB / Akka internal communication
   to extract BBB events.
 - The we update the state based on this events.
   (MeetingEnded, MeetingDestroyed, UserJoined, UserLeft)

Alternatives:

 - Trying to poll the meetings. But it's impractical because
   of BBB api performance
