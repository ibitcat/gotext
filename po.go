/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package gotext

import (
	"io/fs"
	"regexp"
	"strconv"
	"strings"
)

/*
Po parses the content of any PO file and provides all the Translation functions needed.
It's the base object used by all package methods.
And it's safe for concurrent use by multiple goroutines by using the sync package for locking.

Example:

	import (
		"fmt"
		"github.com/leonelquinteros/gotext"
	)

	func main() {
		// Create po object
		po := gotext.NewPo()

		// Parse .po file
		po.ParseFile("/path/to/po/file/translations.po")

		// Get Translation
		fmt.Println(po.Get("Translate this"))
	}
*/
type Po struct {
	// these three public members are for backwards compatibility. they are just set to the value in the domain
	Headers     HeaderMap
	Language    string
	PluralForms string

	domain *Domain
	fs     fs.FS
}

var (
	reBlocks    = regexp.MustCompile(`\n\s*\n`)
	reBlankLine = regexp.MustCompile(`^\s*$`)
)

// NewPo should always be used to instantiate a new Po object
func NewPo() *Po {
	po := new(Po)
	po.domain = NewDomain()

	return po
}

// NewPoFS works like NewPO but adds an optional fs.FS
func NewPoFS(filesystem fs.FS) *Po {
	po := NewPo()
	po.fs = filesystem
	return po
}

func (po *Po) GetDomain() *Domain {
	return po.domain
}

// Convenience interfaces
func (po *Po) DropStaleTranslations() {
	po.domain.DropStaleTranslations()
}

func (po *Po) SetRefs(str string, refs []string) {
	po.domain.SetRefs(str, refs)
}
func (po *Po) GetRefs(str string) []string {
	return po.domain.GetRefs(str)
}

func (po *Po) SetPluralResolver(f func(int) int) {
	po.domain.customPluralResolver = f
}

func (po *Po) Set(id, str string) {
	po.domain.Set(id, str)
}
func (po *Po) Get(str string, vars ...interface{}) string {
	return po.domain.Get(str, vars...)
}
func (po *Po) Append(b []byte, str string, vars ...interface{}) []byte {
	return po.domain.Append(b, str, vars...)
}

func (po *Po) SetN(id, plural string, n int, str string) {
	po.domain.SetN(id, plural, n, str)
}
func (po *Po) GetN(str, plural string, n int, vars ...interface{}) string {
	return po.domain.GetN(str, plural, n, vars...)
}
func (po *Po) AppendN(b []byte, str, plural string, n int, vars ...interface{}) []byte {
	return po.domain.AppendN(b, str, plural, n, vars...)
}

func (po *Po) SetC(id, ctx, str string) {
	po.domain.SetC(id, ctx, str)
}
func (po *Po) GetC(str, ctx string, vars ...interface{}) string {
	return po.domain.GetC(str, ctx, vars...)
}
func (po *Po) AppendC(b []byte, str, ctx string, vars ...interface{}) []byte {
	return po.domain.AppendC(b, str, ctx, vars...)
}

func (po *Po) SetNC(id, plural, ctx string, n int, str string) {
	po.domain.SetNC(id, plural, ctx, n, str)
}
func (po *Po) GetNC(str, plural string, n int, ctx string, vars ...interface{}) string {
	return po.domain.GetNC(str, plural, n, ctx, vars...)
}
func (po *Po) AppendNC(b []byte, str, plural string, n int, ctx string, vars ...interface{}) []byte {
	return po.domain.AppendNC(b, str, plural, n, ctx, vars...)
}

func (po *Po) IsTranslated(str string) bool {
	return po.domain.IsTranslated(str)
}
func (po *Po) IsTranslatedN(str string, n int) bool {
	return po.domain.IsTranslatedN(str, n)
}
func (po *Po) IsTranslatedC(str, ctx string) bool {
	return po.domain.IsTranslatedC(str, ctx)
}
func (po *Po) IsTranslatedNC(str string, n int, ctx string) bool {
	return po.domain.IsTranslatedNC(str, n, ctx)
}

func (po *Po) MarshalText() ([]byte, error) {
	return po.domain.MarshalText()
}

func (po *Po) MarshalBinary() ([]byte, error) {
	return po.domain.MarshalBinary()
}

func (po *Po) UnmarshalBinary(data []byte) error {
	return po.domain.UnmarshalBinary(data)
}

func (po *Po) ParseFile(f string) {
	data, err := getFileData(f, po.fs)
	if err != nil {
		return
	}

	po.Parse(data)
}

// Parse loads the translations specified in the provided byte slice (buf)
func (po *Po) Parse(buf []byte) {
	if po.domain == nil {
		panic("NewPo() was not used to instantiate this object")
	}

	// Lock while parsing
	po.domain.trMutex.Lock()
	po.domain.pluralMutex.Lock()
	defer po.domain.trMutex.Unlock()
	defer po.domain.pluralMutex.Unlock()

	blocks := reBlocks.Split(string(buf), -1)
	if len(blocks) == 0 {
		panic("Po file syntax error")
	}

	for i := 0; i < len(blocks); i++ {
		tr := NewTranslation()

		// Get lines
		var err error
		r := newLineReader(string(blocks[i]))
		if err = r.skipBlankLine(); err == nil {
			for {
				var l string
				if l, _, err = r.currentLine(); err != nil {
					break
				}
				if reBlankLine.MatchString(l) {
					break
				}

				l = strings.TrimSpace(l)
				if len(l) > 0 {
					if l[0] == '#' {
						po.parseComment(tr, r)
					} else {
						po.parseString(tr, r)
					}
				}
			}
		}

		if tr.hasCtx {
			ctxTrs := po.domain.contextTranslations
			if _, ok := ctxTrs[tr.Ctx]; !ok {
				ctxTrs[tr.Ctx] = make(map[string]*Translation)
			}
			ctxTrs[tr.Ctx][tr.ID] = tr
		} else {
			po.domain.translations[tr.ID] = tr
		}
	}

	// Parse headers
	po.domain.parseHeaders()

	// set values on this struct
	// this is for backwards compatibility
	po.Language = po.domain.Language
	po.PluralForms = po.domain.PluralForms
	po.Headers = po.domain.Headers
}

// Either preserves comments before the first "msgid", for later round-trip.
// Or preserves source references for a given translation.
func (po *Po) parseComment(tr *Translation, r *lineReader) {
	var l string
	var err error
	if l, _, err = r.readLine(); err != nil {
		return
	}

	switch l[1] {
	case ' ':
		tr.readTranslatorComment(strings.TrimSpace(l[2:]))
	case '.':
		tr.readExtractedComment(strings.TrimSpace(l[2:]))
	case ':':
		tr.readReferenceComment(l)
	case ',':
		tr.readFlagsComment(l)
	case '|':
		tr.readPrevMsg(r, l)
	default:
		tr.readTranslatorComment(strings.TrimSpace(l[1:]))
	}
}

// parseString takes a well formatted string without prefix
// and creates headers or attach multi-line strings when corresponding
func (po *Po) parseString(tr *Translation, r *lineReader) {
	var l string
	var err error
	if l, _, err = r.readLine(); err != nil {
		return
	}

	s := strings.SplitN(l, " ", 2)
	if len(s) != 2 {
		panic("syntax error")
	}

	key, val := s[0], strings.TrimSpace(s[1])
	switch key {
	case "msgctxt":
		tr.Ctx, err = r.readString(val)
		tr.hasCtx = err == nil
	case "msgid":
		tr.ID, _ = r.readString(val)
	case "msgid_plural":
		tr.PluralID, _ = r.readString(val)
		po.domain.pluralTranslations[tr.PluralID] = tr
	case "msgstr":
		tr.Trs[0], _ = r.readString(val)
	default:
		// msgstr[...]
		if strings.HasPrefix(l, "msgstr") {
			sidx := strings.Index(l, "[")
			eidx := strings.Index(l, "]")
			if sidx != -1 && eidx != -1 {
				i, err := strconv.Atoi(l[sidx+1 : eidx])
				if err == nil {
					tr.Trs[i], _ = r.readString(val)
				}
			}
		}
	}
}
