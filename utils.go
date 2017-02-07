package main

import (
	"sort"
	"strings"
)

type gonicKey []string

func (s gonicKey) Less(i, j int) bool {

	return len(s[i]) > len(s[j])
}

func (s gonicKey) Swap(i, j int) {

	s[i], s[j] = s[j], s[i]
}
func (s gonicKey) Len() int {
	return len(s)
}
func isASCIIUpper(r rune) bool {
	return 'A' <= r && r <= 'Z'
}

//LintGonicKeys Gonic转换时候的关键词，init函数里会进行排序，确保有序性，避免因关键词类似而带来的问题
var LintGonicKeys = gonicKey{
	"API",
	"ASCII",
	"CPU",
	"CSS",
	"DNS",
	"EOF",
	"GUID",
	"HTML",
	"HTTP",
	"HTTPS",
	"ID",
	"IP",
	"JSON",
	"LHS",
	"QPS",
	"RAM",
	"RHS",
	"RPC",
	"SLA",
	"SMTP",
	"SSH",
	"TLS",
	"TTL",
	"UI",
	"UID",
	"UUID",
	"URI",
	"URL",
	"UTF8",
	"VM",
	"XML",
	"XSRF",
	"XSS",
}

func init() {
	sort.Sort(LintGonicKeys)
}

//SnakeCasedName 驼峰式命名命名转换，比如 UserName 转为user_name
func SnakeCasedName(name string) string {
	var newstr []rune
	newstr = make([]rune, 0)
	for idx, chr := range name {
		if isUpper := 'A' <= chr && chr <= 'Z'; isUpper {
			if idx > 0 {
				newstr = append(newstr, '_')
			}
			chr -= ('A' - 'a')
		}
		newstr = append(newstr, chr)
	}

	return string(newstr)
}

//GonicCasedName 类似驼峰式命名命名转换，但是排除一些特殊词，如ID、GUID、URL等，比如 UserID 转为user_id
func GonicCasedName(name string) string {

	for _, v := range LintGonicKeys {

		if strings.Contains(name, v) {
			name = strings.Replace(name, v, "_"+strings.ToLower(v), -1)
		}

	}

	return SnakeCasedName(name)
}
