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
	"github.com/vvatanabe/go-typetalk-stream/stream/tool"
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
	fmt.Printf("config: %#v\n", config)

	var chain tool.Chain
	chain.Use(isTargetSpace)
	chain.Use(isPostMessage)
	chain.Use(isNotMention)
	chain.Use(isNotDM)
	chain.Use(isKeyWord)

	stream := stream.Stream{
		ClientID:     config.App.ClientID,
		ClientSecret: config.App.ClientSecret,
		Handler:      chain.Then(notify),
		PingInterval: 30 * time.Second,
		LoggerFunc:   log.Println,
	}

	go func() {
		log.Println("start to subscribe typetalk stream")
		err := stream.Subscribe()
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
	err := stream.Shutdown(ctx)
	if err != nil {
		log.Println("failed to graceful shutdown", err)
	} else {
		log.Println("completed graceful shutdown")
	}
}

func isTargetSpace(next stream.HandlerFunc) stream.HandlerFunc {
	return stream.HandlerFunc(func(msg *stream.Message) {
		for _, v := range config.Watch.SpaceKeys {
			if v == msg.Data.Space.Key {
				next.Serve(msg)
				break
			}
		}
	})
}

func isPostMessage(next stream.HandlerFunc) stream.HandlerFunc {
	return stream.HandlerFunc(func(msg *stream.Message) {
		if msg.Type == "postMessage" {
			next.Serve(msg)
		}
	})
}

func isNotMention(next stream.HandlerFunc) stream.HandlerFunc {
	return stream.HandlerFunc(func(msg *stream.Message) {
		if !strings.Contains(msg.Data.Post.Message, config.Watch.Mention) {
			next.Serve(msg)
		}
	})
}

func isNotDM(next stream.HandlerFunc) stream.HandlerFunc {
	return stream.HandlerFunc(func(msg *stream.Message) {
		if !msg.Data.DirectMessage {
			next.Serve(msg)
		}
	})
}

//func isNotBot(next stream.HandlerFunc) stream.HandlerFunc {
//	return stream.HandlerFunc(func(msg *stream.Message) {
//		if config.Watch.IgnoreBot && msg.Data.Post.Account.IsBot {
//			next.Serve(msg)
//		}
//	})
//}

func isKeyWord(next stream.HandlerFunc) stream.HandlerFunc {
	return stream.HandlerFunc(func(msg *stream.Message) {
		if strings.Contains(msg.Data.Post.Message, os.Getenv("TARGET_KEYWORD")) {
			next.Serve(msg)
		}
	})
}

func notify(msg *stream.Message) {
	if config.Message.DesktopNotify {
		postURL := fmt.Sprintf(`https://typetalk.com/topics/%d/posts/%d`,
			msg.Data.Topic.ID, msg.Data.Post.ID)
		beeep.Notify(
			msg.Data.Topic.Name,
			msg.Data.Post.Message+" => "+postURL,
			"")
	}
	if config.Message.TypetalkNotify {
		// TODO
	}
}
