package commands

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	twitch "github.com/gempir/go-twitch-irc"
)

type commandArgs struct {
	args   []string
	user   twitch.User
	client *twitch.Client
}

type fn func(c *commandArgs) string

func getTime(c *commandArgs) string {
	return time.Now().Format(time.RFC1123Z)
}
func getProject(c *commandArgs) string {
	return project
}
func setProject(c *commandArgs) string {
	if !checkAdmin(c.user.Username) {
		return "You aren't authorized to perform that command @" + c.user.Username
	}
	project = strings.Join(c.args, " ")
	return "project set to: " + project
}

func getCommands(c *commandArgs) string {
	availableCommands := ""
	for k := range commandMap {
		availableCommands += "!" + k + " "
	}
	return fmt.Sprintf("Available commmands are: %s", availableCommands)
}

func eightball(c *commandArgs) string {
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
}
func uptime(c *commandArgs) string {
	return fmt.Sprintf("Chatbot up for %v", time.Since(startTime))
}
func makePoll(c *commandArgs) string {
	if pollInProgress {
		return "There is currently a poll in progress, type !options to see the options"
	}
	for badge := range c.user.Badges {
		if !isAllowedPoll(badge) {
			return "You aren't allowed to start a poll"
		}
	}
	var err error
	currentPoll, err = newPoll(c.args)
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
					c.client.Say(channel, "Poll complete! Results are:")
					for k, v := range p.tally {
						c.client.Say(channel, fmt.Sprintf("%s got %d votes", k, v))
					}
					pollInProgress = false
					break V
				}
			case msg := <-p.voteChan:
				p.countVote(msg, c.client)
			}
		}
	}(currentPoll)
	c.client.Say(channel, "A poll was created! !vote")
	for k, v := range currentPoll.options {
		c.client.Say(channel, fmt.Sprintf("For %s, !vote %d", v, k))
	}
	return ""
}
func votePoll(c *commandArgs) string {
	go func() {
		currentPoll.voteChan <- votePackage{c.args, c.user.Username}
	}()
	return ""
}
func optionsPoll(c *commandArgs) string {
	if pollInProgress {
		for k, v := range currentPoll.options {
			c.client.Say(channel, fmt.Sprintf("For %s, !vote %d", v, k))
		}
		return ""
	}
	return "No poll in progress"
}
func github(c *commandArgs) string {
	return "https://github.com/terakilobyte/chatbot"
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
