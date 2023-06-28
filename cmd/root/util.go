package root

import (
	"strings"

	"github.com/spf13/viper"
)

func prompt() string {
	prompt := "$"
	username := viper.GetString("username")
	server := viper.GetString("server")
	server = removeScheme(server)
	if username != "" && server != "" {
		prompt = username + "@" + server + prompt
	} else if username != "" {
		prompt = server + prompt
	} else if server != "" {
		prompt = username + prompt
	}
	return prompt
}

func removeScheme(url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url[strings.Index(url, "//")+2:]
	}
	return url
}
