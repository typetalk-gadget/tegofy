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

const envPrefix = "TEGOFY"

var (
	config Config
)

func main() {

	rootCmd := &cobra.Command{
		Use: "tegofy",
		Run: run,
	}

	flags := rootCmd.PersistentFlags()

	flags.StringP("config", "c", "config.yml", "config file path")
	flags.Bool("debug", false, "debug mode")
	flags.String("clientId", "", "typetalk client id [TEGOFY_CLIENTID]")
	flags.String("clientSecret", "", "typetalk client secret [TEGOFY_CLIENTSECRET]")
	flags.Bool("notifyDesktop", false, "enable desktop notifications")
	flags.Int("notifyTypetalk", 0, "enable typetalk notifications with topic id")
	flags.Bool("withMention", false, "with mentions in notifications")
	flags.StringSlice("spaceKeys", nil, "keys of space to include in search")
	flags.StringSlice("keywords", nil, "matching keywords")
	flags.Bool("ignoreBot", false, "ignore bot posts")
	flags.StringSlice("ignoreUsers", nil, "ignore user posts")

	_ = viper.BindPFlag("debug", flags.Lookup("debug"))
	_ = viper.BindPFlag("clientId", flags.Lookup("clientId"))
	_ = viper.BindPFlag("clientSecret", flags.Lookup("clientSecret"))
	_ = viper.BindPFlag("notifyDesktop", flags.Lookup("notifyDesktop"))
	_ = viper.BindPFlag("notifyTypetalk", flags.Lookup("notifyTypetalk"))
	_ = viper.BindPFlag("withMention", flags.Lookup("withMention"))
	_ = viper.BindPFlag("spaceKeys", flags.Lookup("spaceKeys"))
	_ = viper.BindPFlag("ignoreBot", flags.Lookup("ignoreBot"))
	_ = viper.BindPFlag("ignoreUsers", flags.Lookup("ignoreUsers"))

	cobra.OnInitialize(func() {
		configFile, err := flags.GetString("config")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		viper.SetConfigFile(configFile)
		viper.SetEnvPrefix(envPrefix)
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
		viper.AutomaticEnv()
		if err := viper.ReadInConfig(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err := viper.Unmarshal(&config); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		flagKeywords, err := flags.GetStringSlice("keywords")
		if err != nil {
			fmt.Println(err)
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

	// debug config
	// fmt.Printf("config: %#v\n", config.Watch.Keywords)

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
		log.Println("failed to get my profile")
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
		log.Println("start to subscribe typetalk stream")
		err := s.Subscribe()
		if err == stream.ErrStreamClosed {
			return
		}
		if err != nil {
			log.Println("failed to subscribe", err)
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)

	<-sigint

	log.Println("received a signal of graceful shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	err = s.Shutdown(ctx)
	if err != nil {
		log.Println("failed to graceful shutdown", err)
	} else {
		log.Println("completed graceful shutdown")
	}
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
		if strings.Contains(msg.Data.Post.Message, v.Keyword) {
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
				log.Println("failed to notify typetalk:", err)
			}
		}
	})
}
