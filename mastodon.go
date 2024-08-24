/// (c) Bernhard Tittelbach, 2019 - MIT License
package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/McKael/madon/v3"
	"github.com/spf13/viper"
)

/// code adapted from madonctl by McKael -- https://github.com/McKael/madonctl
/// Kudos!!
func madonMustInitClient() (client *madon.Client, err error) {

	appName := viper.GetString("app_name")
	instanceURL := viper.GetString("instance")
	appKey := viper.GetString("app_key")
	appSecret := viper.GetString("app_secret")
	appToken := viper.GetString("app_token")
	appScopes := viper.GetStringSlice("app_scopes")

	if instanceURL == "" {
		LogMadon_.Fatalln("madonInitClient:", "no instance provided")
	}

	LogMadon_.Println("madonInitClient:Instance: ", instanceURL)

	if appKey != "" && appSecret != "" {
		// We already have an app key/secret pair
		client, err = madon.RestoreApp(appName, instanceURL, appKey, appSecret, nil)
		client.SetUserToken(appToken, "", "", appScopes)
		if err != nil {
			return
		}
		// Check instance
		if _, err = client.GetCurrentInstance(); err != nil {
			LogMadon_.Fatalln("madonInitClient:", err, "could not connect to server with provided app ID/secret")
			return
		}
		LogMadon_.Println("madonInitClient:", "Using provided app ID/secret")
		return
	}

	if appKey != "" || appSecret != "" {
		LogMadon_.Fatalln("madonInitClient:", "Warning: provided app id/secrets incomplete -- registering again")
	}

	LogMadon_.Println("madonInitClient:", "Registered new application.")
	return
}

/// code adapted from madonctl by McKael -- https://github.com/McKael/madonctl
/// Kudos!!
func goSubscribeStreamOfTagNames(client *madon.Client, statusOutChan chan<- madon.Notification) {
	streamName := "user"
	evChan := make(chan madon.StreamEvent, 10)
	stop := make(chan bool)
	done := make(chan bool)

	err := client.StreamListener(streamName, "", evChan, stop, done)

	if err != nil {
		LogMadon_.Fatalln("goSubscribeStreamOfTagNames:", err.Error())
	}

LISTENSTREAM:
	for {
		select {
		case v, ok := <-done:
			if !ok || v { // done is closed, end of streaming
				break LISTENSTREAM
			}
		case ev := <-evChan:
			switch ev.Event {
			case "error":
				if ev.Error != nil {
					if ev.Error == io.ErrUnexpectedEOF {
						LogMadon_.Println("goSubscribeStreamOfTagNames:", "The stream connection was unexpectedly closed")
						continue
					}
					LogMadon_.Printf("goSubscribeStreamOfTagNames: Error event: [%s] %s\n", ev.Event, ev.Error)
					// bail if the error starts with "read error"
					if strings.HasPrefix(ev.Error.Error(), "read error") {
						os.Exit(1)
					}
					continue
				}
				LogMadon_.Printf("goSubscribeStreamOfTagNames: Event: [%s]\n", ev.Event)
			case "notification":
				s := ev.Data.(madon.Notification)
				statusOutChan <- s
			case "update", "delete", "status.update":
				continue
			default:
				LogMadon_.Printf("goSubscribeStreamOfTagNames: Unhandled event: [%s] %T\n", ev.Event, ev.Data)
			}
		}
	}
	close(stop)
	close(evChan)
	close(statusOutChan)
	if err != nil {
		LogMadon_.Printf("goSubscribeStreamOfTagNames: Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func getRelation(client *madon.Client, accID string) (madon.Relationship, error) {
	relationshiplist, err := client.GetAccountRelationships([]string{accID})
	if err != nil {
		return madon.Relationship{}, err
	}
	if len(relationshiplist) == 0 {
		return madon.Relationship{}, fmt.Errorf("AccountID not known, got empty result")
	}
	return relationshiplist[0], nil
}

func goFollowBack(client *madon.Client, notification_chan <-chan madon.Notification) {
	for notification := range notification_chan {
		relationship, err := client.FollowAccount(notification.Account.ID, nil)
		LogMadon_.Printf("goFollowBack: Following account %s (id %s), following %t, followed_by %t\n", notification.Account.Acct, relationship.ID, relationship.Following, relationship.FollowedBy)
		if err != nil {
			LogMadon_.Printf("goFollowBack: Error: %s\n", err.Error())
		}
	}
}
