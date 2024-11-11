package gotext

import (
	"io"
	"regexp"
	"strconv"
	"strings"
)

var reStringLine = regexp.MustCompile(`^\s*".*"\s*$`)

type lineReader struct {
	lines   []string
	linenum int
}

func newLineReader(data string) *lineReader {
	data = strings.Replace(data, "\r", "", -1)
	lines := strings.Split(data, "\n")
	return &lineReader{lines: lines}
}

func (r *lineReader) skipBlankLine() error {
	for ; r.linenum < len(r.lines); r.linenum++ {
		if strings.TrimSpace(r.lines[r.linenum]) != "" {
			break
		}
	}
	if r.linenum >= len(r.lines) {
		return io.EOF
	}
	return nil
}

func (r *lineReader) currentLinenum() int {
	return r.linenum
}

func (r *lineReader) currentLine() (s string, linenum int, err error) {
	if r.linenum >= len(r.lines) {
		err = io.EOF
		return
	}
	s, linenum = r.lines[r.linenum], r.linenum
	return
}

func (r *lineReader) readLine() (s string, linenum int, err error) {
	if r.linenum >= len(r.lines) {
		err = io.EOF
		return
	}
	s, linenum = r.lines[r.linenum], r.linenum
	r.linenum++
	return
}

func (r *lineReader) unreadLine() {
	if r.linenum >= 0 {
		r.linenum--
	}
}

func (r *lineReader) readString(head string, vars ...string) (msg string, err error) {
	var prefix string
	if len(vars) > 0 && vars[0] != "" {
		prefix = vars[0]
	}

	if !reStringLine.MatchString(head) {
		return
	}
	msg, err = strconv.Unquote(head)
	if err != nil {
		return
	}

	var s string
	for {
		if s, _, err = r.readLine(); err != nil {
			return
		}
		if len(prefix) > 0 {
			s, _ = strings.CutPrefix(s, prefix)
		}
		if !reStringLine.MatchString(s) {
			r.unreadLine()
			break
		}
		uqs, _ := strconv.Unquote(s)
		msg += uqs
	}
	return
}
