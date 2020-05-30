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
      --clientId string       typetalk client id [TEGOFY_CLIENTID]
      --clientSecret string   typetalk client secret [TEGOFY_CLIENTSECRET]
  -c, --config string         config file path (default "config.yml")
      --debug                 debug mode
  -h, --help                  help for tegofy
      --ignoreBot             ignore bot posts
      --ignoreUsers strings   ignore user posts
      --keywords strings      matching keywords
      --notifyDesktop         enable desktop notifications
      --notifyTypetalk int    enable typetalk notifications with topic id
      --spaceKeys strings     keys of space to include in search
      --withMention           with mentions in notifications
```

## Acknowledgments

[futahashi](https://github.com/futahashi) for the core concept of tegofy

## Bugs and Feedback

For bugs, questions and discussions please use the GitHub Issues.

## License

[MIT License](http://www.opensource.org/licenses/mit-license.php)

## Author

[vvatanabe](https://github.com/vvatanabe)