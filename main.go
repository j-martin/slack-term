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
	"os"
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
		slag

VERSION:
		%s

WEBSITE:
		https://github.com/j-martin/slag

GLOBAL OPTIONS:
	 -f [regex filter for channels]
	 -n [int]
	 -d [slack domain]
     -reset-token Reset the API token for the domain.
	 -help, -h
`
)

var (
	flagRegexFilter       string
	flagDomain            string
	flagResetToken        bool
	flagMessageFetchCount int
)

func init() {
	// Parse flags
	flag.StringVar(
		&flagRegexFilter,
		"f",
		"^random$",
		"Regex filter for channels.",
	)

	flag.IntVar(
		&flagMessageFetchCount,
		"n",
		20,
		"Number of historical messages to fetch, per channels.",
	)

	flag.StringVar(
		&flagDomain,
		"d",
		"",
		"Slack domain/workspace to use.",
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
	if flagDomain == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	var apiToken string
	err := secrets.New("slack").LoadCredentialItem(
		flagDomain,
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
	messages := make([]components.Message, 0)
	r, _ := regexp.Compile(flagRegexFilter)
	watchedChannels := make(map[string]*components.Channel)
	for _, channel := range channels {
		if !r.MatchString(channel.Name) {
			continue
		}
		log.Printf("Fetching: %s ...", channel.Name)
		ch := channel
		watchedChannels[channel.ID] = &ch
		channelMessages, err := svc.GetMessages(channel, flagMessageFetchCount)
		if err != nil {
			log.Fatal(err)
		}
		messages = append(messages, channelMessages...)
	}
	if len(watchedChannels) == 0 {
		log.Fatalf("No channels matched the regex filter: '%s'", flagRegexFilter)
	}

	sort.Sort(sort.Reverse(components.Messages(messages)))

	for _, message := range messages {
		printMessage(message, svc.CurrentTeamInfo)
	}
	err = svc.ListenToEvents(watchedChannels, printMessage)
	if err != nil {
		log.Fatal(err)
	}
}

func printMessage(message components.Message, teamInfo *slack.TeamInfo) {
	fmt.Println(
		color.MagentaString("%s [%s]", message.Time.UTC().Format(time.RFC3339), message.Time.Format("15:04:05Z07:00")),
		color.New().Add(color.Faint).Sprintf("https://%s.slack.com/messages/%s/convo/%s-%s/", teamInfo.Domain, message.Channel.ID, message.Channel.ID, message.ThreadTimestamp),
	)
	fmt.Printf("%s %s ",
		color.RedString("@%s", message.Name),
		color.CyanString("[#%s]", message.Channel.Name),
	)
	if len(message.Content) > 0 {
		fmt.Println(message.Content)
	}
	if len(message.Attachments) > 0 {
		print(color.New().Add(color.Faint).Sprintf("\n%s\n", strings.Join(message.Attachments, "\n")))
	}
	fmt.Println()
}
