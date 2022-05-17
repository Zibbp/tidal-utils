package tidal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	authURL = "https://auth.tidal.com/v1/oauth2"
	// API Keys - https://github.com/yaronzz/Tidal-Media-Downloader/blob/bb5be5e5fba3a648cdda9c8b46c707682fb5472c/TIDALDL-PY/tidal_dl/apiKey.py
	clientId     = "7m7Ap0JC9j1cOM3n"
	clientSecret = "vRAdA108tlvkJpTsGZS8rGZ7xTlbJ0qaZ2K9saEzsgY="
)

type DeviceCode struct {
	DeviceCode              string `json:"deviceCode"`
	UserCode                string `json:"userCode"`
	VerificationURI         string `json:"verificationUri"`
	VerificationURIComplete string `json:"verificationUriComplete"`
	ExpiresIn               int    `json:"expiresIn"`
	Interval                int    `json:"interval"`
}

type LoginResponse struct {
	AuthLogin AuthLogin
	AuthError AuthError
}

type AuthLogin struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	User         User   `json:"user"`
}

type Refresh struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	User        User   `json:"user"`
}

type User struct {
	UserID       int64       `json:"userId"`
	Email        interface{} `json:"email"`
	CountryCode  string      `json:"countryCode"`
	FullName     interface{} `json:"fullName"`
	FirstName    interface{} `json:"firstName"`
	LastName     interface{} `json:"lastName"`
	Nickname     interface{} `json:"nickname"`
	Username     string      `json:"username"`
	Address      interface{} `json:"address"`
	City         interface{} `json:"city"`
	Postalcode   interface{} `json:"postalcode"`
	UsState      interface{} `json:"usState"`
	PhoneNumber  interface{} `json:"phoneNumber"`
	Birthday     interface{} `json:"birthday"`
	Gender       interface{} `json:"gender"`
	ImageID      interface{} `json:"imageId"`
	ChannelID    int64       `json:"channelId"`
	ParentID     int64       `json:"parentId"`
	AcceptedEULA bool        `json:"acceptedEULA"`
	Created      int64       `json:"created"`
	Updated      int64       `json:"updated"`
	FacebookUid  int64       `json:"facebookUid"`
	AppleUid     interface{} `json:"appleUid"`
	GoogleUid    interface{} `json:"googleUid"`
	NewUser      bool        `json:"newUser"`
}

type AuthError struct {
	Status           int64  `json:"status"`
	Error            string `json:"error"`
	SubStatus        int64  `json:"sub_status"`
	ErrorDescription string `json:"error_description"`
}

type Session struct {
	SessionID   string `json:"sessionId"`
	UserID      int64  `json:"userId"`
	CountryCode string `json:"countryCode"`
	ChannelID   int64  `json:"channelId"`
	PartnerID   int64  `json:"partnerId"`
	Client      Client `json:"client"`
}

type Client struct {
	ID                       int64       `json:"id"`
	Name                     string      `json:"name"`
	AuthorizedForOffline     bool        `json:"authorizedForOffline"`
	AuthorizedForOfflineDate interface{} `json:"authorizedForOfflineDate"`
}

type Service struct {
	UserID       string
	AccessToken  string
	RefreshToken string
	clientId     string
	clientSecret string
}

func NewService() *Service {
	log.Debugf("Creating new Tidal Auth Service")
	var service Service
	service.clientId = clientId
	service.clientSecret = clientSecret
	// Check if access token is in config
	configAccessToken := viper.GetString("tidal.access_token")
	configRefreshToken := viper.GetString("tidal.refresh_token")
	if configAccessToken != "" {
		// Check if access token is valid
		session, err := checkSession(configAccessToken)
		if err != nil {
			log.Info("Failed to get session, attempting to get new access token")
			// Get session failed - access token is probably invalid, let's refresh it
			refresh, err := refreshAccessToken(configRefreshToken)
			if err != nil {
				log.Errorf("Failed to refresh access token: %s", err)
				// Failed to refresh access token - User will need to re-authenticate
				viper.Set("tidal.access_token", "")
				viper.Set("tidal.refresh_token", "")
				viper.WriteConfig()
				NewService()
			}
			// Write new access token to config
			viper.Set("tidal.access_token", refresh.AccessToken)
			viper.WriteConfig()
			service.AccessToken = refresh.AccessToken
			service.RefreshToken = configRefreshToken
			service.UserID = strconv.Itoa(int(refresh.User.UserID))
			return &service
		}
		log.Info("Tidal access token is valid.")
		// Access token is valid
		service.AccessToken = configAccessToken
		service.RefreshToken = configRefreshToken
		service.UserID = strconv.Itoa(int(session.UserID))
		return &service
	} else {
		log.Info("No access token found in config.")
		// Get device code for token login
		deviceCode, err := getDeviceCode()
		if err != nil {
			log.Panicf("Error getting device code: %w", err)
		}
		fmt.Printf("Please visit %s and authorize your Tidal account.\n", deviceCode.VerificationURIComplete)
		// Start polling for sucessfull oauth login
		for {
			loginResponse, err := tokenLogin(deviceCode)
			if err != nil {
				log.Panicf("Error logging in: %w", err)
			}
			if (AuthLogin{} == loginResponse.AuthLogin) {
				// No auth token - check what errors occured
				// If error is expired_token, the device ID expired (5 minutes)
				if loginResponse.AuthError.Error == "expired_token" {
					log.Fatal("Device token expired (expires in 5 minutes). Please try again.")
				}
			} else {
				log.Info("Successfully logged in.")
				service.UserID = strconv.Itoa(int(loginResponse.AuthLogin.User.UserID))
				service.AccessToken = loginResponse.AuthLogin.AccessToken
				service.RefreshToken = loginResponse.AuthLogin.RefreshToken
				viper.Set("tidal.access_token", service.AccessToken)
				viper.Set("tidal.refresh_token", service.RefreshToken)
				viper.Set("tidal.user_id", service.UserID)
				viper.WriteConfig()
				// Set
				break
			}
			d := time.Duration(deviceCode.Interval) * time.Second
			log.Debugf("Waiting %d seconds before trying again.", deviceCode.Interval)
			time.Sleep(d)
		}
	}
	return &service
}

func getDeviceCode() (DeviceCode, error) {
	var deviceCode DeviceCode

	client := &http.Client{}

	// Set body
	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("scope", "r_usr+w_usr+w_sub")

	encodedData := data.Encode()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/device_authorization", authURL), strings.NewReader(encodedData))
	if err != nil {
		return DeviceCode{}, err
	}

	// Set Headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return DeviceCode{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return DeviceCode{}, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Get Device Code failed with code: %v", resp.StatusCode)
		log.Error(string(body))
		return DeviceCode{}, err
	}

	err = json.Unmarshal(body, &deviceCode)
	if err != nil {
		return DeviceCode{}, err
	}

	return deviceCode, nil

}

func tokenLogin(deviceCode DeviceCode) (LoginResponse, error) {
	var loginResponse LoginResponse

	client := &http.Client{}
	// Set body
	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("device_code", deviceCode.DeviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	data.Set("scope", "r_usr+w_usr+w_sub")

	encodedData := data.Encode()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/token", authURL), strings.NewReader(encodedData))
	if err != nil {
		return LoginResponse{}, err
	}

	req.SetBasicAuth(clientId, clientSecret)

	// Set Headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return LoginResponse{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return LoginResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		var authError AuthError
		err = json.Unmarshal(body, &authError)
		if err != nil {
			return LoginResponse{}, err
		}
		loginResponse.AuthError = authError
		return loginResponse, err
	}

	var authLogin AuthLogin
	err = json.Unmarshal(body, &authLogin)
	if err != nil {
		return LoginResponse{}, err
	}
	loginResponse.AuthLogin = authLogin
	return loginResponse, nil

}

func checkSession(accessToken string) (Session, error) {
	var session Session

	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/sessions", apiURL), nil)
	if err != nil {
		return Session{}, err
	}

	// Set Headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := client.Do(req)
	if err != nil {
		return Session{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Session{}, err
	}

	if resp.StatusCode != http.StatusOK {
		err := errors.New("Error getting session")
		log.Debugf("Get Session failed with code: %v", resp.StatusCode)
		return Session{}, err
	}

	err = json.Unmarshal(body, &session)
	if err != nil {
		return Session{}, err
	}

	return session, nil
}

func refreshAccessToken(refreshToken string) (Refresh, error) {
	var refresh Refresh

	client := &http.Client{}
	// Set body
	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")
	data.Set("scope", "r_usr+w_usr+w_sub")

	encodedData := data.Encode()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/token", authURL), strings.NewReader(encodedData))
	if err != nil {
		return Refresh{}, err
	}

	req.SetBasicAuth(clientId, clientSecret)

	// Set Headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return Refresh{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Refresh{}, err
	}

	if resp.StatusCode != http.StatusOK {
		err := errors.New("Error refreshing access token")
		return Refresh{}, err
	}

	err = json.Unmarshal(body, &refresh)
	if err != nil {
		return Refresh{}, err
	}
	return refresh, nil
}
