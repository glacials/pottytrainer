package server

import (
	"fmt"
	"os"

	"github.com/Timothylock/go-signin-with-apple/apple"
)

const (
	// keyAppleClientID is the environment variable name for the Apple client ID.
	// This is typically in the reverse-domain-name format (e.g.
	// dev.twos.pottytrainer) and can be found on
	// https://developer.apple.com/account/resources/identifiers/list.
	keyAppleClientID = "APPLE_CLIENT_ID"
	// keyAppleKeyID is the key for environment variable that holds the key ID of
	// the Apple private key. This can be seen by viewing the key from
	// https://developer.apple.com/account/resources/authkeys/list.
	keyAppleKeyID = "APPLE_SIGNING_KEY_ID"
	// keyAppleSigningKeyFile is the key for the environment variable that
	// contains the path to the Apple signing key file. This file is generated
	// from https://developer.apple.com/account/resources/authkeys/list.
	keyAppleSigningKeyFile = "APPLE_SIGNING_KEY_FILE"
	// keyAppleTeamID is the key for the environment variable that contains the
	// Apple team ID. This can be found on https://developer.apple.com/account.
	keyAppleTeamID = "APPLE_TEAM_ID"
)

type AppleClient struct {
	clientID     string
	clientSecret string
	client       *apple.Client
}

func NewAppleClient() (*AppleClient, error) {
	clientID := os.Getenv(keyAppleClientID)
	if clientID == "" {
		return nil, fmt.Errorf(
			"missing environment variable %s (from %s)",
			keyAppleClientID,
			"https://developer.apple.com/account/resources/identifiers/list",
		)
	}

	keyID := os.Getenv(keyAppleKeyID)
	if keyID == "" {
		return nil, fmt.Errorf(
			"missing environment variable %s (from %s)",
			keyAppleKeyID,
			"https://developer.apple.com/account/resources/authkeys/list",
		)
	}

	signingKeyFile := os.Getenv(keyAppleSigningKeyFile)
	if signingKeyFile == "" {
		return nil, fmt.Errorf(
			"missing environment variable %s (from %s)",
			keyAppleSigningKeyFile,
			"https://developer.apple.com/account/resources/authkeys/list",
		)
	}
	keyBytes, err := os.ReadFile(os.Getenv(keyAppleSigningKeyFile))
	if err != nil {
		return nil, err
	}
	key := string(keyBytes)

	teamID := os.Getenv(keyAppleTeamID)
	if teamID == "" {
		return nil, fmt.Errorf(
			"missing environment variable %s (from %s)",
			keyAppleTeamID,
			"https://developer.apple.com/account",
		)
	}

	secret, err := apple.GenerateClientSecret(key, teamID, clientID, keyID)
	if err != nil {
		return nil, err
	}

	return &AppleClient{
		clientID:     clientID,
		clientSecret: secret,
		client:       apple.New(),
	}, nil
}
