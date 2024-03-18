package internal

import (
	"net/url"

	"gopkg.in/yaml.v3"
)

type Config struct {
	CursorFile string      `yaml:"cursorFile"`
	Feeds      []*Feed     `yaml:"feeds"`
	Notifiers  []*Notifier `yaml:"notifiers"`
}

type NotifierType string

const (
	NotifierTypeSlack NotifierType = "slack"
)

type Notifier struct {
	Name  string        `yaml:"name"`
	Type  NotifierType  `yaml:"type"`
	Slack NotifierSlack `yaml:"slack"`
}

type NotifierSlack struct {
	WebhookURL *url.URL `yaml:"webhookURL"`
}

func (v *NotifierSlack) UnmarshalYAML(value *yaml.Node) error {
	slack := struct {
		WebhookURL string `yaml:"webhookURL"`
	}{}
	if err := value.Decode(&slack); err != nil {
		return err
	}

	u, err := url.Parse(slack.WebhookURL)
	if err != nil {
		return err
	}

	v.WebhookURL = u
	return nil
}

type Feed struct {
	URL      *url.URL
	Notifier string
	Mentions []string
	Message  string
}

func (f *Feed) UnmarshalYAML(value *yaml.Node) error {
	feed := struct {
		URL      string   `yaml:"url"`
		Notifier string   `yaml:"notifier"`
		Mentions []string `yaml:"mentions"`
		Message  string   `yaml:"message"`
	}{}
	if err := value.Decode(&feed); err != nil {
		return err
	}

	u, err := url.Parse(feed.URL)
	if err != nil {
		return err
	}
	f.URL = u
	f.Notifier = feed.Notifier
	f.Mentions = feed.Mentions
	f.Message = feed.Message

	return nil
}
