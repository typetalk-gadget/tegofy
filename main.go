package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gen2brain/beeep"
	v1 "github.com/nulab/go-typetalk/v3/typetalk/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/typetalk-gadget/go-typetalk-stream/stream"
	"github.com/typetalk-gadget/go-typetalk-token-source/source"
	"golang.org/x/oauth2"
)

const cmdName = "tegofy"

var (
	envPrefix = strings.ToUpper(cmdName)
	config    Config
)

const (
	flagNameConfig         = "config"
	flagNameDebug          = "debug"
	flagNameClientID       = "client_id"
	flagNameClientSecret   = "client_secret"
	flagNameNotifyDesktop  = "notify_desktop"
	flagNameNotifyTypetalk = "notify_typetalk"
	flagNameWithMention    = "with_mention"
	flagNameSpaceKeys      = "space_keys"
	flagNameKeywords       = "keywords"
	flagNameIgnoreBot      = "ignore_bot"
	flagNameIgnoreUsers    = "ignore_users"
)

type Config struct {
	Debug          bool          `mapstructure:"debug"`
	ClientID       string        `mapstructure:"client_id"`
	ClientSecret   string        `mapstructure:"client_secret"`
	NotifyDesktop  bool          `mapstructure:"notify_desktop"`
	NotifyTypetalk int           `mapstructure:"notify_typetalk"`
	WithMention    bool          `mapstructure:"with_mention"`
	SpaceKeys      []string      `mapstructure:"space_keys"`
	Keywords       []Keyword     `mapstructure:"keywords"`
	IgnoreBot      bool          `mapstructure:"ignore_bot"`
	IgnoreUsers    []interface{} `mapstructure:"ignore_users"`
}

type Keyword struct {
	Keyword string `mapstructure:"keyword"`
	TopicID int    `mapstructure:"topic_id"`
}

func main() {

	rootCmd := &cobra.Command{
		Use:     cmdName,
		Run:     run,
		Version: FmtVersion(),
	}

	flags := rootCmd.PersistentFlags()

	flags.StringP(flagNameConfig, "c", "config.yml", "config file path")
	flags.Bool(flagNameDebug, false, "debug mode")
	flags.String(flagNameClientID, "", "typetalk client id [TEGOFY_CLIENT_ID]")
	flags.String(flagNameClientSecret, "", "typetalk client secret [TEGOFY_CLIENT_SECRET]")
	flags.Bool(flagNameNotifyDesktop, false, "enable desktop notifications [TEGOFY_NOTIFY_DESKTOP]")
	flags.Int(flagNameNotifyTypetalk, 0, "enable typetalk notifications with topic id [TEGOFY_NOTIFY_TYPETALK]")
	flags.Bool(flagNameWithMention, false, "with mentions in notifications[TEGOFY_WITH_MENTION]")
	flags.StringSlice(flagNameSpaceKeys, nil, "keys of space to include in search [TEGOFY_SPACE_KEYS]")
	flags.StringSlice(flagNameKeywords, nil, "matching keywords [TEGOFY_KEYWORDS]")
	flags.Bool(flagNameIgnoreBot, false, "ignore bot posts [TEGOFY_IGNORE_BOT]")
	flags.StringSlice(flagNameIgnoreUsers, nil, "ignore user posts [TEGOFY_IGNORE_USERS]")

	_ = viper.BindPFlag(flagNameDebug, flags.Lookup(flagNameDebug))
	_ = viper.BindPFlag(flagNameClientID, flags.Lookup(flagNameClientID))
	_ = viper.BindPFlag(flagNameClientSecret, flags.Lookup(flagNameClientSecret))
	_ = viper.BindPFlag(flagNameNotifyDesktop, flags.Lookup(flagNameNotifyDesktop))
	_ = viper.BindPFlag(flagNameNotifyTypetalk, flags.Lookup(flagNameNotifyTypetalk))
	_ = viper.BindPFlag(flagNameWithMention, flags.Lookup(flagNameWithMention))
	_ = viper.BindPFlag(flagNameSpaceKeys, flags.Lookup(flagNameSpaceKeys))
	_ = viper.BindPFlag(flagNameIgnoreBot, flags.Lookup(flagNameIgnoreBot))
	_ = viper.BindPFlag(flagNameIgnoreUsers, flags.Lookup(flagNameIgnoreUsers))

	cobra.OnInitialize(func() {
		configFile, err := flags.GetString(flagNameConfig)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		viper.SetConfigFile(configFile)
		viper.SetConfigType("yaml")
		viper.SetEnvPrefix(envPrefix)
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
		viper.AutomaticEnv()
		if err := viper.ReadInConfig(); err != nil {
			printError("failed to read config", err)
			os.Exit(1)
		}

		if err := viper.Unmarshal(&config); err != nil {
			printError("failed to unmarshal config", err)
			os.Exit(1)
		}

		flagKeywords, err := flags.GetStringSlice(flagNameKeywords)
		if err != nil {
			printError("failed to get keywords", err)
			os.Exit(1)
		}

		for _, v := range flagKeywords {
			config.Keywords = append(config.Keywords, Keyword{
				Keyword: v,
			})
		}

	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var myAccount *v1.Account

func run(c *cobra.Command, args []string) {

	if config.Debug {
		printDebug(fmt.Sprintf("config: %#v\n", config))
	}

	scope := "my topic.read"
	if config.NotifyTypetalk > 0 {
		scope += " topic.post"
	}

	tokenSource := &source.TokenSource{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Scope:        scope,
	}

	tc := oauth2.NewClient(context.Background(), tokenSource)
	api := v1.NewClient(tc)

	myProfile, _, err := api.Accounts.GetMyProfile(context.Background())
	if err != nil {
		printError("failed to get my profile", err)
		return
	}
	myAccount = myProfile.Account
	s := stream.Stream{
		TokenSource:  tokenSource,
		Handler:      notify(api),
		PingInterval: 30 * time.Second,
		LoggerFunc:   log.Println,
	}

	go func() {
		printInfo("start to subscribe typetalk stream")
		err := s.Subscribe()
		if err == stream.ErrStreamClosed {
			return
		}
		if err != nil {
			printError("failed to subscribe", err)
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)

	<-sigint

	printInfo("received a signal of graceful shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	err = s.Shutdown(ctx)
	if err != nil {
		printError("failed to graceful shutdown", err)
		return
	}
	printInfo("completed graceful shutdown")
}

func isTargetSpace(msg *stream.Message) bool {
	for _, v := range config.SpaceKeys {
		if msg.Data.Space != nil && v == msg.Data.Space.Key {
			return true
		}
	}
	return false
}

func isPostMessage(msg *stream.Message) bool {
	return msg.Type == "postMessage"
}

func isNotifyTopic(msg *stream.Message) bool {
	return config.NotifyTypetalk > 0 && msg.Data.Topic.ID == config.NotifyTypetalk
}

func isMention(msg *stream.Message) bool {
	return strings.Contains(msg.Data.Post.Message, myAccount.Name)
}

func isDM(msg *stream.Message) bool {
	return msg.Data.DirectMessage != nil
}

func isBot(msg *stream.Message) bool {
	return config.IgnoreBot && msg.Data.Post.Account.IsBot
}

func isIgnoreUser(msg *stream.Message) bool {
	var match bool
	for _, name := range config.IgnoreUsers {
		if msg.Data.Post.Account.Name == name {
			match = true
			break
		}
	}
	return match
}

func containsKeyWords(msg *stream.Message) []string {
	var matches []string
	for _, v := range config.Keywords {
		if len(v.Keyword) > 0 && strings.Contains(msg.Data.Post.Message, v.Keyword) {
			if v.TopicID <= 0 {
				matches = append(matches, v.Keyword)
				continue
			}
			if v.TopicID == msg.Data.Topic.ID {
				matches = append(matches, v.Keyword)
			}
		}
	}
	return matches
}

func notify(api *v1.Client) stream.Handler {
	return stream.HandlerFunc(func(msg *stream.Message) {
		if !isTargetSpace(msg) {
			return
		}
		if !isPostMessage(msg) {
			return
		}
		if isNotifyTopic(msg) {
			return
		}
		if isMention(msg) {
			return
		}
		if isDM(msg) {
			return
		}
		if isBot(msg) {
			return
		}
		if isIgnoreUser(msg) {
			return
		}
		matches := containsKeyWords(msg)
		if len(matches) == 0 {
			return
		}

		postURL := fmt.Sprintf(`https://typetalk.com/topics/%d/posts/%d`,
			msg.Data.Topic.ID, msg.Data.Post.ID)

		if config.NotifyDesktop {
			var post strings.Builder
			post.WriteString(postURL)
			post.WriteString("\n")
			post.WriteString(msg.Data.Post.Message)
			beeep.Notify(
				msg.Data.Topic.Name,
				post.String(),
				"")
		}

		if config.NotifyTypetalk > 0 {
			var post strings.Builder
			post.WriteString(postURL)
			post.WriteString("\n")
			post.WriteString(fmt.Sprintf("matches: %v", matches))
			post.WriteString("\n")
			if config.WithMention {
				post.WriteString(fmt.Sprintf(`@%s by tegofy`, myAccount.Name))
			}
			_, _, err := api.Messages.PostMessage(context.Background(),
				config.NotifyTypetalk, post.String(), nil)
			if err != nil {
				printError("failed to notify typetalk:", err)
			}
		}
	})
}

func printDebug(args ...interface{}) {
	args = append([]interface{}{cmdName + ":", "[DEBUG]"}, args...)
	log.Println(args...)
}

func printInfo(args ...interface{}) {
	args = append([]interface{}{cmdName + ":", "[INFO]"}, args...)
	log.Println(args...)
}

func printError(args ...interface{}) {
	args = append([]interface{}{cmdName + ":", "[ERROR]"}, args...)
	log.Println(args...)
}
