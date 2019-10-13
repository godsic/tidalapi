package tidalapi

var (
	IMGPATH = "https://resources.tidal.com/images/%s/%dx%d.jpg"
)

type Login struct {
	SessionId   string  `json:"sessionId"`
	CountryCode string  `json:"countryCode"`
	UserId      float32 `json:"userId"`
}

type Error struct {
	userMessage string
}

type Artist struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Popularity int    `json:popularity`
	Url        string `json:url`
	Picture    string `json:picture`
}

type Album struct {
	Id             int    `json:"id"`
	Title          string `json:"title"`
	Cover          string `json:"cover"`
	NumberOfTracks int    `json:"numberOfTracks"`
	Duration       int    `json:"duration"`
	Artist         Artist `json:"artist"`
	ReleaseDate    string `json:"releaseDate"`
	Copyright      string `json:"copyright"`
	Upc            string `json:"ups"`
	explicit       bool   `json:"explicit"`
	Tracks         *Tracks
}

type TrackPath struct {
	Url                   string `json:"Url"`
	TrackID               int    `json:"trackId"`
	PlayTimeLeftInMinutes int    `json:"playTimeLeftInMinutes"`
	SoundQuality          string `json:"soundQuality"`
	EncryptionKey         string `json:"encryptionKey"`
	Codec                 string `json:"codec"`
}

type Track struct {
	Duration        int           `json:"duration"`
	ReplayGain      float32       `json:"replayGain"`
	Copyright       string        `json:"copyright"`
	Artists         []Artist      `json:"artists"`
	URL             string        `json:"url"`
	ISRC            string        `json:"isrc"`
	Editable        bool          `json:"editable"`
	SurroundTypes   []interface{} `json:"surroundTypes"`
	Artist          Artist        `json:"artist"`
	Explicit        bool          `json:"explicit"`
	AudioQuality    string        `json:"audioQuality"`
	ID              int           `json:"id"`
	Peak            float32       `json:"peak"`
	StreamReady     bool          `json:"streamReady"`
	StreamStartDate string        `json:"streamStartDate"`
	Popularity      int           `json:"popularity"`
	Album           Album         `json:"album"`
	Title           string        `json:"title"`
	AllowStreaming  bool          `json:"allowStreaming"`
	TrackNumber     int           `json:"trackNumber"`
	VolumeNumber    int           `json:"volumeNumber"`
	Version         string        `json:"version"`
	Path            TrackPath
}
type Item struct {
	Created string `json:"created"`
	Item    Track  `json:"item"`
}

type Tracks struct {
	Limit              int     `json:"limit"`
	Offset             int     `json:"offset"`
	TotalNumberOfItems int     `json:"totalNumberOfItems"`
	Items              []Track `json:"items"`
}

type TracksFavorite struct {
	Limit              int    `json:"limit"`
	Offset             int    `json:"offset"`
	TotalNumberOfItems int    `json:"totalNumberOfItems"`
	Items              []Item `json:"items"`
}

type Creator struct {
	Id int `json:"id"`
}

type Playlist struct {
	UUID           string  `json:"uuid"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Creator        Creator `json:"creator"`
	Type           string  `json:"type"`
	IsPublic       bool    `json:"publicPlaylist"`
	Created        string  `json:"created"`
	LastUpdated    string  `json:"lastUpdated"`
	NumberOfTracks int     `json:"numberOfTracks"`
	Duration       int     `json:"duration"`
	Url            string  `json:"url"`
	Image          string  `json:"image"`
	SquareImage    string  `json:"squareImage"`
	Popularity     int     `json:"popularity"`
	Tracks         *Tracks
}

type User struct {
	Id           int
	Username     string
	FirstName    string
	LastName     string
	Email        string
	CountryCode  string
	Created      string
	Picture      string
	Newsletter   bool
	AcceptedEULA bool
	Gender       bool
	DateOfBirth  string
	FacebookUid  int
}
