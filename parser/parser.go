package parser

import (
	"fmt"
	"strings"
)

type Parser struct {
	text string
	tags map[string][]Parser
}

var singleTags = map[string]bool{
	"area":     true,
	"base":     true,
	"basefont": true,
	"bgsound":  true,
	"br":       true,
	"col":      true,
	"command":  true,
	"embed":    true,
	"hr":       true,
	"img":      true,
	"input":    true,
	"isindex":  true,
	"keygen":   true,
	"link":     true,
	"meta":     true,
	"param":    true,
	"source":   true,
	"track":    true,
	"wbr":      true,
	"!DOCTYPE": true,
	"!--":      true,
}

func New(text string) (Parser, error) {
	if len(text) == 0 {
		return Parser{}, nil
	}

	if strings.HasPrefix(text, "<script") {
		return Parser{text: text}, nil
	}

	type Item struct {
		tag   string
		index int
	}
	stack := make([]Item, 0)
	tags := make(map[string][]Parser)

	for i := 0; i < len(text); {
		c := text[i]
		if c != '<' {
			i++
			continue
		}

		tagEnd := strings.Index(text[i:], ">")

		if tagEnd == -1 {
			return Parser{}, fmt.Errorf("invalid text")
		}

		tag := text[i+1 : i+tagEnd]

		if len(stack) > 0 && stack[len(stack)-1].tag == "script" && tag != "/script" && tag != "script" {
			i += tagEnd + 1
			continue
		}

		if tag[0] == '!' {
			i += tagEnd + 1
			continue
		}

		if tag[0] != '/' {
			spaceInd := strings.Index(tag, " ")
			if spaceInd != -1 {
				tag = tag[:spaceInd]
			}
			if !singleTags[tag] {
				stack = append(stack, Item{tag: tag, index: i + tagEnd + 1})
			}
			i += tagEnd + 1
			continue
		}

		if singleTags[tag[1:]] {
			i += tagEnd + 1
			continue
		}

		if len(stack) == 0 {
			return Parser{}, fmt.Errorf("invalid text")
		}

		if stack[len(stack)-1].tag != tag[1:] {
			if len(stack) == 0 || stack[len(stack)-1].tag != tag[1:] {
				return Parser{}, fmt.Errorf("invalid text")
			}
		}

		if len(stack) == 1 {
			start := stack[0]
			tag, index := start.tag, start.index
			if tag == "script" {
				tags[tag] = append(tags[tag], Parser{text: text[index:i]})
			} else {
				temp, err := New(text[index:i])
				if err != nil {
					return Parser{}, err
				}
				tags[tag] = append(tags[tag], temp)
			}
		}
		stack = stack[:len(stack)-1]
		i += tagEnd + 1
	}
	if len(stack) != 0 {
		return Parser{}, fmt.Errorf("invalid text")
	}
	return Parser{text: text, tags: tags}, nil
}

func (p Parser) Find(tag string) Parser {
	if parsers, ok := p.tags[tag]; ok {
		return parsers[0]
	}
	return Parser{}
}

func (p Parser) FindAll(tag string) []Parser {
	return p.tags[tag]
}

func (p Parser) Text() string {
	return p.text
}
