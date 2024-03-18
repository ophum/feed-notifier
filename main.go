package main

import (
	"log"
	"os"
	"sort"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/ophum/feed-notifier/internal"
	"gopkg.in/yaml.v3"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

var config internal.Config
var cursor *internal.Cursor

var notifiers = map[string]*internal.Notifier{}

func loadConfig() error {
	f, err := os.Open("config.yaml")
	if err != nil {
		return err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&config); err != nil {
		return err
	}

	for _, v := range config.Notifiers {
		log.Println("added notifier:", v.Name, v.Slack.WebhookURL)
		notifiers[v.Name] = v
	}

	return nil
}

func loadCursor() (err error) {
	cursor, err = internal.NewCursor(config.CursorFile)
	return
}

func init() {
	if err := loadConfig(); err != nil {
		panic(err)
	}

	if err := loadCursor(); err != nil {
		panic(err)
	}
}

func main() {
	for _, feedSrc := range config.Feeds {
		fp := gofeed.NewParser()
		feed, err := fp.ParseURL(feedSrc.URL.String())
		if err != nil {
			log.Println(err)
			continue
		}

		sort.Slice(feed.Items, func(i, j int) bool {
			return feed.Items[i].UpdatedParsed.Before(*feed.Items[j].UpdatedParsed)
		})

		if notifier, ok := notifiers[feedSrc.Notifier]; ok {
			notifierCursor := cursor.GetNotifierCursor(notifier.Name)

			lastUpdatedAt, ok := notifierCursor[feedSrc.URL.String()]
			if !ok {
				lastUpdatedAt = time.Time{}
			}

			for _, item := range feed.Items {
				if lastUpdatedAt.Before(*item.UpdatedParsed) {
					converter := md.NewConverter("", true, nil)
					markdown, err := converter.ConvertString(item.Content)
					if err != nil {
						markdown = item.Content
					}
					log.Println("notice:", notifier.Name, "type:", notifier.Type, "webhookURL:", notifier.Slack.WebhookURL)
					log.Println("to:", feedSrc.Mentions)
					log.Println("message:", feedSrc.Message)
					log.Println("feed title", item.Title)
					log.Println("feed content\n", markdown)

					lastUpdatedAt = *item.UpdatedParsed
				}
			}
			notifierCursor[feedSrc.URL.String()] = lastUpdatedAt
			cursor.SetNotifierCursor(notifier.Name, notifierCursor)
		} else {
			log.Println("notifier", feedSrc.Notifier, "not found")
		}
	}

	if err := cursor.Write(config.CursorFile); err != nil {
		panic(err)
	}
}
