package commands

import (
	"strings"
	"time"

	twitch "github.com/gempir/go-twitch-irc"
)

// Command struct
type Command struct {
	client  *twitch.Client
	channel string
}

var (
	commandMap        map[string]fn
	project           string
	admins            = []string{"swarmlogic", "aidenmontgomery", "swarmlogic_bot"}
	badgesWeCareAbout = []string{"broadcaster", "moderator"}
	pollInProgress    = false
	currentPoll       *poll
	startTime         time.Time
	channel           string
)

// NewCommand returns an instantiated Command struct
func NewCommand(client *twitch.Client, twitchChannel string) *Command {
	project = "programming"
	startTime = time.Now()
	channel = twitchChannel
	commandMap = map[string]fn{
		"time":       getTime,
		"project":    getProject,
		"setproject": setProject,
		"8ball":      eightball,
		"uptime":     uptime,
		"poll":       makePoll,
		"vote":       votePoll,
		"options":    optionsPoll,
		"github":     github,
		"commands":   getCommands,
	}
	return &Command{client, channel}
}

// HandleCommand handles...
func (c *Command) HandleCommand(cmd, channel string, user twitch.User, message twitch.Message) {
	if user.Username == "swarmlogic_bot" {
		return
	}
	parsed := strings.Split(cmd, " ")
	cArgs := &commandArgs{parsed[1:], user, c.client}
	for k, v := range commandMap {
		if k == parsed[0] {
			c.client.Say(channel, v(cArgs))
			return
		}
	}
	c.client.Say(channel, user.DisplayName+", wtf you talkin about willis!?")
}
