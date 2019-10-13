package tidalapi

import (
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
)

const (
	MASTER = iota
	LOSSLESS
	HIGH
	LOW
)

const (
	TIDALAPITOKEN = "MbjR4DLXz1ghC4rV"
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
	var err error
	c.apiLocation, err = url.Parse("https://api.tidal.com/v1/")
	if err != nil {
		log.Fatal(err)
	}
	c.apiToken = TIDALAPITOKEN
	v := url.Values{}
	v.Add("limit", "999")
	c.values = v
}

type Session struct {
	ClientUniqueKey string
	SessionID       string
	CountryCode     string
	User            int
	Quality         string
	client          *http.Client
	configuration   *Config
}

func NewSession(quality int) *Session {
	var s Session
	var c Config
	c.Init(quality)
	s.Quality = Quality[quality]
	s.configuration = &c
	s.client = &http.Client{Timeout: 0 * time.Second, Transport: tr}
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

func (s *Session) generateClientUniqueKey() {
	num := rand.Int63()
	s.ClientUniqueKey = fmt.Sprintf("%02x", num)
}

func (s *Session) Login(username, password string) error {
	if s.ClientUniqueKey == "" {
		s.generateClientUniqueKey()
	}
	params := url.Values{}
	data := url.Values{}
	data.Add("clientUniqueKey", s.ClientUniqueKey)
	data.Add("username", username)
	data.Add("password", password)
	data.Add("User-Agent", "TIDAL_ANDROID/680 okhttp/3.3.1")
	data.Add("token", s.configuration.apiToken)
	data.Add("clientVersion", "1.12.2")

	l := new(Login)

	err := s.request("POST", LOGIN, params, data, l)
	if err != nil {
		return err
	}
	s.SessionID = l.SessionId
	s.CountryCode = l.CountryCode
	s.User = int(l.UserId)
	// log.Println(l)
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
	apiPath := fmt.Sprintf(what, id)
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

	req.Header.Add("User-Agent", "TIDAL_ANDROID/680 okhttp/3.3.1")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if s.SessionID != "" {
		req.Header.Add("X-Tidal-SessionId", s.SessionID)
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
