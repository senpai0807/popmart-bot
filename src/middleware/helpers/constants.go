package helpers

import (
	_ "embed"
	"sync"

	"github.com/hajimehoshi/oto"
)

//go:embed assets/success.mp3
var Success []byte

//go:embed assets/decline.mp3
var Decline []byte

var (
	audioCtx *oto.Context
	ctxInit  sync.Once
	initErr  error
)

var (
	Carted     int
	CheckedOut int
	Mu         sync.Mutex
	MaxRetries = 50000
	SecChUa    = `"Not)A;Brand";v="8", "Chromium";v="138", "Google Chrome";v="138"`
	UserAgent  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36"
)

var StateAbbreviations = map[string]string{
	"Alabama": "AL", "Alaska": "AK", "Arizona": "AZ", "Arkansas": "AR", "California": "CA", "Colorado": "CO",
	"Connecticut": "CT", "Delaware": "DE", "District of Columbia": "DC", "Florida": "FL", "Georgia": "GA",
	"Hawaii": "HI", "Idaho": "ID", "Illinois": "IL", "Indiana": "IN", "Iowa": "IA", "Kansas": "KS",
	"Kentucky": "KY", "Louisiana": "LA", "Maine": "ME", "Maryland": "MD", "Massachusetts": "MA",
	"Michigan": "MI", "Minnesota": "MN", "Mississippi": "MS", "Missouri": "MO", "Montana": "MT", "Nebraska": "NE",
	"Nevada": "NV", "New Hampshire": "NH", "New Jersey": "NJ", "New Mexico": "NM", "New York": "NY",
	"North Carolina": "NC", "North Dakota": "ND", "Ohio": "OH", "Oklahoma": "OK", "Oregon": "OR",
	"Pennsylvania": "PA", "Rhode Island": "RI", "South Carolina": "SC", "South Dakota": "SD",
	"Tennessee": "TN", "Texas": "TX", "Utah": "UT", "Vermont": "VT", "Virginia": "VA", "Washington": "WA",
	"West Virginia": "WV", "Wisconsin": "WI", "Wyoming": "WY",
}
