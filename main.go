package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
// 	"strings"
	"text/template"

    "github.com/sj14/review-bot/common"
	"github.com/sj14/review-bot/hoster/github"
	"github.com/sj14/review-bot/hoster/gitlab"
	"github.com/sj14/review-bot/slackermost"
)

func main() {
	var (
		host                 = flag.String("host", "", "host address (e.g. github.com, gitlab.com or self-hosted gitlab url)")
		token                = flag.String("token", "", "host API token")
		repo                 = flag.String("repo", "", "repository (format: 'owner/repo'), or project id (only gitlab)")
		reviewersPath        = flag.String("reviewers", "examples/reviewers.json", "path to the reviewers file")
		templatePath         = flag.String("template", "", "path to the template file")
		webhook              = flag.String("webhook", "", "slack/mattermost webhook URL")
		webhookAuthorization = flag.String("webhook-auth", "", "webhook authorization header")
		channelOrUser        = flag.String("channel", "", "mattermost channel (e.g. MyChannel) or user (e.g. @AnyUser)")
	)
	flag.Parse()

	if *host == "" {
		log.Fatalln("missing host")
	}
	if *repo == "" {
		log.Fatalln("missing repository")
	}

	reviewers := loadReviewers(*reviewersPath)

	var tmpl *template.Template
	if *templatePath != "" {
		tmpl = loadTemplate(*templatePath)
	} else if *host == "github.com" {
		tmpl = github.DefaultTemplate()
	} else {
		tmpl = gitlab.DefaultTemplate()
	}

	var reminder string
	if *host == "github.com" {
	    panic("not implemented")
// 		ownerRespo := strings.SplitN(*repo, "/", 2)
// 		if len(ownerRespo) != 2 {
// 			log.Fatalln("wrong repo format (use 'owner/repo')")
// 		}
// 		repo, reminders := github.AggregateReminder(*token, ownerRespo[0], ownerRespo[1], reviewers)
// 		if len(reminders) == 0 {
// 			// prevent from sending the header only
// 			return
// 		}
// 		reminder = github.ExecTemplate(tmpl, repo, reminders)

	} else {
		project, reminders := gitlab.AggregateReminder(*host, *token, *repo, reviewers)
		if len(reminders) == 0 {
			// prevent from sending the header only
			return
		}
		reminder = gitlab.ExecTemplate(tmpl, project, reminders)
	}

	if reminder == "" {
		return
	}

	fmt.Println(reminder)

	if *webhook != "" {
		if err := slackermost.Send(*channelOrUser, reminder, *webhook, *webhookAuthorization); err != nil {
			log.Fatalf("failed sending slackermost message: %v", err)
		}
	}
}

func loadTemplate(path string) *template.Template {
	t, err := template.ParseFiles(path)
	if err != nil {
		log.Fatalf("failed to read template file: %v", err)
	}
	return t
}

// load reviewers from given json file
// formatting:
// [{"username": "GitLabUserID", "discordId": "<@DiscordID>", "labels": [...]}, ...]
func loadReviewers(path string) map[string]common.Reviewer {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read reviewers file: %v", err)
	}

	reviewersFlat := []common.Reviewer{}
	if err := json.Unmarshal(b, &reviewersFlat); err != nil {
		log.Fatalf("failed to unmarshal reviewers: %v", err)
	}

    reviewers := map[string]common.Reviewer{}
    for _, r := range reviewersFlat {
        reviewers[r.Username] = r
    }

	return reviewers
}
