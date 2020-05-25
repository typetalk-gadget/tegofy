package main

type App struct {
	ClientID     string `yaml:"clientId"`
	ClientSecret string `yaml:"clientSecret"`
}

type Keyword struct {
	Keyword string `yaml:"keyword"`
	Mention bool   `yaml:"mention"`
}

type Watch struct {
	Mention     string    `yaml:"mention"`
	Keywords    []Keyword `yaml:"keywords"`
	SpaceKeys   []string  `yaml:"spaceKeys"`
	IgnoreBot   bool      `yaml:"ignoreBot"`
	IgnoreUsers []string  `yaml:"ignoreUsers"`
}

type Config struct {
	Debug   bool  `yaml:"debug"`
	App     App   `yaml:"app"`
	Watch   Watch `yaml:"watch"`
	Message struct {
		Mention            string `yaml:"mention"`
		DesktopNotify      bool   `yaml:"desktopNotify"`
		TypetalkNotify     bool   `yaml:"typetalkNotify"`
		NotifyTopicID      int    `yaml:"notifyTopicId"`
		MaxTopicNameLength int    `yaml:"maxTopicNameLength"`
		MaxMessageLength   int    `yaml:"maxMessageLength"`
		MaxMessageLine     int    `yaml:"maxMessageLine"`
		HidePostURL        bool   `yaml:"hidePostUrl"`
		Emoji              struct {
			Topic      string `yaml:"topic"`
			User       string `yaml:"user"`
			Like       string `yaml:"like"`
			UserStatus string `yaml:"userStatus"`
			Notify     string `yaml:"notify"`
		} `yaml:"emoji"`
		Color struct {
			Topic      string `yaml:"topic"`
			Like       string `yaml:"like"`
			User       string `yaml:"user"`
			Message    string `yaml:"message"`
			URL        string `yaml:"url"`
			Log        string `yaml:"log"`
			Omit       string `yaml:"omit"`
			Notify     string `yaml:"notify"`
			UserStatus string `yaml:"userStatus"`
		} `yaml:"color"`
		OmitWord string `yaml:"omitWord"`
	} `yaml:"message"`
	Endpoints struct {
		Typetalk struct {
			URL   string `yaml:"url"`
			Paths struct {
				AccessToken string `yaml:"accessToken"`
			} `yaml:"paths"`
		} `yaml:"typetalk"`
		Message struct {
			URL   string `yaml:"url"`
			Paths struct {
				Streaming string `yaml:"streaming"`
			} `yaml:"paths"`
		} `yaml:"message"`
	} `yaml:"endpoints"`
}
