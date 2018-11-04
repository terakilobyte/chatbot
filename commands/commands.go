package commands

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	twitch "github.com/gempir/go-twitch-irc"
)

// Command struct
type Command struct {
	client  *twitch.Client
	channel string
}

type poll struct {
	duration time.Duration
	tally    map[string]int
	options  map[int]string
}

func newPoll(args []string) (*poll, error) {
	duration, err := time.ParseDuration(args[0])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse specified time for poll")
	}
	providedOptions := args[1:]
	tally := make(map[string]int)
	options := make(map[int]string)
	for i := range providedOptions {
		tally[providedOptions[i]] = 0
		options[i] = providedOptions[i]
	}
	return &poll{duration, tally, options}, nil
}

type fn func(args []string, user twitch.User) string

var commandMap map[string]fn
var project string
var admins = []string{"swarmlogic", "aidenmontgomery", "swarmlogic_bot"}
var badgesWeCareAbout = []string{"broadcaster", "moderator"}
var pollInProgress = false
var currentPoll *poll

func checkAdmin(username string) bool {
	for i := range admins {
		if username == admins[i] {
			return true
		}
	}
	return false
}

func isAllowedPoll(badge string) bool {
	for i := range badgesWeCareAbout {
		if badge == badgesWeCareAbout[i] {
			return true
		}
	}
	return false
}

// NewCommand returns an instantiated Command struct
func NewCommand(client *twitch.Client, channel string) *Command {
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
		"ident": func(_ []string, user twitch.User) string {
			for badge := range user.Badges {
				if isAllowedPoll(badge) {
					return fmt.Sprintf("You are a %s, %s", badge, user.DisplayName)
				}
			}
			return "You don't appear to have badges we care about " + user.DisplayName
		},
		"poll": func(args []string, user twitch.User) string {
			if pollInProgress {
				return "There is currently a poll in progress, type !options to see the options"
			}
			for badge := range user.Badges {
				if !isAllowedPoll(badge) {
					return "You aren't allowed to start a poll"
				}
			}
			var err error
			currentPoll, err = newPoll(args)
			if err != nil {
				return fmt.Sprintf("%v", err)
			}
			pollInProgress = true
			go func(p *poll) {
				t := time.NewTimer(p.duration)
				<-t.C
				client.Say(channel, fmt.Sprintf("Poll complete! The results were %v", p.tally))
				pollInProgress = false
			}(currentPoll)
			return fmt.Sprintf("A poll was created with options %v", currentPoll)
		},
		"vote": func(args []string, _ twitch.User) string {
			userVote, err := strconv.Atoi(args[0])
			if err != nil {
				return ""
			}
			if userVote > len(currentPoll.options)-1 {
				return ""
			}
			option := currentPoll.options[userVote]
			if _, ok := currentPoll.tally[option]; !ok {
				return ""
			}
			currentPoll.tally[option]++
			return ""
		},
		"options": func(_ []string, _ twitch.User) string {
			if pollInProgress {
				return fmt.Sprintf("%v", currentPoll.options)
			}
			return "No poll in progress"
		},
		"github": func(_ []string, _ twitch.User) string {
			return "https://github.com/terakilobyte/chatbot"
		},
	}
	return &Command{client, channel}
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
