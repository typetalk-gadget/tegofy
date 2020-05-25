package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/vvatanabe/go-typetalk-stream/stream"
	"github.com/vvatanabe/go-typetalk-stream/stream/tool"
	"golang.org/x/net/context"
)

func main() {

	log.SetFlags(0)

	var c tool.Chain
	c.Use(isTargetSpace)
	c.Use(isPostMessage)
	c.Use(isKeyWord)

	stream := stream.Stream{
		ClientID:     os.Getenv("TYPETALK_CLIENT_ID"),
		ClientSecret: os.Getenv("TYPETALK_CLIENT_SECRET"),
		Handler:      c.Then(notify),
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
		if msg.Data.Space.Key == os.Getenv("TARGET_SPACE_KEY") {
			next.Serve(msg)
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

func isKeyWord(next stream.HandlerFunc) stream.HandlerFunc {
	return stream.HandlerFunc(func(msg *stream.Message) {
		if strings.Contains(msg.Data.Post.Message, os.Getenv("TARGET_KEYWORD")) {
			next.Serve(msg)
		}
	})
}

func notify(msg *stream.Message) {
	postURL := fmt.Sprintf(`https://typetalk.com/topics/%d/posts/%d`,
		msg.Data.Topic.ID, msg.Data.Post.ID)

	beeep.Notify(
		msg.Data.Topic.Name,
		msg.Data.Post.Message+" => "+postURL,
		"")
}
