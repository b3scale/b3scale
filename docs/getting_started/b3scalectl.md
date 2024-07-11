# Setting up the command line client

The `b3scalectl`command line client is the most convenient way to get started with b3scale. It can handle all day to day administration tasks. On a technical level, it is a wrapper around b3scale's REST API and adds some convenience
features such as testing room creation for frontends, or displaying frontend settings conveniently.

## Installing b3scalectl

=== "Debian/Ubuntu"

    Starting with version 1.0.3, b3scale provides `.deb` packages for use in Ubuntu. Download the `b3scaled-*.deb` asset from the [release page](https://github.com/b3scale/b3scale/releases) on GitHub. On an Intel/AMD 64 Bit system, install the `.deb` package like this:

    ```bash
    sudo dpkg -i b3scalectl_1.0.3_linux_amd64.deb
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

=== "Windows / MacOS / Other distributions"

    Pre-compiled binaries are available from the [release page](https://github.com/b3scale/b3scale/releases) on GitHub. Pick the `b3scalectl` archive required for your operating system and system architecture. Extract the `b3scalectl` binary and put it in e.g. `/usr/local/bin`

## Setting up auto-completion

`b3scalectl` ships with completions for the popular bash and zsh shells.

```bash
eval $(b3scalectl completions bash)
```

=== zsh

```zsh
eval $(b3scalectl completions zsh)
```

## Authorizing b3scalectl

b3scale will fetch an JWT token from the b3scale daemon. The easiest way is to provide the b3scale secret used in the `b3scaled` config:

```bash
b3scalectl --api https://bbb-api.example.com auth authorize

** Authorization required for https://infra.run **

Please paste your shared secret here. The generated
access token will be stored in:

   /.../b3scale/https_infra.run.access_token

Secret:
```

You can now use `b3scalectl` as described in the other chapters.