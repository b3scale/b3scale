#!/bin/sh

getent group bigbluebutton > /dev/null || groupadd -r bigbluebutton
getent passwd bigbluebutton > /dev/null ||  useradd -d /var/bigbluebutton -g bigbluebutton -r bigbluebutton -M
mkdir -p /var/bigbluebutton
chown bigbluebutton:bigbluebutton /var/bigbluebutton
