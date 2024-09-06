#!/usr/bin/ruby
# encoding: UTF-8

#
# Import the recording after publishing into b3scale.
# Place this file in /usr/local/bigbluebutton/core/scripts/post_publish
# 
# Good luck.
#

require "optimist"

require File.expand_path('../../../lib/recordandplayback', __FILE__)

opts = Optimist::options do
  opt :meeting_id, "recording id to import", :type => String
  opt :format, "playback format name", :type => String
end
meeting_id = opts[:meeting_id]

logger = Logger.new("/var/log/bigbluebutton/post_publish.log", 'weekly' )
logger.level = Logger::INFO
BigBlueButton.logger = logger

metadata_file = "/var/bigbluebutton/published/presentation/#{meeting_id}/metadata.xml"

access_token = "<insert access token here> or <read it from somewhere safe>"

b3scale_api = "https://bbb.example.com/api/v1"
b3scale_recordings_import = "#{b3scale_api}/recordings-import"

# Upload metadata.xml using curl because it is easier to use
# than Net::HTTP. *shrug*
system(
  "curl",
  "-X", "POST",
  "-H", "Authorization: Bearer #{access_token}",
  "-H", "Content-Type: application/xml",
  "-d", "@#{metadata_file}",
  b3scale_recordings_import)

exit $?.exitstatus
