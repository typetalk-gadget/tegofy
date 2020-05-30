package main

type Config struct {
	Debug          bool          `yaml:"debug"`
	ClientID       string        `yaml:"clientId"`
	ClientSecret   string        `yaml:"clientSecret"`
	NotifyDesktop  bool          `yaml:"notifyDesktop"`
	NotifyTypetalk int           `yaml:"notifyTypetalk"`
	WithMention    bool          `yaml:"withMention"`
	SpaceKeys      []string      `yaml:"spaceKeys"`
	Keywords       []Keyword     `yaml:"keywords"`
	IgnoreBot      bool          `yaml:"ignoreBot"`
	IgnoreUsers    []interface{} `yaml:"ignoreUsers"`
}

type Keyword struct {
	Keyword string `yaml:"keyword"`
	TopicID int    `yaml:"topicId"`
}
