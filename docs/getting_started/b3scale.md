# Setting up the main service

## Installing b3scale

=== "Debian/Ubuntu"

    Starting with version 1.0.3, b3scale provides `.deb` packages for use in Ubuntu. Download the `b3scaled-*.deb` asset from the [release page](https://github.com/b3scale/b3scale/releases) on GitHub. On an Intel/AMD 64 Bit system, install the `.deb` package like this:

    ```bash
    sudo dpkg -i b3scale_1.0.3_linux_amd64.deb
    ```

=== "openSUSE"

    For openSUSE Tumbleweed run the following as root:

    ```bash
    zypper addrepo https://download.opensuse.org/repositories/home:dmolkentin:infrarun/openSUSE_Tumbleweed/home:dmolkentin:infrarun.repo
    zypper refresh
    zypper install b3scale
    ```

    For openSUSE Leap 15.5 run the following as root:
    ```bash
    zypper addrepo https://download.opensuse.org/repositories/home:dmolkentin:infrarun/15.5/home:dmolkentin:infrarun.repo
    zypper refresh
    zypper install b3scale
    ```

=== "Other distributions"

    Pre-compiled binaries are available from the [release page](https://github.com/b3scale/b3scale/releases) on GitHub. While you can use those binaries, you will need to setup systemd services by yourself. You can find the skeleton files [here](https://github.com/b3scale/b3scale/tree/main/doc).

=== "Kubernetes/Docker"

    Starting with version 1.0.3, we provide an experimental, scratch-based container image, available from the GitHub Container Registry:

    ```bash
    docker pull ghcr.io/b3scale/b3scaled:1.0.3
    ```

## Configuring the service

`b3scale` is configured exclusively through environment variables. This makes it work both in a traditional and cloud-native environment.
If you have installed `b3scale` using packages, you can find the settings in `/etc/default/b3scale` (Debian/Ubuntu) or `/etc/sysconfig/b3scale`
(OpenSUSE).

Define at least `B3SCALE_DB_URL` and point it to a previously created database on your Postgres server:

```bash
B3SCALE_DB_URL="postgres://b3scale:secretpassword@localhost:5432/b3scale"
```

The next important variable is `B3SCALE_API_JWT_SECRET`. It is required as seed for generating JWTs, which control the access of API clients.

```bash
B3SCALE_API_JWT_SECRET="a_string_of_random_caracters"
```

you can create a safe secret using the `pwgen` tool:

```bash
pwgen -snc 42 1
```

Leave the remaining options as they are. We will cover them in a later chapter. A simple Postgres database can be created like so:

```bash
sudo -u postgres psql
postgres=# create database b3scale;
postgres=# create user myuser with encrypted password 'secretpassword';
postgres=# grant all privileges on database b3scale to b3scale;
```

Finally, start the service:

```bash
systemctl start b3scale
```

## TLS termination

By default, b3scale binds to port `42352` on `localhost`. Use a reverse-proxy capable web server for TLS termination.
### nginx

To use nginx as a frontend proxy and TLS termination, adjust and use the following snippet:

```nginx
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name api.bbb.example.org;

    # From the conservative settings on https://ssl-config.mozilla.org/
    ssl_protocols               TLSv1.2 TLSv1.3;
    ssl_ciphers                 ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers   off;
    ssl_certificate             /etc/nginx/certs.d/api.bbb.example.org.pem;
    ssl_certificate_key         /etc/nginx/certs.d/api.bbb.example.org.key;
    ssl_session_timeout         1d;
    ssl_session_cache           shared:MozSSL:10m;
    ssl_session_tickets         off;
    ssl_stapling                on;
    ssl_stapling_verify         on;

    # access_log                /var/log/nginx/access.log;
    error_log                   /var/log/nginx/error.log;

    add_header                  Strict-Transport-Security "max-age=15768000;" always;
    add_header                  X-Robots-Tag "none" always;

    location / {
        proxy_pass          http://127.0.0.1:42352;
        proxy_http_version  1.1;
        proxy_set_header    X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header    X-Forwarded-Proto $scheme;
    }

    location /metrics {
        proxy_pass          http://127.0.0.1:42352;
        allow               <ipv4_of_prometheus_host>/32;
        allow               <ipv6_of_prometheus_host>/128;
        deny                all;
        proxy_http_version  1.1;
        proxy_set_header    X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header    X-Forwarded-Proto $scheme;
    }

    location = / {
        return              403;
    }
}
```

### Other load balancers

Other load balancers and TLS terminators should just work just the same way. If you would like to contribute documentation, please open an issue.

## Scale and redundancy of b3scale

While a single instance of `b3scaled` has worked fine during the Covid 19 pandemic, serving some 100k concurrent users while maintaining more than a hundred backends, and multiple frontends, keeping b3scale redundant seems like a good idea. To do so, it is possible to launch several instances of `b3scaled`, as its operations are designed to be atomic towards the database.

!!! note
    Your database and your TLS terminator are also single points of failure. Make sure to make them redundant as well if you aim for systematic redundancy, and make sure they don't become your weakest link if you scale up `b3scaled`.

## Bootstrapping and migrations

For the next step, you will need the `b3scalectl` client tool. Download it from the GitHub release page.

```bash
b3scalectl --api https://api.bbb.example.org
```

!!! note
    Set up a convenience alias like this:
    ```bash
    alias b3scalectl_example='b3scalectl --api https://api.bbb.example.org'
    ```

`b3scalectl` will now ask for the shared API secret that you provided to `b3scaled` as `B3SCALE_API_JWT_SECRET`. It uses this secret to derive
a JWT token. The secret itself will not be stored. Next, run `b3scalectl --api https://api.bbb.example.org db migrate` to apply the initial
database structure. b3scale is now operational.
