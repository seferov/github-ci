// Farhad Safarov <farhad.safarov@gmail.com>

package main

import "testing"

func TestGetCloneURL(t *testing.T) {
	configuration.Github.Username = "seferov"
	configuration.Github.Password = "passw0rd"
	configuration.Github.Organization = "egenis"
	configuration.Github.Repo = "website"

	if getCloneURL() != "https://seferov:passw0rd@github.com/egenis/website.git" {
		t.Fail()
	}
}
