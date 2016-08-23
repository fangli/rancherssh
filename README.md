Rancher SSH
===========

Native SSH Client for Rancher Containers, provided a powerful native terminal to manage your docker containers

  * It's dead simple. like the ssh cli, you do `rancherssh container_name` to SSH into any containers
  * It's flexible. rancherssh reads configurations from ENV, from yml or json file
  * It's powerful. rancherssh searches the whole rancher deployment, SSH into any containers from your workstation, regardless which host it belongs to
  * It's smart. rancherssh uses fuzzy container name matching. Forget the container name? it doesn't matter, use "*" or "%" instead

[![asciicast](demo.gif)](https://asciinema.org/a/83555)

[![asciicast](https://asciinema.org/a/83555.png)](https://asciinema.org/a/83555)

Installation
============

**Homebrew**

`# brew install fangli/dev/rancherssh`


**Or via Golang**

`# go get github.com/fangli/rancherssh`



usage
=====

`rancherssh [<flags>] <container>`

Example
=======

  rancherssh my-server-1
  
  rancherssh "my-server*"  (equals to) rancherssh my-server%
  
  rancherssh %proxy%
  
  rancherssh "projectA-app-*" (equals to) rancherssh projectA-app-%

Configuration
=============

  We read configuration from config.json or config.yml in ./, /etc/rancherssh/ and ~/.rancherssh/ folders.

  If you want to use JSON format, create a config.json in the folders with content:

      {
          "endpoint": "https://rancher.server/v1", // Or "https://rancher.server/v1/projects/xxxx"
          "user": "your_access_key",
          "password": "your_access_password"
      }

  If you want to use YAML format, create a config.yml with content:

      endpoint: https://your.rancher.server // Or https://rancher.server/v1/projects/xxxx
      user: your_access_key
      password: your_access_password

  We accept environment variables as well:

      SSHRANCHER_ENDPOINT=https://your.rancher.server   // Or https://rancher.server/v1/projects/xxxx
      SSHRANCHER_USER=your_access_key
      SSHRANCHER_PASSWORD=your_access_password


Flags
=====

      -h, --help         Show context-sensitive help (also try --help-long and --help-man).
      --version      Show application version.
      --endpoint=""  Rancher server endpoint, https://your.rancher.server/v1 or https://your.rancher.server/v1/projects/xxx.
      --user=""      Rancher API user/accesskey.
      --password=""  Rancher API password/secret.

**Args**

`<container>  Container name, fuzzy match`
