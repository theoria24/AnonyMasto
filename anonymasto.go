package main

import (
	"context"
	"fmt"
	"html"
	"log"
	"os"
	"regexp"

	"github.com/joho/godotenv"
	"github.com/mattn/go-mastodon"
)

func normalizeText(str string) string {
	re := regexp.MustCompile("<br( //)?>")
	str = re.ReplaceAllString(str, "\n")
	re = regexp.MustCompile("</p>\n*<p>")
	str = re.ReplaceAllString(str, "\n\n")
	re = regexp.MustCompile(`<("[^"]*"|'[^']*'|[^'">])*>`)
	str = re.ReplaceAllString(str, "")
	re = regexp.MustCompile("^@AnonyMasto( |\n)")
	str = re.ReplaceAllString(str, "")
	str = html.UnescapeString(str)
	return str
}

func main() {
	err := godotenv.Load()

	c := mastodon.NewClient(&mastodon.Config{
		Server:       os.Getenv("MSTDN_SERVER"),
		ClientID:     os.Getenv("MSTDN_CLIENT_ID"),
		ClientSecret: os.Getenv("MSTDN_CLIENT_SECRET"),
		AccessToken:  os.Getenv("MSTDN_ACCESS_TOKEN"),
	})

	wsc := c.NewWSClient()
	q, err := wsc.StreamingWSUser(context.Background())
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Start Watching")
	}
	for e := range q {
		if t, ok := e.(*mastodon.NotificationEvent); ok {
			if t.Notification.Type == "mention" {
				if t.Notification.Status.Visibility == "direct" {
					// fmt.Printf("%#v\n", t.Notification.Status)
					fmt.Println(t.Notification.Status.Account.Acct + ": 【" + t.Notification.Status.SpoilerText + "】" + normalizeText(t.Notification.Status.Content))
					_, err = c.PostStatus(context.Background(), &mastodon.Toot{
						SpoilerText: t.Notification.Status.SpoilerText,
						Status:      normalizeText(t.Notification.Status.Content),
					})
					if err != nil {
						fmt.Println(err)
					}
				} else if t.Notification.Status.Visibility == "public" {
					fmt.Println("[BT]:" + t.Notification.Status.Account.Acct + ": 【" + t.Notification.Status.SpoilerText + "】" + normalizeText(t.Notification.Status.Content))
					_, err = c.Reblog(context.Background(), t.Notification.Status.ID)
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}
}
