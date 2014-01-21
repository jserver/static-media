static-media
============

A Simple static and media server.

I use this to serve Django's static and media directories.
When given the backing flag, media will be copied down from the backing url.
Use the port flag to change from the default 8001.

example
=======
go-assets --port=9001 --backing=https://some-prod-site.cloudfront.net/ /home/user/assets
