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

type votePackage struct {
	vote []string
	user string
}

type poll struct {
	cmdChan    chan string
	voteChan   chan votePackage
	duration   time.Duration
	usersVoted map[string]string
	tally      map[string]int
	options    map[int]string
}

func newPoll(args []string) (*poll, error) {
	duration, err := time.ParseDuration(args[0])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse specified time for poll")
	}
	providedOptions := args[1:]
	tally := make(map[string]int)
	options := make(map[int]string)
	usersVoted := make(map[string]string)
	cmdChan := make(chan string)
	voteChan := make(chan votePackage)
	for i := range providedOptions {
		tally[providedOptions[i]] = 0
		options[i] = providedOptions[i]
	}
	return &poll{cmdChan, voteChan, duration, usersVoted, tally, options}, nil
}

func (p *poll) countVote(vp votePackage, c *twitch.Client) {
	if _, ok := p.usersVoted[vp.user]; ok {
		c.Whisper(vp.user, "You've already voted for "+vp.vote[0])
		return
	}
	userVote, err := strconv.Atoi(vp.vote[0])
	if err != nil {
		return
	}
	if userVote > len(currentPoll.options)-1 {
		return
	}
	option := p.options[userVote]
	if _, ok := p.tally[option]; !ok {
		return
	}
	p.tally[option]++
	p.usersVoted[vp.user] = vp.vote[0]

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
				go func() {
					t := time.NewTimer(p.duration)
					<-t.C
					p.cmdChan <- "complete"
				}()

			V:
				for {
					select {
					case msg := <-p.cmdChan:
						switch msg {
						case "complete":
							client.Say(channel, "Poll complete! Results are:")
							for k, v := range p.tally {
								client.Say(channel, fmt.Sprintf("%s got %d votes", k, v))
							}
							pollInProgress = false
							break V
						}
					case msg := <-p.voteChan:
						p.countVote(msg, client)
					}
				}
			}(currentPoll)
			client.Say(channel, "A poll was created! !vote")
			for k, v := range currentPoll.options {
				client.Say(channel, fmt.Sprintf("For %s, !vote %d", v, k))
			}
			return ""
		},
		"vote": func(args []string, user twitch.User) string {
			if pollInProgress {
				go func() {
					currentPoll.voteChan <- votePackage{args, user.Username}
				}()
			}
			return ""
		},
		"options": func(_ []string, _ twitch.User) string {
			if pollInProgress {
				for k, v := range currentPoll.options {
					client.Say(channel, fmt.Sprintf("For %s, !vote %d", v, k))
				}
				return ""
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
