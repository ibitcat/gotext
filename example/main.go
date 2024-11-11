package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/ibitcat/gotext"
)

func main() {
	// Set PO content
	str := `msgid ""
msgstr ""
"Project-Id-Version: \n"
"POT-Creation-Date: \n"
"PO-Revision-Date: \n"
"Last-Translator: \n"
"Language-Team: \n"
"Language: zh_CN\n"
"MIME-Version: 1.0\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"X-Generator: Poedit 3.4.2\n"

# test
msgctxt "问候1"
msgid "你好"
msgstr "hello"

#, fuzzy
msgctxt "问候2"
msgid "你好"
msgstr ""
"hello\n"
"hi"

# 测试注释
msgid "世界"
msgstr "world"
`

	// re := regexp.MustCompile(`(?s)\S+?.+\n`)
	re := regexp.MustCompile(`\n\s*\n`)
	tables := re.Split(str, -1)
	if len(tables) == 0 {
		panic("sql文件为空")
	}

	// Create Po object
	po := new(gotext.Po)
	po = gotext.NewPo()
	po.Parse([]byte(str))
	po.SetRefs("世界", []string{"aaa", "bbbb"})
	poBytes, _ := po.MarshalText()
	fmt.Println(string(poBytes))
	os.WriteFile("output.po", poBytes, 0o644)

	fmt.Println()
	fmt.Println(po.GetC("你好", "问候2"))

	fmt.Println("-------------")
	l := gotext.NewLocale("./locales", "en_US")
	fmt.Println("GetPath", l.GetPath())
	l.AddDomain("default")
	l.AddRefs("猫", "aaa:1", "bbbb:2", "bbbb:3", "bbbb:4", "bbbb:5", "bbbb:6")
	l.MarshalPo()
	fmt.Println(l.Get("世界"))
}
