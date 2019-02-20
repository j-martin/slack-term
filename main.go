package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/j-martin/slag/components"
	"github.com/j-martin/slag/secrets"
	"github.com/j-martin/slag/service"
	"github.com/nlopes/slack"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	VERSION = "v0.1.0"
	USAGE   = `NAME:
		slag - slack channel aggregator for your terminal

USAGE:
		slag DOMAIN

VERSION:
		%s

WEBSITE:
		https://github.com/j-martin/slag

ARGUMENTS
	 DOMAIN   Domain/workspace to use. 

GLOBAL OPTIONS:
	 -f [REGEX]        Regex to filter channels. Default: '.*'
	 -n [INT]          Number of previous message to display per channel.
	 -reset-token      Reset the API token for the domain.
	 -help, -h
`
)

var (
	flagRegexFilter       string
	flagResetToken        bool
	flagMessageFetchCount int
)

func init() {
	// Parse flags
	flag.StringVar(
		&flagRegexFilter,
		"f",
		".*",
		"Regex filter for channels.",
	)

	flag.IntVar(
		&flagMessageFetchCount,
		"n",
		20,
		"Number of historical messages to fetch, per channels.",
	)

	flag.BoolVar(
		&flagResetToken,
		"reset-token",
		false,
		"Reset the API token for the domain.",
	)
	flag.Usage = func() {
		fmt.Printf(USAGE, VERSION)
	}

	flag.Parse()
	if flag.Arg(0) == "" || len(flag.Args()) != 1 {
		flag.Usage()
		log.Fatal("The domain must be passed as an argument.")
	}
}

func main() {
	var apiToken string
	err := secrets.New("slack").LoadCredentialItem(
		flag.Arg(0),
		&apiToken,
		"Generate the api token at: https://api.slack.com/custom-integrations/legacy-tokens",
		flagResetToken)
	if err != nil {
		log.Fatal(err)
	}
	svc, err := service.NewSlackService(apiToken)
	if err != nil {
		log.Fatal(err)
	}
	channels, err := svc.GetChannels()
	messagesCh := make(chan []components.Message)
	r, _ := regexp.Compile(flagRegexFilter)
	watchedChannels := make(map[string]*components.Channel)
	watchedChannelNames := make([]string, 0)
	for _, channel := range channels {
		if !r.MatchString(channel.Name) {
			continue
		}
		ch := channel
		watchedChannels[channel.ID] = &ch
		watchedChannelNames = append(watchedChannelNames, ch.Name)
		if flagMessageFetchCount == 0 {
			continue
		}
		go func() {
			channelMessages, err := svc.GetMessages(ch, flagMessageFetchCount)
			if err != nil {
				log.Fatal(err)
			}
			messagesCh <- channelMessages
		}()
	}
	messages := make([]components.Message, 0)
	if flagMessageFetchCount != 0 {
		log.Printf("Fetching: %s ...", strings.Join(watchedChannelNames, ", "))

		for i := 0; i < len(watchedChannels); i++ {
			messages = append(messages, <-messagesCh...)
		}
	}

	close(messagesCh)

	if len(watchedChannels) == 0 {
		log.Fatalf("No channels matched the regex filter: '%s'", flagRegexFilter)
	}

	sort.Sort(sort.Reverse(components.Messages(messages)))

	for _, message := range messages {
		printMessage(message, svc.CurrentTeamInfo)
	}
	if flagMessageFetchCount == 0 {
		log.Printf("Listening to %s for new messages ...", strings.Join(watchedChannelNames, ", "))
	}
	err = svc.ListenToEvents(watchedChannels, printMessage)
	if err != nil {
		log.Fatal(err)
	}
}

func printMessage(message components.Message, teamInfo *slack.TeamInfo) {
	threadSymbol := ""
	if message.IsReply {
		threadSymbol = "≡"
	}
	fmt.Println(
		color.MagentaString("%s [%s]", message.Time.UTC().Format(time.RFC3339), message.Time.Format("15:04:05Z07:00")),
		color.New().Add(color.Faint).Sprintf("https://%s.slack.com/messages/%s/convo/%s-%s/", teamInfo.Domain, message.Channel.ID, message.Channel.ID, message.ThreadTimestamp),
		threadSymbol,
	)
	fmt.Printf("%s %s ",
		color.CyanString("[#%s]", message.Channel.Name),
		color.RedString("@%s", message.Name),
	)
	if len(message.Content) > 0 {
		fmt.Println(message.Content)
	}
	if len(message.Attachments) > 0 {
		print(color.New().Add(color.Faint).Sprintf("\n%s\n", strings.Join(message.Attachments, "\n")))
	}
	fmt.Println()
}
