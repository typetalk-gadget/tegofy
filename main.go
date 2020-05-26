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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vvatanabe/go-typetalk-stream/stream"
)

var (
	configFile string
	envPrefix  string = "TEGOFY"

	config Config
)

func main() {

	rootCmd := &cobra.Command{
		Use: "tegofy",
		Run: run,
	}

	flags := rootCmd.PersistentFlags()

	flags.StringVarP(&configFile, "config", "c", "config.yml", "config file name")
	flags.Bool("debug", false, "debug mode")
	flags.String("clientId", "", "typetalk client id [TEGOFY_APP_CLIENTID]")
	flags.String("clientSecret", "", "typetalk client secret [TEGOFY_APP_CLIENTSECRET]")
	flags.StringSlice("keywords", nil, "matching keywords")
	flags.String("mention", "", "ignore mention to you")
	flags.StringSlice("spaceKeys", nil, "keys of space to include in search")
	flags.Bool("ignoreBot", false, "ignore bot posts")
	flags.StringSlice("ignoreUsers", nil, "ignore user posts")
	flags.Bool("desktopNotify", false, "enable desktop notifications")
	flags.Bool("typetalkNotify", true, "enable typetalk notifications")

	_ = viper.BindPFlag("debug", flags.Lookup("debug"))
	_ = viper.BindPFlag("app.clientId", flags.Lookup("clientId"))
	_ = viper.BindPFlag("app.clientSecret", flags.Lookup("clientSecret"))
	_ = viper.BindPFlag("watch.mention", flags.Lookup("mention"))
	_ = viper.BindPFlag("watch.spaceKeys", flags.Lookup("spaceKeys"))
	_ = viper.BindPFlag("watch.ignoreBot", flags.Lookup("ignoreBot"))
	_ = viper.BindPFlag("watch.ignoreUsers", flags.Lookup("ignoreUsers"))
	_ = viper.BindPFlag("message.desktopNotify", flags.Lookup("desktopNotify"))
	_ = viper.BindPFlag("message.typetalkNotify", flags.Lookup("typetalkNotify"))

	cobra.OnInitialize(func() {
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
			config.Watch.Keywords = append(config.Watch.Keywords, Keyword{
				Keyword: v,
				Mention: false,
			})
		}

	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(c *cobra.Command, args []string) {

	// debug config
	// fmt.Printf("config: %#v\n", config.Watch.Keywords)

	s := stream.Stream{
		ClientID:     config.App.ClientID,
		ClientSecret: config.App.ClientSecret,
		Handler:      stream.HandlerFunc(notify),
		PingInterval: 30 * time.Second,
		LoggerFunc:   log.Println,
	}

	go func() {
		log.Println("start to subscribe typetalk stream")
		err := s.Subscribe()
		//if err == stream.ErrStreamClosed {
		//	return
		//}
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
	err := s.Shutdown(ctx)
	if err != nil {
		log.Println("failed to graceful shutdown", err)
	} else {
		log.Println("completed graceful shutdown")
	}
}

func isTargetSpace(msg *stream.Message) bool {
	for _, v := range config.Watch.SpaceKeys {
		if msg.Data.Space != nil && v == msg.Data.Space.Key {
			return true
		}
	}
	return false
}

func isPostMessage(msg *stream.Message) bool {
	return msg.Type == "postMessage"
}

func isMention(msg *stream.Message) bool {
	return strings.Contains(msg.Data.Post.Message, config.Watch.Mention)
}

func isDM(msg *stream.Message) bool {
	return msg.Data.DirectMessage != nil
}

func isBot(msg *stream.Message) bool {
	return msg.Data.Post.Account.IsBot
}

func containsKeyWord(msg *stream.Message) bool {
	for _, v := range config.Watch.Keywords {
		if strings.Contains(msg.Data.Post.Message, v.Keyword) {
			return true
		}
	}
	return false
}

func notify(msg *stream.Message) {
	if !isTargetSpace(msg) {
		return
	}
	if !isPostMessage(msg) {
		return
	}
	if isMention(msg) {
		return
	}
	if isDM(msg) {
		return
	}
	if config.Watch.IgnoreBot && isBot(msg) {
		return
	}
	if !containsKeyWord(msg) {
		return
	}
	if config.Message.DesktopNotify {
		postURL := fmt.Sprintf(`https://typetalk.com/topics/%d/posts/%d`,
			msg.Data.Topic.ID, msg.Data.Post.ID)
		beeep.Notify(
			msg.Data.Topic.Name,
			postURL+" : "+msg.Data.Post.Message,
			"")
	}
	if config.Message.TypetalkNotify {
		// TODO
	}
}
