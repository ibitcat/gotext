/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package gotext

import (
	"strings"
)

// Translation is the struct for the Translations parsed via Po or Mo files and all coming parsers
type Translation struct {
	ID       string
	PluralID string
	Trs      map[int]string
	Refs     []string
	dirty    bool

	hasCtx            bool
	Ctx               string
	TranslatorComment []string // #  translator-comments // TrimSpace
	ExtractedComment  []string // #. extracted-comments
	Flags             []string // #, fuzzy,c-format,range:0..10
	PrevMsgContext    string   // #| msgctxt previous-context
	PrevMsgId         string   // #| msgid previous-untranslated-string
}

// NewTranslation returns the Translation object and initialized it.
func NewTranslation() *Translation {
	return &Translation{
		Trs: make(map[int]string),
	}
}

func NewTranslationWithRefs(refs []string) *Translation {
	return &Translation{
		Trs:  make(map[int]string),
		Refs: refs,
	}
}

func (t *Translation) IsStale() bool {
	return !t.dirty
}

func (t *Translation) SetRefs(refs []string) {
	t.Refs = refs
	t.dirty = true
}

func (t *Translation) hasRef(ref string) bool {
	for i := 0; i < len(t.Refs); i++ {
		if ref == t.Refs[i] {
			return true
		}
	}
	return false
}

func (t *Translation) AddRefs(refs []string) {
	for i := 0; i < len(refs); i++ {
		if !t.hasRef(refs[i]) {
			t.dirty = true
			t.Refs = append(t.Refs, refs[i])
		}
	}
}

func (t *Translation) ClearRefs() {
	t.Refs = t.Refs[:0]
	t.dirty = true
}

func (t *Translation) Set(str string) {
	t.Trs[0] = str
	t.dirty = true
}

// Get returns the string of the translation
func (t *Translation) Get() string {
	// Look for Translation index 0
	if _, ok := t.Trs[0]; ok {
		if t.Trs[0] != "" {
			return t.Trs[0]
		}
	}

	// Return untranslated id by default
	return t.ID
}

func (t *Translation) SetN(n int, str string) {
	t.Trs[n] = str
	t.dirty = true
}

// GetN returns the string of the plural translation
func (t *Translation) GetN(n int) string {
	// Look for Translation index
	if _, ok := t.Trs[n]; ok {
		if t.Trs[n] != "" {
			return t.Trs[n]
		}
	}

	// Return untranslated singular if corresponding
	if n == 0 {
		return t.ID
	}

	// Return untranslated plural by default
	return t.PluralID
}

// IsTranslated reports whether a string is translated
func (t *Translation) IsTranslated() bool {
	tr, ok := t.Trs[0]
	return tr != "" && ok
}

// IsTranslatedN reports whether a plural string is translated
func (t *Translation) IsTranslatedN(n int) bool {
	tr, ok := t.Trs[n]
	return tr != "" && ok
}

// comment
func (t *Translation) readTranslatorComment(s string) {
	t.TranslatorComment = append(t.TranslatorComment, s)
}

func (t *Translation) readExtractedComment(s string) {
	t.ExtractedComment = append(t.ExtractedComment, s)
}

func (t *Translation) readReferenceComment(s string) {
	const prefix = "#:"
	if len(s) < len(prefix) || s[:len(prefix)] != prefix {
		return
	}

	ss := strings.Split(strings.TrimSpace(s[len(prefix):]), " ")
	for i := 0; i < len(ss); i++ {
		if len(ss[i]) > 0 {
			t.Refs = append(t.Refs, ss[i])
		}
	}
}

func (t *Translation) readFlagsComment(s string) {
	const prefix = "#,"

	if len(s) < len(prefix) || s[:len(prefix)] != prefix {
		return
	}
	ss := strings.Split(strings.TrimSpace(s[len(prefix):]), ",")
	for i := 0; i < len(ss); i++ {
		t.Flags = append(t.Flags, strings.TrimSpace(ss[i]))
	}
}

func (t *Translation) readPrevMsg(r *lineReader, l string) {
	// #| msgid "aaaa \n"
	// #| "xxxxx"
	const prefix = "#|"
	if len(l) < len(prefix) || l[:len(prefix)] != prefix {
		return
	}

	str := strings.TrimSpace(l[2:])
	if strings.HasPrefix(str, "msgid") {
		s, _ := strings.CutPrefix(str, "msgid")
		t.PrevMsgId, _ = r.readString(s, prefix)
	} else if strings.HasPrefix(str, "msgctxt") {
		s, _ := strings.CutPrefix(str, "msgctxt")
		t.PrevMsgContext, _ = r.readString(s, prefix)
	}
}
