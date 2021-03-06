package notify

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/infracloudio/botkube/pkg/config"
	"github.com/infracloudio/botkube/pkg/events"
	log "github.com/infracloudio/botkube/pkg/logging"
	"github.com/nlopes/slack"
)

var attachmentColor map[events.Level]string

// Slack contains Token for authentication with slack and Channel name to send notification to
type Slack struct {
	Token       string
	Channel     string
	ClusterName string
}

// NewSlack returns new Slack object
func NewSlack() Notifier {
	attachmentColor = map[events.Level]string{
		events.Info:     "good",
		events.Warn:     "warning",
		events.Debug:    "good",
		events.Error:    "danger",
		events.Critical: "danger",
	}

	c, err := config.New()
	if err != nil {
		log.Logger.Fatal(fmt.Sprintf("Error in loading configuration. Error:%s", err.Error()))
	}

	return &Slack{
		Token:       c.Communications.Slack.Token,
		Channel:     c.Communications.Slack.Channel,
		ClusterName: c.Settings.ClusterName,
	}
}

// SendEvent sends event notification to slack
func (s *Slack) SendEvent(event events.Event) error {
	log.Logger.Info(fmt.Sprintf(">> Sending to slack: %+v", event))

	api := slack.New(s.Token)
	params := slack.PostMessageParameters{
		AsUser: true,
	}
	attachment := slack.Attachment{
		Fields: []slack.AttachmentField{
			{
				Title: "Kind",
				Value: event.Kind,
				Short: true,
			},
			{

				Title: "Name",
				Value: event.Name,
				Short: true,
			},
		},
		Footer: "BotKube",
	}

	// Add timestamp
	ts := json.Number(strconv.FormatInt(event.TimeStamp.Unix(), 10))
	if ts > "0" {
		attachment.Ts = ts
	}

	if event.Namespace != "" {
		attachment.Fields = append(attachment.Fields, slack.AttachmentField{
			Title: "Namespace",
			Value: event.Namespace,
			Short: true,
		})
	}

	if event.Reason != "" {
		attachment.Fields = append(attachment.Fields, slack.AttachmentField{
			Title: "Reason",
			Value: event.Reason,
			Short: true,
		})
	}

	if len(event.Messages) > 0 {
		message := ""
		for _, m := range event.Messages {
			message = message + m
		}
		attachment.Fields = append(attachment.Fields, slack.AttachmentField{
			Title: "Message",
			Value: message,
		})
	}

	if event.Action != "" {
		attachment.Fields = append(attachment.Fields, slack.AttachmentField{
			Title: "Action",
			Value: event.Action,
		})
	}

	if len(event.Recommendations) > 0 {
		rec := ""
		for _, r := range event.Recommendations {
			rec = rec + r
		}
		attachment.Fields = append(attachment.Fields, slack.AttachmentField{
			Title: "Recommendations",
			Value: rec,
		})
	}

	// Add clustername in the message
	attachment.Fields = append(attachment.Fields, slack.AttachmentField{
		Title: "Cluster",
		Value: s.ClusterName,
	})

	attachment.Color = attachmentColor[event.Level]
	params.Attachments = []slack.Attachment{attachment}

	channelID, timestamp, err := api.PostMessage(s.Channel, "", params)
	if err != nil {
		log.Logger.Errorf("Error in sending slack message %s", err.Error())
		return err
	}

	log.Logger.Infof("Message successfully sent to channel %s at %s", channelID, timestamp)
	return nil
}

// SendMessage sends message to slack channel
func (s *Slack) SendMessage(msg string) error {
	log.Logger.Info(fmt.Sprintf(">> Sending to slack: %+v", msg))

	api := slack.New(s.Token)
	params := slack.PostMessageParameters{
		AsUser: true,
	}

	channelID, timestamp, err := api.PostMessage(s.Channel, msg, params)
	if err != nil {
		log.Logger.Errorf("Error in sending slack message %s", err.Error())
		return err
	}

	log.Logger.Infof("Message successfully sent to channel %s at %s", channelID, timestamp)
	return nil
}
