# tegofy

A command line tool for ego-search notify in Typetalk.

## Description

tegofy observes Typetalk's posts and notifies you of posts that contain the specified keyword. It never misses the information you want. tegofy lets you choose desktop and Typetalk as notification destinations.

## Installation

### Go
If you have the Go(go1.14+) installed, you can also install it with go get command.

```sh
$ go get github.com/typetalk-gadget/tegofy
```

## Usage

```
Usage:
  tegofy [flags]

Flags:
      --client-id string       typetalk client id [TEGOFY_CLIENT_ID]
      --client-secret string   typetalk client secret [TEGOFY_CLIENT_SECRET]
  -c, --config string          config file path (default "config.yml")
      --debug                  debug mode
  -h, --help                   help for tegofy
      --ignore-bot             ignore bot posts
      --ignore-users strings   ignore user posts
      --keywords strings       matching keywords
      --notify-desktop         enable desktop notifications
      --notify-typetalk int    enable typetalk notifications with topic id
      --space-keys strings     keys of space to include in search
      --with-mention           with mentions in notifications
```

## Config File

### YAML

```yaml
debug: true
client_id: "deadbeef"
client_secret: "deadcode"
notify_dsesktop: true
notify_typetalk: 9999999999999999
with_mention: true
space_keys:
  - "your_space_key"
keywords:
  - keyword: "Hello"
    topic_id: 11111111
  - keyword: "bye"
    topic_id: 22222222
ignore_bot: false
ignore_users: []
```

## Acknowledgments

[futahashi](https://github.com/futahashi) for the core concept of tegofy

## Bugs and Feedback

For bugs, questions and discussions please use the GitHub Issues.

## License

[MIT License](http://www.opensource.org/licenses/mit-license.php)

## Author

[vvatanabe](https://github.com/vvatanabe)