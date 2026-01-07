package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"hash"
	"hash/fnv"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var adjectives = []string{
	// Nature/Weather
	"swift", "bright", "calm", "dark", "eager", "fair", "grand", "happy",
	"idle", "jolly", "keen", "lively", "merry", "noble", "odd", "proud",
	"quiet", "rapid", "smart", "tall", "unique", "vivid", "warm", "young",
	"ancient", "bold", "clever", "daring", "elegant", "fierce", "gentle", "humble",
	// Colors
	"azure", "bronze", "coral", "crimson", "golden", "ivory", "jade", "amber",
	"silver", "violet", "scarlet", "emerald", "cobalt", "copper", "pearl", "onyx",
	// Textures/Qualities
	"smooth", "rough", "silky", "crisp", "soft", "sharp", "clear", "dense",
	"light", "heavy", "thin", "thick", "wide", "narrow", "deep", "shallow",
	// Emotions/Character
	"brave", "shy", "wild", "tame", "free", "bound", "lost", "found",
	"hidden", "open", "secret", "known", "silent", "loud", "still", "moving",
	// Size/Shape
	"tiny", "huge", "vast", "small", "great", "minor", "major", "prime",
	"round", "flat", "curved", "straight", "twisted", "bent", "hollow", "solid",
	// Time/Age
	"fresh", "aged", "new", "old", "early", "late", "first", "last",
	"eternal", "brief", "long", "short", "quick", "slow", "instant", "gradual",
}

var nouns = []string{
	// Nature
	"apple", "brook", "cloud", "delta", "eagle", "flame", "grove", "harbor",
	"island", "jungle", "kite", "lantern", "meadow", "nest", "ocean", "peak",
	"quartz", "river", "stone", "tower", "umbrella", "valley", "wave", "zenith",
	"anchor", "beacon", "canyon", "dune", "forest", "glacier", "horizon", "inlet",
	// Celestial
	"comet", "cosmos", "eclipse", "galaxy", "meteor", "nebula", "nova", "orbit",
	"planet", "pulsar", "quasar", "saturn", "star", "sun", "moon", "venus",
	// Elements
	"crystal", "diamond", "ember", "frost", "gem", "ice", "magma", "mist",
	"prism", "rain", "sand", "snow", "spark", "steam", "thunder", "wind",
	// Places
	"castle", "citadel", "dome", "fort", "gate", "hall", "keep", "manor",
	"palace", "plaza", "port", "realm", "shrine", "temple", "vault", "villa",
	// Objects
	"arrow", "blade", "crown", "drum", "flag", "globe", "hammer", "horn",
	"jewel", "key", "lamp", "mirror", "needle", "orb", "ring", "scroll",
	// Abstract
	"cipher", "code", "echo", "enigma", "flux", "glyph", "helix", "loop",
	"matrix", "nexus", "path", "pulse", "rhythm", "signal", "trace", "vertex",
}

var animals = []string{
	// Mammals
	"fox", "owl", "bear", "deer", "hawk", "wolf", "lynx", "seal",
	"lion", "tiger", "puma", "orca", "whale", "moose", "bison", "horse",
	"rabbit", "badger", "otter", "beaver", "marten", "ferret", "mink", "stoat",
	"panther", "jaguar", "leopard", "cheetah", "cougar", "bobcat", "ocelot", "serval",
	// Birds
	"crow", "dove", "ibis", "jay", "kiwi", "raven", "swan", "wren",
	"falcon", "condor", "osprey", "vulture", "pelican", "heron", "crane", "stork",
	"parrot", "toucan", "macaw", "finch", "sparrow", "robin", "thrush", "oriole",
	"eagle", "hawk", "kite", "harrier", "buzzard", "merlin", "goshawk", "kestrel",
	// Reptiles/Amphibians
	"frog", "newt", "viper", "cobra", "python", "gecko", "iguana", "turtle",
	// Fish/Sea
	"salmon", "trout", "bass", "pike", "perch", "carp", "tuna", "marlin",
	// Insects
	"ant", "bee", "wasp", "moth", "beetle", "cricket", "mantis", "firefly",
	// Mythical (for variety)
	"dragon", "phoenix", "griffin", "sphinx", "hydra", "kraken", "wyrm", "roc",
}

var loremWords = []string{
	// Classic Lorem
	"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit",
	"sed", "do", "eiusmod", "tempor", "incididunt", "ut", "labore", "et",
	"dolore", "magna", "aliqua", "enim", "ad", "minim", "veniam", "quis",
	"nostrud", "exercitation", "ullamco", "laboris", "nisi", "aliquip", "ex", "ea",
	"commodo", "consequat", "duis", "aute", "irure", "in", "reprehenderit", "voluptate",
	"velit", "esse", "cillum", "fugiat", "nulla", "pariatur", "excepteur", "sint",
	"occaecat", "cupidatat", "non", "proident", "sunt", "culpa", "qui", "officia",
	"deserunt", "mollit", "anim", "id", "est", "laborum", "ac", "ante",
	// Extended Lorem
	"pellentesque", "habitant", "morbi", "tristique", "senectus", "netus", "malesuada", "fames",
	"turpis", "egestas", "proin", "nibh", "nisl", "condimentum", "viverra", "maecenas",
	"accumsan", "lacus", "vel", "facilisis", "volutpat", "blandit", "cursus", "risus",
	"ultricies", "gravida", "dictum", "fusce", "placerat", "orci", "porta", "felis",
	"bibendum", "arcu", "vitae", "elementum", "curabitur", "sodales", "ligula", "pulvinar",
	"mattis", "nunc", "fringilla", "urna", "porttitor", "rhoncus", "purus", "quam",
	"fermentum", "posuere", "leo", "diam", "sollicitudin", "auctor", "ornare", "odio",
	"semper", "lectus", "mauris", "tincidunt", "lobortis", "feugiat", "vivamus", "atque",
	// Technical sounding
	"protocol", "interface", "module", "system", "process", "function", "method", "object",
	"variable", "constant", "parameter", "argument", "return", "value", "result", "output",
	"input", "stream", "buffer", "cache", "queue", "stack", "array", "list",
	"node", "tree", "graph", "edge", "vertex", "path", "route", "link",
	"request", "response", "client", "server", "socket", "packet", "frame", "segment",
	"token", "session", "context", "scope", "state", "event", "handler", "callback",
	"config", "setting", "option", "feature", "flag", "mode", "level", "status",
	"action", "trigger", "effect", "update", "change", "delta", "diff", "patch",
}

// Pre-computed capitalized versions to avoid allocations
var (
	adjectivesCap []string
	nounsCap      []string
)

func init() {
	adjectivesCap = make([]string, len(adjectives))
	for i, s := range adjectives {
		r, size := utf8.DecodeRuneInString(s)
		adjectivesCap[i] = string(unicode.ToUpper(r)) + s[size:]
	}
	nounsCap = make([]string, len(nouns))
	for i, s := range nouns {
		r, size := utf8.DecodeRuneInString(s)
		nounsCap[i] = string(unicode.ToUpper(r)) + s[size:]
	}
}

type state struct {
	tags         map[string]string
	requirements map[string]string
	params       map[string]string
	customFields map[string]string
	files        map[string]string
	tagCounter   int
	reqCounter   int
	paramCounter int
	cfCounter    int
	fileCounter  int
	buf          strings.Builder // main buffer for row processing
	textBuf      strings.Builder // buffer for readableText
	paramBuf     strings.Builder // buffer for obfuscatePreservingParams
	hasher       hash.Hash32
}

// GOEXPERIMENT=jsonv2 go run ./cmd/anonymize -f input.csv > obfuscated.csv
func main() {
	filename := flag.String("f", "", "file to read")
	flag.Parse()

	if *filename == "" {
		panic("no -f flag provided")
	}

	f, err := os.OpenFile(*filename, os.O_RDONLY, 0o644)
	if err != nil {
		panic(err)
	}

	reader := csv.NewReader(f)
	redacted := csv.NewWriter(os.Stdout)
	reader.ReuseRecord = true
	read, err := reader.Read()
	if err != nil {
		panic(err)
	}
	redacted.Write(read)

	requirementsMDLinkRegex := regexp.MustCompile(`\[(.*)\]\((.*)\)`)

	const (
		HeaderFolder = iota
		HeaderType
		HeaderName
		HeaderLegacyID
		HeaderDraft
		HeaderPriority
		HeaderTags
		HeaderRequirements
		HeaderLinks
		HeaderFiles
		HeaderPreconditions
		HeaderSteps
		HeaderParameterValues
		HeaderTemplateSuffixParams
	)

	anon := state{
		tags:         make(map[string]string, 64),
		requirements: make(map[string]string, 64),
		params:       make(map[string]string, 32),
		customFields: make(map[string]string, 32),
		files:        make(map[string]string, 32),
		hasher:       fnv.New32a(),
	}
	buf := &anon.buf

	for read, err = reader.Read(); err == nil; read, err = reader.Read() {
		read[HeaderName] = anon.readableText(read[HeaderName])

		if read[HeaderTags] != "" {
			buf.Reset()
			for tag := range strings.SplitSeq(read[HeaderTags], ",") {
				if _, ok := anon.tags[tag]; !ok {
					anon.tags[tag] = anon.readableTag()
				}
				buf.WriteString(anon.tags[tag])
				buf.WriteByte(',')
			}
			read[HeaderTags] = buf.String()[:buf.Len()-1]
		}

		if read[HeaderRequirements] != "" {
			buf.Reset()
			for requirement := range strings.SplitSeq(read[HeaderRequirements], ",") {
				matches := requirementsMDLinkRegex.FindStringSubmatch(requirement)
				name := matches[1]

				if _, ok := anon.requirements[name]; !ok {
					anon.requirements[name] = anon.readableRequirement()
				}
				reqID := anon.requirements[name]
				buf.WriteByte('[')
				buf.WriteString(reqID)
				buf.WriteString("](https://example.com/requirements/")
				anon.writeLowerDashed(buf, reqID)
				buf.WriteString("),")
			}
			read[HeaderRequirements] = buf.String()[:buf.Len()-1]
		}

		if read[HeaderFiles] != "" {
			var files []struct {
				ID       string `json:"id"           `
				FileName string `json:"fileName"     `
				MimeType string `json:"mimeType"     `
				Size     int    `json:"size"`
				URL      string `json:"url,omitempty"`
			}
			err = json.Unmarshal([]byte(read[HeaderFiles]), &files)
			if err != nil {
				panic(err)
			}
			for i := range files {
				ext := filepath.Ext(files[i].FileName)
				files[i].ID = anon.readableFileID(files[i].ID)
				files[i].FileName = anon.readableFileName(ext)
				buf.Reset()
				buf.WriteString("https://s3/files/")
				buf.WriteString(files[i].ID)
				buf.WriteByte('/')
				buf.WriteString(files[i].FileName)
				files[i].URL = buf.String()
			}
			buf.Reset()
			err = json.NewEncoder(buf).Encode(files)
			if err != nil {
				panic(err)
			}
			read[HeaderFiles] = buf.String()[:buf.Len()-1] // trim trailing newline
		}

		read[HeaderPreconditions] = anon.obfuscatePreservingParams(read[HeaderPreconditions])

		if read[HeaderSteps] != "" {
			var steps []struct {
				Description  string `json:"description,omitempty"`
				Expected     string `json:"expected,omitempty"`
				SharedStepID *int   `json:"sharedStepId,omitempty"`
				SubSteps     []struct {
					Description string `json:"description,omitempty"`
					Expected    string `json:"expected,omitempty"`
				} `json:"subSteps,omitempty"`
			}
			err = json.Unmarshal([]byte(read[HeaderSteps]), &steps)
			if err != nil {
				panic(err)
			}
			for i := range steps {
				steps[i].Description = anon.obfuscatePreservingParams(steps[i].Description)
				steps[i].Expected = anon.obfuscatePreservingParams(steps[i].Expected)
				for j := range steps[i].SubSteps {
					steps[i].SubSteps[j].Description = anon.obfuscatePreservingParams(steps[i].SubSteps[j].Description)
					steps[i].SubSteps[j].Expected = anon.obfuscatePreservingParams(steps[i].SubSteps[j].Expected)
				}
			}
			buf.Reset()
			err = json.NewEncoder(buf).Encode(steps)
			if err != nil {
				panic(err)
			}
			read[HeaderSteps] = buf.String()[:buf.Len()-1] // trim trailing newline
		}

		var parameterValueKeys map[string]string
		if read[HeaderParameterValues] != "" {
			var parameterValues []struct {
				Priority string            `json:"priority,omitempty"`
				Values   map[string]string `json:"values"`
			}
			err = json.Unmarshal([]byte(read[HeaderParameterValues]), &parameterValues)
			if err != nil {
				panic(err)
			}
			parameterValueKeys = make(map[string]string, len(parameterValues[0].Values))
			for i := range parameterValues {
				obfuscatedMap := make(map[string]string, len(parameterValues[0].Values))
				for k, v := range parameterValues[i].Values {
					readableKey := anon.readableParam(k)
					parameterValueKeys[k] = readableKey
					obfuscatedMap[readableKey] = anon.obfuscatePreservingParams(v)
				}
				parameterValues[i].Values = obfuscatedMap
			}
			buf.Reset()
			err = json.NewEncoder(buf).Encode(parameterValues)
			if err != nil {
				panic(err)
			}
			read[HeaderParameterValues] = buf.String()[:buf.Len()-1] // trim trailing newline
		}

		if read[HeaderTemplateSuffixParams] != "" {
			buf.Reset()
			for key := range strings.SplitSeq(read[HeaderTemplateSuffixParams], ",") {
				buf.WriteString(parameterValueKeys[key])
				buf.WriteByte(',')
			}
			read[HeaderTemplateSuffixParams] = buf.String()[:buf.Len()-1]
		}

		// read custom fields after template suffix params
		for i, jsonVal := range read[HeaderTemplateSuffixParams+1:] {
			// surgically remove json syntax
			// do not worry about escaped quotes as value is hashed anyway
			// skip the {"isDefault":true} values
			if strings.HasPrefix(jsonVal, `{"value":"`) {
				value := jsonVal[len(`{"value":"`) : len(jsonVal)-2]
				buf.Reset()
				buf.WriteString(`{"value":"`)
				buf.WriteString(anon.readableCustomField(value))
				buf.WriteString(`"}`)
				read[HeaderTemplateSuffixParams+1+i] = buf.String()
			}
		}

		redacted.Write(read)
	}

	redacted.Flush()
}

var parameterPlaceholderRegex = regexp.MustCompile(`\${([\w_-]+\w)}`)

func (a *state) obfuscatePreservingParams(s string) string {
	if s == "" {
		return ""
	}
	matches := parameterPlaceholderRegex.FindAllStringSubmatch(s, -1)
	if matches == nil {
		return a.readableText(s)
	}
	split := parameterPlaceholderRegex.Split(s, -1)

	a.paramBuf.Reset()
	a.paramBuf.WriteString(a.readableText(split[0]))
	for i, sub := range split[1:] {
		// Replace the parameter name inside ${...} with its readable version
		paramName := matches[i][1]
		a.paramBuf.WriteString("${")
		a.paramBuf.WriteString(a.readableParam(paramName))
		a.paramBuf.WriteByte('}')
		a.paramBuf.WriteString(a.readableText(sub))
	}
	return a.paramBuf.String()
}

func (a *state) readableText(s string) string {
	if s == "" {
		return ""
	}

	wordCount := countWords(s)
	if wordCount == 0 {
		wordCount = 1
	}

	a.hasher.Reset()
	a.hasher.Write([]byte(s))
	seed := a.hasher.Sum32()

	a.textBuf.Reset()
	loremLen := uint32(len(loremWords))
	for i := range wordCount {
		if i > 0 {
			a.textBuf.WriteByte(' ')
		}
		idx := (seed + uint32(i)) % loremLen
		word := loremWords[idx]
		// Capitalize first word
		if i == 0 {
			for j, r := range word {
				a.textBuf.WriteRune(unicode.ToUpper(r))
				a.textBuf.WriteString(word[j+utf8.RuneLen(r):])
				break
			}
		} else {
			a.textBuf.WriteString(word)
		}
	}
	return a.textBuf.String()
}

func countWords(s string) int {
	count := 0
	inWord := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			inWord = false
		} else if !inWord {
			inWord = true
			count++
		}
	}
	return count
}

func (a *state) readableTag() string {
	adj := adjectives[a.tagCounter%len(adjectives)]
	animal := animals[a.tagCounter%len(animals)]
	a.tagCounter++
	return adj + "-" + animal
}

func (a *state) readableRequirement() string {
	adj := adjectivesCap[a.reqCounter%len(adjectivesCap)]
	noun := nounsCap[a.reqCounter%len(nounsCap)]
	a.reqCounter++
	return "REQ " + adj + " " + noun
}

func (a *state) readableParam(original string) string {
	if existing, ok := a.params[original]; ok {
		return existing
	}
	noun := nouns[a.paramCounter%len(nouns)]
	a.paramCounter++
	readable := "param_" + noun
	a.params[original] = readable
	return readable
}

func (a *state) readableCustomField(original string) string {
	if existing, ok := a.customFields[original]; ok {
		return existing
	}
	adj := adjectivesCap[a.cfCounter%len(adjectivesCap)]
	noun := nounsCap[a.cfCounter%len(nounsCap)]
	a.cfCounter++
	readable := adj + " " + noun
	a.customFields[original] = readable
	return readable
}

func (a *state) readableFileID(original string) string {
	if existing, ok := a.files[original]; ok {
		return existing
	}
	a.fileCounter++
	readable := "file-" + padInt(a.fileCounter, 3)
	a.files[original] = readable
	return readable
}

func (a *state) readableFileName(ext string) string {
	adj := adjectives[a.fileCounter%len(adjectives)]
	noun := nouns[a.fileCounter%len(nouns)]
	return "document-" + adj + "-" + noun + ext
}

func (a *state) writeLowerDashed(buf *strings.Builder, s string) {
	for _, r := range s {
		if r == ' ' {
			buf.WriteByte('-')
		} else {
			buf.WriteRune(unicode.ToLower(r))
		}
	}
}

func padInt(n, w int) string {
	s := strconv.Itoa(n)
	if len(s) >= w {
		return s
	}
	return strings.Repeat("0", w-len(s)) + s
}
