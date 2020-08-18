package tidalapi

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// const (
// 	LOGIN = iota
// 	TRACK
// 	FAVORITETRACKS
// 	TRACKURL
// 	TRACKRADIO
// 	ARTIST
// 	ARTISTTRACKS
// 	ARTISTRADIO
// 	ALBUM
// 	ALBUMTRACKS
// 	PLAYLIST
// 	PLAYLISTTRACKS
// )

const (
	LOGIN           = "login/username"
	FAVORITETRACKS  = "users/%v/favorites/tracks"
	TRACK           = "tracks/%v"
	TRACKURL        = "tracks/%v/streamUrl"
	TRACKRADIO      = "tracks/%v/radio"
	ARTIST          = "artists/%v"
	ARTISTTOPTRACKS = "artists/%v/toptracks"
	ARTISTRADIO     = "artists/%v/radio"
	ALBUM           = "albums/%v"
	ALBUMTRACKS     = "albums/%v/tracks"
	PLAYLIST        = "playlists/%v"
	PLAYLISTTRACKS  = "playlists/%v/tracks"
	USER            = "users/%v"
	SEARCH          = "search?query=%v"
	SESSIONS        = "sessions"
)

const (
	MASTER = iota
	LOSSLESS
	HIGH
	LOW
)

const (
	API_TOKEN = "wc8j_yBJd20zOmx0"
	CLIENT_ID = "ck3zaWMi8Ka_XdI0"
)

var Quality = map[int]string{MASTER: "HI_RES", LOSSLESS: "LOSSLESS", HIGH: "HIGH", LOW: "LOW"}

var Types = map[string]interface{}{
	LOGIN: new(Login),
}

var tr = &http.Transport{
	MaxIdleConns:          10,
	IdleConnTimeout:       0 * time.Second,
	TLSHandshakeTimeout:   0 * time.Second,
	ResponseHeaderTimeout: 0 * time.Second,
	ExpectContinueTimeout: 0 * time.Second,
	DisableCompression:    true,
}

// var API = map[int]string{
// 	LOGIN:          "login/username",
// 	FAVORITETRACKS: "users/%v/favorites/tracks",
// 	TRACK:          "tracks/%v",
// 	TRACKURL:       "tracks/%v/streamUrl",
// 	TRACKRADIO:     "tracks/%v/radio",
// 	ARTIST:         "artists/%v",
// 	ARTISTTRACKS:   "artists/%v/favorites/tracks",
// 	ARTISTRADIO:    "artists/%v/radio",
// 	ALBUM:          "albums/%v",
// 	ALBUMTRACKS:    "albums/%v/tracks",
// 	PLAYLIST:       "playlists/%v",
// 	PLAYLISTTRACKS: "playlists/%v/tracks",
// }

// Config stores TIDAL client configuration
type Config struct {
	quality     string
	apiLocation *url.URL
	apiToken    string
	values      url.Values
}

func (c *Config) Init(quality int) {
	c.quality = Quality[quality]
	c.apiLocation, _ = url.Parse("https://api.tidal.com/v1/")
	v := url.Values{}
	v.Add("limit", "999")
	c.values = v
}

func NewSession(quality int) *Session {
	var s Session
	var c Config
	c.Init(quality)
	s.Quality = Quality[quality]
	s.configuration = &c
	s.client = &http.Client{Timeout: 0 * time.Second, Transport: tr}

	cc := make([]byte, 32)
	rand.Read(cc)
	s.CodeVerifier = base64.RawURLEncoding.EncodeToString(cc)

	cs := sha256.Sum256([]byte(s.CodeVerifier))
	s.CodeChallenge = base64.RawURLEncoding.EncodeToString(cs[:])

	s.generateClientUniqueKey()

	s.conf = &oauth2.Config{
		ClientID:     CLIENT_ID,
		ClientSecret: s.ClientUniqueKey,
		Scopes:       []string{"r_usr", "w_usr", "w_sub"},
		RedirectURL:  "https://tidal.com/android/login/auth",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.tidal.com/authorize",
			TokenURL: "https://auth.tidal.com/v1/oauth2/token",
		},
	}
	return &s
}

func (s *Session) LoadSession(fn string) (err error) {
	outBytes, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}
	err = json.Unmarshal(outBytes, s)
	if err != nil {
		return err
	}
	if s.Quality != s.configuration.quality {
		return errors.New("Cannot load session made for the different quality level")
	}
	s.configuration.values.Add("countryCode", s.CountryCode)
	if s.Token != nil {
		s.conf = &oauth2.Config{
			ClientID:     CLIENT_ID,
			ClientSecret: s.ClientUniqueKey,
			Scopes:       []string{"r_usr", "w_usr", "w_sub"},
			RedirectURL:  "https://tidal.com/android/login/auth",
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://login.tidal.com/authorize",
				TokenURL: "https://auth.tidal.com/v1/oauth2/token",
			},
		}
		s.client = s.conf.Client(context.Background(), s.Token)
	}
	return nil
}

func (s *Session) SaveSession(fn string) (err error) {
	outBytes, err := json.Marshal(s)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fn, outBytes, 0600)
	return err
}

func (s *Session) IsValid() bool {
	usr := new(User)
	err := s.Get(USER, s.UserID, usr)
	if err != nil {
		return false
	}
	return true
}

func (s *Session) generateClientUniqueKey() {
	cc := make([]byte, 16)
	rand.Read(cc)
	s.ClientUniqueKey = fmt.Sprintf("%02x", cc)
}

func (s *Session) Login(username, password string) error {

	s.configuration.apiToken = API_TOKEN

	params := url.Values{}
	data := url.Values{}
	data.Add("clientUniqueKey", s.ClientUniqueKey)
	data.Add("username", username)
	data.Add("password", password)

	l := new(Login)

	err := s.request("POST", LOGIN, params, data, l)
	if err != nil {
		return err
	}
	s.SessionID = l.SessionId

	err = s.Get(SESSIONS, nil, s)
	if err != nil {
		return err
	}

	s.configuration.values.Add("countryCode", s.CountryCode)

	return nil
}

func (s *Session) GetOauth2URL() (authURL string) {

	authURL = s.conf.AuthCodeURL("state",
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("appMode", "android"),
		oauth2.SetAuthURLParam("code_challenge", s.CodeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("response_type", "code"),
		oauth2.SetAuthURLParam("client_unique_key", s.ClientUniqueKey),
		oauth2.SetAuthURLParam("lang", "en_US"))

	return authURL
}

func (s *Session) LoginWithOauth2Code(code string) (err error) {

	ctx := context.Background()

	s.Token, err = s.conf.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", s.CodeVerifier),
		oauth2.SetAuthURLParam("client_id", CLIENT_ID),
		oauth2.SetAuthURLParam("client_unique_key", s.ClientUniqueKey))
	if err != nil {
		return err
	}

	s.client = s.conf.Client(ctx, s.Token)

	err = s.Get(SESSIONS, nil, s)
	if err != nil {
		return err
	}
	s.configuration.values.Add("countryCode", s.CountryCode)

	return nil
}

func (s *Session) DownloadImage(id string) ([]byte, error) {
	var err error
	id = strings.Replace(id, "-", "/", -1)
	url := fmt.Sprintf(IMGPATH, id, 750, 750)
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return body, nil
}

func (s *Session) Get(what string, id interface{}, obj interface{}) error {
	apiPath := what
	if id != nil {
		apiPath = fmt.Sprintf(what, url.QueryEscape(fmt.Sprintf("%v", id)))
	}
	params := url.Values{}
	data := url.Values{}
	params.Add("soundQuality", s.configuration.quality)
	err := s.request("GET", apiPath, params, data, obj)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) request(method, uri string, params, data url.Values, response interface{}) error {

	refURI, err := url.Parse(uri)
	if err != nil {
		return err
	}
	reqURL := s.configuration.apiLocation.ResolveReference(refURI)

	form := url.Values{}
	for k, v := range s.configuration.values {
		form.Add(k, v[0])
	}
	for k, v := range params {
		form.Add(k, v[0])
	}

	reqURL.RawQuery = form.Encode()

	req, err := http.NewRequest(method, reqURL.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	if s.configuration.apiToken != "" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("X-Tidal-Token", s.configuration.apiToken)
		if s.SessionID != "" {
			req.Header.Add("X-Tidal-SessionId", s.SessionID)
		}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// log.Println(resp)
	// log.Println(string(body))

	err = ToMap(body, response)
	if resp.StatusCode >= 400 {
		return errors.New(resp.Status)
	}
	// log.Println(response)
	return err
}

func ToMap(obj []byte, data interface{}) error {
	err := json.Unmarshal(obj, data)
	if err != nil {
		return err
	}
	return nil
}
