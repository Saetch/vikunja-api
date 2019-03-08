---
date: "2019-02-12:00:00+02:00"
title: "Install Frontend"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "setup"
---

# Frontend

Installing the frontend is just a matter of hosting a bunch of static files somewhere.

With nginx or apache, you have to [download](https://vikunja.io/en/download/) the frontend files first.
Unzip them and store them somewhere your server can access them.

You also need to configure a rewrite condition to internally redirect all requests to `index.html` which handles all urls. 

## Docker

The docker image is based on nginx and just contains all nessecary files for the frontend.

To run it, all you need is

{{< highlight bash >}}
docker run -p 80:80 vikunja/frontend
{{< /highlight >}}

which will run the docker image and expose port 80 on the host.

See [full docker example]({{< ref "full-docker-example.md">}}) for more varations of this config.

## NGINX

Below are two example configurations which you can put in your `nginx.conf`:

You may need to adjust `server_name` and `root` accordingly.

After configuring them, you need to reload nginx (`service nginx reload`).

### with gzip enabled (recommended)

{{< highlight conf >}}
gzip  on;
gzip_disable "msie6";

gzip_vary on;
gzip_proxied any;
gzip_comp_level 6;
gzip_buffers 16 8k;
gzip_http_version 1.1;
gzip_min_length 256;
gzip_types text/plain text/css application/json application/x-javascript text/xml application/xml application/xml+rss text/javascript application/vnd.ms-fontobject application/x-font-ttf font/opentype image/svg+xml;

server {
    listen       80;
    server_name  localhost;

    location / {
        root   /path/to/vikunja/static/frontend/files;
        try_files $uri $uri/ /;
        index  index.html index.htm;
    }
}
{{< /highlight >}}

### without gzip

{{< highlight conf >}}
server {
    listen       80;
    server_name  localhost;

    location / {
        root   /path/to/vikunja/static/frontend/files;
        try_files $uri $uri/ /;
        index  index.html index.htm;
    }
}
{{< /highlight >}}

## Apache

Apache needs to have `mod_rewrite` enabled for this to work properly:

{{< highlight bash >}}
a2enmod rewrite
service apache2 restart
{{< /highlight >}}

Put the following config in `cat /etc/apache2/sites-available/vikunja.conf`:

{{< highlight aconf >}}
<VirtualHost *:80>
    ServerName localhost
    DocumentRoot /path/to/vikunja/static/frontend/files
    RewriteEngine On
    RewriteRule ^\/?(config\.json|favicon\.ico|css|fonts|images|img|js) - [L]
    RewriteRule ^(.*)$ /index.html [QSA,L]
</VirtualHost>
{{< /highlight >}}

You probably want to adjust `ServerName` and `DocumentRoot`.

Once you've customized your config, you need to enable it:

{{< highlight bash >}}
a2ensite vikunja
service apache2 reload
{{< /highlight >}}

## Updating

To update, it should be enough to download the new files and overwrite the old ones.
The paths contain hashes, so all caches are invalidated automatically.