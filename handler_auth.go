package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nlopes/slack"
)

const slackBaseURL = "https://slack.com"

var (
	slackClientID     = os.Getenv("SLACK_CLIENT_ID")
	slackClientSecret = os.Getenv("SLACK_CLIENT_SECRET")

	slackAPIScopes = []string{
		"team:read",
		"users:read",
		"usergroups:read",
		"im:write",
		"chat:write:user",
		"chat:write:bot",
	}
)

func handleAuthInitiate(c *gin.Context) {
	redirectURI, err := authorizeURI(relativeURI(c, "/auth/complete"))
	if err != nil {
		log.Fatal(err)
	}
	c.Redirect(http.StatusSeeOther, redirectURI)
}

func handleAuthComplete(c *gin.Context) {
	code := c.Query("code")

	response, err := slack.GetOAuthResponse(slackClientID, slackClientSecret, code, relativeURI(c, "/auth/complete"), false)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	setAuthorizedToken(c, response.AccessToken)

	c.Redirect(http.StatusSeeOther, relativeURI(c, "/"))
}

func handleAuthLogout(c *gin.Context) {
	setAuthorizedToken(c, "")

	c.Redirect(http.StatusSeeOther, relativeURI(c, "/"))
}

func authorizeURI(redirectURI string) (string, error) {
	redirectURL, err := url.Parse(slackBaseURL + "/oauth/authorize")
	if err != nil {
		return "", err
	}
	q := redirectURL.Query()
	q.Set("client_id", slackClientID)
	q.Set("scope", strings.Join(slackAPIScopes, ","))
	q.Set("redirect_uri", redirectURI)
	redirectURL.RawQuery = q.Encode()

	return redirectURL.String(), nil
}

func setAuthorizedToken(c *gin.Context, token string) {
	c.SetCookie(cookiePrefix+"slacktoken", token, 86400, "", "", true, true)
}

func isAuthorized(c *gin.Context) bool {
	token, err := c.Cookie(cookiePrefix + "slacktoken")
	return token != "" && err == nil
}

func authorizedToken(c *gin.Context) string {
	token, _ := c.Cookie(cookiePrefix + "slacktoken")
	return token
}

func slackAPI(c *gin.Context) (*slack.Client, error) {
	token := authorizedToken(c)
	if token == "" {
		return nil, fmt.Errorf("access token not found")
	}
	return slack.New(token), nil
}
