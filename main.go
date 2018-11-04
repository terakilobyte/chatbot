package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gempir/go-twitch-irc"
	"github.com/joho/godotenv"
	"github.com/terakilobyte/chatbot/commands"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("poop")
	}
	twitchOauth := os.Getenv("TWITCH_OAUTH")
	twitchChannel := os.Getenv("TWITCH_ACCOUNT")
	client := twitch.NewClient(twitchChannel, twitchOauth)
	commandHandler := commands.NewCommand(client, twitchChannel)

	client.OnNewUnsetMessage(func(foobar string) {
		fmt.Println("Got an unset message")
		fmt.Println(foobar)
	})

	client.OnNewMessage(func(channel string, user twitch.User, message twitch.Message) {
		if user.Username == "swarmlogic_bot" {
			fmt.Println(message.Text)
		}
		if message.Text[:1] == "!" {
			commandHandler.HandleCommand(message.Text[1:], channel, user, message)
		}
	})

	client.OnConnect(func() {
		fmt.Println("Connected to the channel, bot online")
	})

	client.Join(twitchChannel)
	if err := client.Connect(); err != nil {
		panic(err)
	}
}
