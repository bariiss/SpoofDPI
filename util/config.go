package util

import (
	"fmt"
	"regexp"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
)

type Config struct {
	Addr            string
	Port            int
	DnsAddr         string
	DnsPort         int
	DnsIPv4Only     bool
	EnableDoh       bool
	Debug           bool
	Silent          bool
	SystemProxy     bool
	Timeout         int
	WindowSize      int
	AllowedPatterns []*regexp.Regexp
}

var config *Config

func GetConfig() *Config {
	if config == nil {
		config = new(Config)
	}
	return config
}

func (c *Config) Load(args *Args) {
	c.Addr = args.Addr
	c.Port = int(args.Port)
	c.DnsAddr = args.DnsAddr
	c.DnsPort = int(args.DnsPort)
	c.DnsIPv4Only = args.DnsIPv4Only
	c.Debug = args.Debug
	c.EnableDoh = args.EnableDoh
	c.Silent = args.Silent
	c.SystemProxy = args.SystemProxy
	c.Timeout = int(args.Timeout)
	c.AllowedPatterns = parseAllowedPattern(args.AllowedPattern)
	c.WindowSize = int(args.WindowSize)
}

func parseAllowedPattern(patterns StringArray) []*regexp.Regexp {
	var allowedPatterns []*regexp.Regexp

	for _, pattern := range patterns {
		allowedPatterns = append(allowedPatterns, regexp.MustCompile(pattern))
	}

	return allowedPatterns
}

func PrintColoredBanner() {
	cyan := putils.LettersFromStringWithStyle("Spoof", pterm.NewStyle(pterm.FgCyan))
	purple := putils.LettersFromStringWithStyle("DPI", pterm.NewStyle(pterm.FgMagenta))
	err := pterm.DefaultBigText.WithLetters(cyan, purple).Render()
	if err != nil {
		return
	}

	err = pterm.DefaultBulletList.WithItems([]pterm.BulletListItem{
		{Level: 0, Text: "ADDR    : " + fmt.Sprint(config.Addr)},
		{Level: 0, Text: "PORT    : " + fmt.Sprint(config.Port)},
		{Level: 0, Text: "DNS     : " + fmt.Sprint(config.DnsAddr)},
		{Level: 0, Text: "DEBUG   : " + fmt.Sprint(config.Debug)},
		{Level: 0, Text: "SILENT  : " + fmt.Sprint(config.Silent)},
		{Level: 0, Text: "SYSTEM  : " + fmt.Sprint(config.SystemProxy)},
		{Level: 0, Text: "TIMEOUT : " + fmt.Sprint(config.Timeout)},
		{Level: 0, Text: "WINDOW  : " + fmt.Sprint(config.WindowSize)},
		{Level: 0, Text: "DOH     : " + fmt.Sprint(config.EnableDoh)},
		{Level: 0, Text: "DNSPORT : " + fmt.Sprint(config.DnsPort)},
		{Level: 0, Text: "DNSV4   : " + fmt.Sprint(config.DnsIPv4Only)},
		{Level: 0, Text: "ALLOWED : " + fmt.Sprint(config.AllowedPatterns)},
	}).Render()
	if err != nil {
		return
	}

	pterm.DefaultBasicText.Println("Çıkmak için 'CTRL + c' tuşlarına basın.")
}
