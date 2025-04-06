package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/essentialkaos/ek/v13/fmtc"
	"github.com/essentialkaos/ek/v13/fmtutil"
	"github.com/essentialkaos/ek/v13/options"
	"github.com/essentialkaos/ek/v13/pager"
	"github.com/essentialkaos/ek/v13/support"
	"github.com/essentialkaos/ek/v13/support/deps"
	"github.com/essentialkaos/ek/v13/terminal"
	"github.com/essentialkaos/ek/v13/terminal/tty"
	"github.com/essentialkaos/ek/v13/timeutil"
	"github.com/essentialkaos/ek/v13/usage"
	"github.com/essentialkaos/ek/v13/usage/completion/bash"
	"github.com/essentialkaos/ek/v13/usage/completion/fish"
	"github.com/essentialkaos/ek/v13/usage/completion/zsh"
	"github.com/essentialkaos/ek/v13/usage/man"
	"github.com/essentialkaos/ek/v13/usage/update"

	"github.com/tidwall/gjson"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Basic utility info
const (
	APP  = "lj"
	VER  = "0.1.1"
	DESC = "Tool for viewing JSON logs"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Options
const (
	OPT_PAGER    = "P:pager"
	OPT_FOLLOW   = "F:follow"
	OPT_NO_COLOR = "nc:no-color"
	OPT_HELP     = "h:help"
	OPT_VER      = "v:version"

	OPT_VERB_VER     = "vv:verbose-version"
	OPT_COMPLETION   = "completion"
	OPT_GENERATE_MAN = "generate-man"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// optMap contains information about all supported options
var optMap = options.Map{
	OPT_FOLLOW:   {Type: options.BOOL},
	OPT_PAGER:    {Type: options.BOOL},
	OPT_NO_COLOR: {Type: options.BOOL},
	OPT_HELP:     {Type: options.BOOL},
	OPT_VER:      {Type: options.MIXED},

	OPT_VERB_VER:     {Type: options.BOOL},
	OPT_COMPLETION:   {},
	OPT_GENERATE_MAN: {Type: options.BOOL},
}

// color tags for app name and version
var colorTagApp, colorTagVer string

// textColors is map with message text colors
var textColors = map[string]string{
	"":      "",
	"debug": "{s-}",
	"info":  "",
	"warn":  "{#220}",
	"error": "{#208}",
	"fatal": "{#196}",
}

// textColors is a map with marker colors
var markerColors = map[string]string{
	"":      "{s-}",
	"debug": "{s-}",
	"info":  "{s-}",
	"warn":  "{#220}",
	"error": "{#208}",
	"fatal": "{#196}",
}

// labels is a map with level labels
var labels = map[string]string{
	"warn":  "WARN",
	"error": "ERR",
	"fatal": "CRIT",
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Run is main utility function
func Run(gitRev string, gomod []byte) {
	preConfigureUI()

	args, errs := options.Parse(optMap)

	if !errs.IsEmpty() {
		terminal.Error("Options parsing errors:")
		terminal.Error(errs.Error("- "))
		os.Exit(1)
	}

	configureUI()

	switch {
	case options.Has(OPT_COMPLETION):
		os.Exit(printCompletion())
	case options.Has(OPT_GENERATE_MAN):
		printMan()
		os.Exit(0)
	case options.GetB(OPT_VER):
		genAbout(gitRev).Print(options.GetS(OPT_VER))
		os.Exit(0)
	case options.GetB(OPT_VERB_VER):
		support.Collect(APP, VER).
			WithRevision(gitRev).
			WithDeps(deps.Extract(gomod)).
			Print()
		os.Exit(0)
	case options.GetB(OPT_HELP) || (!hasStdinData() && len(args) == 0):
		genUsage().Print()
		os.Exit(0)
	}

	err := process(args)

	if err != nil {
		terminal.Error(err.Error())
		os.Exit(1)
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// preConfigureUI preconfigures UI based on information about user terminal
func preConfigureUI() {
	if !tty.IsTTY() {
		fmtc.DisableColors = true
	}

	switch {
	case fmtc.IsTrueColorSupported():
		colorTagApp, colorTagVer = "{*}{#35D0B6}", "{#35D0B6}"
	case fmtc.Is256ColorsSupported():
		colorTagApp, colorTagVer = "{*}{#79}", "{#79}"
	default:
		colorTagApp, colorTagVer = "{*}{c}", "{c}"
	}

	fmtutil.SeparatorColorTag = "{s-}"
	fmtutil.SeparatorTitleColorTag = "{s-}"
	fmtutil.SeparatorTitleAlign = "c"
}

// configureUI configures user interface
func configureUI() {
	if options.GetB(OPT_NO_COLOR) {
		fmtc.DisableColors = true
	}
}

// process starts arguments processing
func process(args options.Arguments) error {
	source, filters, err := getDataSource(args)

	if err != nil {
		return err
	}

	if options.GetB(OPT_FOLLOW) {
		readDataStream(source, parseFilters(filters))
	} else {
		readData(source, parseFilters(filters))
	}

	return nil
}

// getSource returns data source
func getDataSource(args options.Arguments) (*os.File, []string, error) {
	if hasStdinData() {
		return os.Stdin, args.Strings(), nil
	}

	fd, err := os.OpenFile(args.Get(0).Clean().String(), os.O_RDONLY, 0)

	if err != nil {
		return nil, nil, fmt.Errorf("Can't open file for reading: %w", err)
	}

	return fd, args[1:].Strings(), nil
}

// readData reads all data from given source
func readData(source *os.File, filters Filters) {
	r := bufio.NewReader(source)
	s := bufio.NewScanner(r)

	if options.GetB(OPT_PAGER) {
		if pager.Setup() == nil {
			defer pager.Complete()
		}
	}

	for s.Scan() {
		data := s.Text()
		data = strings.TrimSpace(data)

		if data == "" {
			continue
		}

		renderLine(data, filters)
	}

	source.Close()
}

// readDataStream reads stream of data from given source
func readDataStream(source *os.File, filters Filters) {
	r := bufio.NewReader(source)
	lastPrint := time.Now()

	for {
		line, err := r.ReadString('\n')

		if err != nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		line = strings.TrimRight(line, "\r\n")

		if time.Since(lastPrint) > 30*time.Second {
			fmtutil.Separator(true, timeutil.ShortDuration(lastPrint, false))
		}

		if renderLine(line, filters) {
			lastPrint = time.Now()
		}
	}
}

// renderLine renders log line
func renderLine(line string, filters Filters) bool {
	info := gjson.Parse(line).Map()

	if !filters.IsMatch(info) {
		return false
	}

	var msg, level, caller string
	var ts float64

	for k, v := range info {
		switch k {
		case "msg", "log":
			msg = v.String()
			delete(info, k)
		case "level":
			level = v.String()
			delete(info, k)
		case "caller":
			caller = v.String()
			delete(info, k)
		case "ts":
			ts = v.Float()
			delete(info, k)
		}
	}

	if msg == "" {
		return false
	}

	recDate := time.UnixMicro(int64(ts * 1_000_000))

	fmtc.Print(markerColors[level] + "▎{!}")

	fmtc.Printf(
		"{s-}[ {s}%s{s-}.%s ]{!} ",
		timeutil.Format(recDate, "%y/%m/%d %H:%M:%S"),
		timeutil.Format(recDate, "%K"),
	)

	switch level {
	case "warn", "error", "fatal":
		fmtc.Printf(textColors[level]+"{@}{*} %s {!} ", labels[level])
	}

	if caller != "" {
		fmtc.Printf("{s-}(%s){!} ", caller)
	}

	fmtc.Printf(textColors[level]+"%s{!}\n", msg)

	if len(info) != 0 {
		prefixSize := 26

		if caller != "" {
			prefixSize += len(caller) + 3
		}

		renderFields(level, prefixSize, info)
	}

	return true
}

// renderFields renders log fields
func renderFields(level string, prefixSize int, fields map[string]gjson.Result) {
	var f, m string

	for k, v := range fields {
		switch v.Type {
		case gjson.String:
			f = fmt.Sprintf("%s:\"%v\"", k, v.Value())
		case gjson.False, gjson.True:
			f = fmt.Sprintf("%s:%t", k, v.Bool())
		case gjson.Null:
			f = fmt.Sprintf("%s:nil", k)
		default:
			f = fmt.Sprintf("%s:%v", k, v.Value())
		}

		if len(m)+len(f) > 88 {
			fmtc.Print(markerColors[level] + "▎{!}" + strings.Repeat(" ", prefixSize))
			fmtc.Printfn("{#65}%s{!}", strings.TrimRight(m, " •"))
			m = ""
		}

		m += f + " • "
	}

	fmtc.Print(markerColors[level] + "▎{!}" + strings.Repeat(" ", prefixSize))
	fmtc.Printfn("{#65}%s{!}", strings.TrimRight(m, " •"))
}

// hasStdinData return true if there is some data in stdin
func hasStdinData() bool {
	stdin, err := os.Stdin.Stat()

	if err != nil {
		return false
	}

	if stdin.Mode()&os.ModeCharDevice != 0 {
		return false
	}

	return true
}

// ////////////////////////////////////////////////////////////////////////////////// //

// printCompletion prints completion for given shell
func printCompletion() int {
	info := genUsage()

	switch options.GetS(OPT_COMPLETION) {
	case "bash":
		fmt.Print(bash.Generate(info, APP))
	case "fish":
		fmt.Print(fish.Generate(info, APP))
	case "zsh":
		fmt.Print(zsh.Generate(info, optMap, APP))
	default:
		return 1
	}

	return 0
}

// printMan prints man page
func printMan() {
	fmt.Println(man.Generate(genUsage(), genAbout("")))
}

// genUsage generates usage info
func genUsage() *usage.Info {
	info := usage.NewInfo("", "?source|filter", "?filter…")

	info.AddSpoiler(`You can filter log records using a simple query language.

  {s}•{!} {b}value{!}        {s}—{!} search for occurrences in {c}msg{!} field
  {s}•{!} {c}field{!}{s}:{!}{b}value{!}  {s}—{!} positive exact search
  {s}•{!} {c}field{!}{s}:{!}{y}!{!}{b}value{!} {s}—{!} negative exact search
  {s}•{!} {c}field{!}{s}:{!}{y}~{!}{b}value{!} {s}—{!} search for occurrences
  {s}•{!} {c}field{!}{s}:{!}{y}>{!}{b}value{!} {s}—{!} equal or greater
  {s}•{!} {c}field{!}{s}:{!}{y}<{!}{b}value{!} {s}—{!} equal or less`)

	info.AppNameColorTag = colorTagApp

	info.AddOption(OPT_FOLLOW, "Read log stream")
	info.AddOption(OPT_PAGER, "Paginate output")
	info.AddOption(OPT_NO_COLOR, "Disable colors in output")
	info.AddOption(OPT_HELP, "Show this help message")
	info.AddOption(OPT_VER, "Show version")

	info.AddRawExample(
		"lj log.json",
		"Read log file",
	)

	info.AddRawExample(
		"lj < log.json",
		"Read log file with redirect",
	)

	info.AddRawExample(
		"lj -P log.json",
		"Read log file with pager",
	)

	info.AddRawExample(
		"tail -100 log.json | lj ",
		"Read log file from the tail and filter data",
	)

	info.AddRawExample(
		"kubectl logs -f mypod | lj -F",
		"Read log from k8s pod",
	)

	info.AddRawExample(
		"lj log.json level:warn 'caller:~app/db.go' 'proc-time:>15'",
		"Read log file and filter records",
	)

	return info
}

// genAbout generates info about version
func genAbout(gitRev string) *usage.About {
	about := &usage.About{
		App:     APP,
		Version: VER,
		Desc:    DESC,
		Year:    2009,
		Owner:   "ESSENTIAL KAOS",

		AppNameColorTag: colorTagApp,
		VersionColorTag: colorTagVer,
		DescSeparator:   "{s}—{!}",

		License:       "Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>",
		BugTracker:    "https://github.com/essentialkaos/lj/issues",
		UpdateChecker: usage.UpdateChecker{"essentialkaos/lj", update.GitHubChecker},
	}

	if gitRev != "" {
		about.Build = "git:" + gitRev
	}

	return about
}

// ////////////////////////////////////////////////////////////////////////////////// //
