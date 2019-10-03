package config

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"io/ioutil"
	"path/filepath"
)

const configFilename = "config.json"
const configCertFilename = "cert.pem"

const DefaultServerHostname = "localhost"
const DefaultServerPort = 6620
const DefaultServerSsl = true
const DefaultServerSelfSigned = true
const DefaultUsername = "lyra"
const DefaultPassword = "lyra"
const DefaultBufferLength = 100
const DefaultTimeout = 600
const DefaultSortLanguage = "english"

var localeTags = map[string]language.Tag{
	"afrikaans":            language.Afrikaans,
	"amharic":              language.Amharic,
	"arabic":               language.Arabic,
	"modernStandardArabic": language.ModernStandardArabic,
	"azerbaijani":          language.Azerbaijani,
	"bulgarian":            language.Bulgarian,
	"bengali":              language.Bengali,
	"catalan":              language.Catalan,
	"czech":                language.Czech,
	"danish":               language.Danish,
	"german":               language.German,
	"greek":                language.Greek,
	"english":              language.English,
	"americanEnglish":      language.AmericanEnglish,
	"britishEnglish":       language.BritishEnglish,
	"spanish":              language.Spanish,
	"europeanSpanish":      language.EuropeanSpanish,
	"latinAmericanSpanish": language.LatinAmericanSpanish,
	"estonian":             language.Estonian,
	"persian":              language.Persian,
	"finnish":              language.Finnish,
	"filipino":             language.Filipino,
	"french":               language.French,
	"canadianFrench":       language.CanadianFrench,
	"gujarati":             language.Gujarati,
	"hebrew":               language.Hebrew,
	"hindi":                language.Hindi,
	"croatian":             language.Croatian,
	"hungarian":            language.Hungarian,
	"armenian":             language.Armenian,
	"indonesian":           language.Indonesian,
	"icelandic":            language.Icelandic,
	"italian":              language.Italian,
	"japanese":             language.Japanese,
	"georgian":             language.Georgian,
	"kazakh":               language.Kazakh,
	"khmer":                language.Khmer,
	"kannada":              language.Kannada,
	"korean":               language.Korean,
	"kirghiz":              language.Kirghiz,
	"lao":                  language.Lao,
	"lithuanian":           language.Lithuanian,
	"latvian":              language.Latvian,
	"macedonian":           language.Macedonian,
	"malayalam":            language.Malayalam,
	"mongolian":            language.Mongolian,
	"marathi":              language.Marathi,
	"malay":                language.Malay,
	"burmese":              language.Burmese,
	"nepali":               language.Nepali,
	"dutch":                language.Dutch,
	"norwegian":            language.Norwegian,
	"punjabi":              language.Punjabi,
	"polish":               language.Polish,
	"portuguese":           language.Portuguese,
	"brazilianPortuguese":  language.BrazilianPortuguese,
	"europeanPortuguese":   language.EuropeanPortuguese,
	"romanian":             language.Romanian,
	"russian":              language.Russian,
	"sinhala":              language.Sinhala,
	"slovak":               language.Slovak,
	"slovenian":            language.Slovenian,
	"albanian":             language.Albanian,
	"serbian":              language.Serbian,
	"serbianLatin":         language.SerbianLatin,
	"swedish":              language.Swedish,
	"swahili":              language.Swahili,
	"tamil":                language.Tamil,
	"telugu":               language.Telugu,
	"thai":                 language.Thai,
	"turkish":              language.Turkish,
	"ukrainian":            language.Ukrainian,
	"urdu":                 language.Urdu,
	"uzbek":                language.Uzbek,
	"vietnamese":           language.Vietnamese,
	"chinese":              language.Chinese,
	"simplifiedChinese":    language.SimplifiedChinese,
	"traditionalChinese":   language.TraditionalChinese,
	"zulu":                 language.Zulu,
}

type ClientConfig struct {
	ConfigDir string
	DebugMode bool

	*ClientEditableConfig

	collator *collate.Collator
}

func (c *ClientConfig) Collator() *collate.Collator {
	if c.collator == nil {
		c.collator = collate.New(localeTags[c.SortLanguage])
	}
	return c.collator
}

type ClientEditableConfig struct {
	ServerHostname   string `json:"serverHostname"`
	ServerPort       int64  `json:"serverPort"`
	ServerSsl        bool   `json:"serverSsl"`
	ServerSelfSigned bool   `json:"serverSelfSigned"`
	SortLanguage     string `json:"sortLanguage"`
	BufferLength     int64  `json:"bufferLength"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	Timeout          int64  `json:"timeout"`
}

func NewClientEditableConfig(draftClientEditableConfig *ClientEditableConfig) *ClientEditableConfig {
	var clientEditableConfig ClientEditableConfig

	if draftClientEditableConfig == nil {
		clientEditableConfig = ClientEditableConfig{
			ServerHostname:   DefaultServerHostname,
			ServerPort:       DefaultServerPort,
			ServerSsl:        DefaultServerSsl,
			ServerSelfSigned: DefaultServerSelfSigned,
			SortLanguage:     DefaultSortLanguage,
			BufferLength:     DefaultBufferLength,
			Username:         DefaultUsername,
			Password:         DefaultPassword,
			Timeout:          DefaultTimeout,
		}
	} else {
		clientEditableConfig = *draftClientEditableConfig

		// Check config values
		if _, ok := localeTags[clientEditableConfig.SortLanguage]; !ok {
			clientEditableConfig.SortLanguage = DefaultSortLanguage
		}
		if clientEditableConfig.BufferLength <= 10 {
			clientEditableConfig.BufferLength = 10
		} else if clientEditableConfig.BufferLength > 5000 {
			clientEditableConfig.BufferLength = 5000
		}
		if clientEditableConfig.Timeout <= 10 {
			clientEditableConfig.Timeout = 10
		} else if clientEditableConfig.Timeout >= 3600 {
			clientEditableConfig.Timeout = 3600
		}

	}

	return &clientEditableConfig
}

func (c *ClientConfig) Save() {
	logrus.Debugf("Save config file: %s", c.GetCompleteConfigFilename())
	rawConfig, err := json.MarshalIndent(c.ClientEditableConfig, "", "\t")
	if err != nil {
		logrus.Fatalf("Unable to serialize config file: %v\n", err)
	}
	err = ioutil.WriteFile(c.GetCompleteConfigFilename(), rawConfig, 0660)
	if err != nil {
		logrus.Fatalf("Unable to save config file: %v\n", err)
	}
}

func (c ClientConfig) GetCompleteConfigFilename() string {
	return filepath.Join(c.ConfigDir, configFilename)
}

func (c ClientConfig) GetCompleteConfigCertFilename() string {
	return filepath.Join(c.ConfigDir, configCertFilename)
}
