package commands

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	twitch "github.com/gempir/go-twitch-irc"
)

// Command struct
type Command struct {
	client *twitch.Client
}

type fn func(args []string, user twitch.User) string

var commandMap map[string]fn
var project string
var admins = []string{"swarmlogic"}

func checkAdmin(username string) bool {
	for i := range admins {
		if username == admins[i] {
			return true
		}
	}
	return false
}

// NewCommand returns an instantiated Command struct
func NewCommand(client *twitch.Client) *Command {
	project = "programming"
	startTime := time.Now()
	commandMap = map[string]fn{
		"time": func(_ []string, _ twitch.User) string {
			return time.Now().Format(time.RFC1123Z)
		},
		"project": func(_ []string, _ twitch.User) string {
			return project
		},
		"setproject": func(args []string, user twitch.User) string {
			if !checkAdmin(user.Username) {
				return "You aren't authorized to perform that command @" + user.Username
			}
			project = strings.Join(args, " ")
			return "project set to: " + project
		},
		"8ball": func(args []string, _ twitch.User) string {
			answers := []string{
				"It is certain",
				"It is decidedly so",
				"Without a doubt",
				"Yes - definitely",
				"You may rely on it",
				"As I see it, yes",
				"Most likely",
				"Outlook good",
				"Yes",
				"Signs point to yes",
				"Reply hazy, try again",
				"Ask again later",
				"Better not tell you now",
				"Cannot predict now",
				"Concentrate and ask again",
				"Don't count on it",
				"My reply is no",
				"My sources say no",
				"Outlook not so good",
				"Very doubtful",
			}
			rand.Seed(time.Now().Unix())
			return answers[rand.Intn(len(answers))]
		},
		"uptime": func(_ []string, _ twitch.User) string {
			return fmt.Sprintf("Chatbot up for %v", time.Now().Sub(startTime))
		},
	}
	return &Command{client}
}

// HandleCommand handles...
func (c *Command) HandleCommand(cmd, channel string, user twitch.User, message twitch.Message) {
	if user.Username == "swarmlogic_bot" {
		return
	}
	parsed := strings.Split(cmd, " ")
	for k, v := range commandMap {
		if k == parsed[0] {
			c.client.Say(channel, v(parsed[1:], user))
			return
		}
	}
	c.client.Say(channel, user.DisplayName+", wtf you talkin about willis!?")
}
