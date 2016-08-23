Rancher SSH
===========

Native SSH Client for Rancher Containers, provided a powerful native terminal to manage your docker containers

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
          "endpoint": "https://your.rancher.server",
          "user": "your_access_key",
          "password": "your_access_password"
      }

  If you want to use YAML format, create a config.yml with content:
  
      endpoint: https://your.rancher.server
      user: your_access_key
      password: your_access_password

  We accept environment variables as well:
  
      SSHRANCHER_ENDPOINT=https://your.rancher.server
      SSHRANCHER_USER=your_access_key
      SSHRANCHER_PASSWORD=your_access_password

Flags
=====

      -h, --help         Show context-sensitive help (also try --help-long and --help-man).
      --version      Show application version.
      --endpoint=""  Rancher server endpoint, https://your.rancher.server/, DO NOT include '/v1'.
      --user=""      Rancher API user/accesskey.
      --password=""  Rancher API password/secret.

**Args**

`<container>  Container name, fuzzy match`
