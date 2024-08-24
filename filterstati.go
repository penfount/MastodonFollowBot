/// (c) Bernhard Tittelbach, 2019 - MIT License
package main

import "github.com/McKael/madon/v3"

func goFilterStati(client *madon.Client, statusIn <-chan madon.Notification, statusOut chan<- madon.Notification) {
	defer close(statusOut)
FILTERFOR:
	for status := range statusIn {
		// Skip all that aren't follow notifications
		if status.Type != "follow" {
			continue FILTERFOR
		}

		// check if we already follow this person, and skip if so
		already_followed := false
		if relationship, relerr := getRelation(client, status.Account.ID); relerr == nil {
			already_followed = relationship.Following && !relationship.Blocking
		} else {
			LogMadon_.Println("goFilterStati::FollowCheck:", relerr)
			continue FILTERFOR
		}
		if already_followed {
			LogMadon_.Printf("goFilterStati: Already following %s, skipping\n", status.Account.Acct)
			continue FILTERFOR
		}

		statusOut <- status
	}
}
