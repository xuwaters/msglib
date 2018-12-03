package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"unicode"
)

type MsgCompiler struct {
	Messages    []*MessageSchema
	EnumList    []*EnumSchema
	EnumMap     map[string]*EnumSchema
	ProtoName   string
	JavaPackage string

	reader *bytes.Reader
}

func NewMsgCompiler() *MsgCompiler {
	compiler := &MsgCompiler{}
	compiler.EnumMap = make(map[string]*EnumSchema)
	return compiler
}

type FieldSchema struct {
	TypeName      string
	TypeParams    []string
	FieldName     string
	FieldID       int
	Comments      []string
	SuffixComment string
}

func NewFieldSchema() *FieldSchema {
	var obj = &FieldSchema{}
	return obj
}

type EnumSchema struct {
	Name     string
	Comments []string
	Fields   []*EnumFieldSchema
}

func NewEnumSchema() *EnumSchema {
	var obj = &EnumSchema{}
	obj.Fields = make([]*EnumFieldSchema, 0)
	return obj
}

type EnumFieldSchema struct {
	FieldName     string
	FieldValue    int
	Comments      []string
	SuffixComment string
}

func NewEnumFieldSchema() *EnumFieldSchema {
	var obj = &EnumFieldSchema{}
	return obj
}

type MessageSchema struct {
	Name     string
	Comments []string
	Fields   []*FieldSchema
}

func NewMessageSchema() *MessageSchema {
	var obj = &MessageSchema{}
	obj.Fields = make([]*FieldSchema, 0)
	return obj
}

func (self *MessageSchema) GetFieldByID(id int) *FieldSchema {
	for _, field := range self.Fields {
		if field.FieldID == id {
			return field
		}
	}
	return nil
}

/////////////////////////////////////////////////////////////////////// Parsing Proto File

func (self *MsgCompiler) ParseProtoFile(protofile string) error {
	data, err := ioutil.ReadFile(protofile)
	if err != nil {
		return err
	}
	self.reader = bytes.NewReader(data)

	for !self.IsEOF() {
		comments, err := self.SkipWhiteAndReturnComments(true)
		if err != nil {
			return err
		}
		if self.IsEOF() {
			break
		}
		word := self.NextWord()
		if word == "javapackage" {
			self.JavaPackage = self.NextWord()
		} else if word == "proto" {
			self.ProtoName = self.NextWord()
		} else if word == "message" {
			msg, err := self.ParseMessage()
			if err != nil {
				return err
			}
			msg.Comments = comments
			self.Messages = append(self.Messages, msg)
		} else if word == "enum" {
			msg, err := self.ParseEnum()
			if err != nil {
				return err
			}
			msg.Comments = comments
			self.EnumList = append(self.EnumList, msg)
			self.EnumMap[msg.Name] = msg
		} else {
			return errors.New("Unknown Token: '" + word + "'")
		}
	}

	return nil
}

func (self *MsgCompiler) NextWord() string {
	self.SkipWhite()
	var sb []rune
	for IsWordChar(self.PeekChar()) {
		sb = append(sb, self.NextChar())
		if self.IsEOF() {
			break
		}
	}
	return string(sb)
}

func (self *MsgCompiler) ParseMessage() (*MessageSchema, error) {
	msg := NewMessageSchema()
	// name
	// fectch comments
	msg.Name = self.NextWord()
	if !self.MatchChar('{') {
		return nil, errors.New("Invalid message definition, expect '{' near '" + msg.Name + "'")
	}

	for {
		lastComments, err := self.SkipWhiteAndReturnComments(true)
		if err != nil {
			return nil, err
		}
		if self.MatchChar('}') {
			break
		}
		field, err := self.ParseField(msg.Name, lastComments)
		if err != nil {
			return nil, err
		}
		msg.Fields = append(msg.Fields, field)
		if self.IsEOF() {
			break
		}
	}
	return msg, nil
}

func (self *MsgCompiler) ParseEnum() (*EnumSchema, error) {
	msg := NewEnumSchema()
	msg.Name = self.NextWord()
	if !self.MatchChar('{') {
		return nil, errors.New("Invalid enum definition, expect '{' near '" + msg.Name + "'")
	}

	var (
		lastComments []string
		err          error
		field        *EnumFieldSchema
	)
	for {
		field, err = self.ParseEnumField(msg.Name, lastComments)
		if err != nil {
			return nil, err
		}
		msg.Fields = append(msg.Fields, field)
		lastComments, err = self.SkipWhiteAndReturnComments(true)
		if err != nil {
			return nil, err
		}
		if self.MatchChar('}') {
			break
		}
		if self.IsEOF() {
			break
		}
	}
	return msg, nil
}

func (self *MsgCompiler) IsEOF() bool {
	if _, _, err := self.reader.ReadRune(); err == io.EOF {
		return true
	}
	self.reader.UnreadRune()
	return false
}

func (self *MsgCompiler) PeekChar() rune {
	ch := self.NextChar()
	if ch < 0 {
		return ch
	}
	self.reader.UnreadRune()
	return ch
}

func (self *MsgCompiler) NextChar() rune {
	ch, _, err := self.reader.ReadRune()
	if err == io.EOF {
		return -1
	}
	return ch
}

func IsWordChar(c rune) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '.'
}

func IsNumberChar(c rune) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '.'
}

func (self *MsgCompiler) SkipWhite() {
	self.SkipWhiteAndReturnComments(false)
}

func (self *MsgCompiler) NextNumberWord() string {
	self.SkipWhite()
	var sb []rune
	for IsNumberChar(self.PeekChar()) {
		sb = append(sb, self.NextChar())
		if self.IsEOF() {
			break
		}
	}
	return string(sb)
}

func (self *MsgCompiler) NextInteger() (int, error) {
	self.SkipWhite()
	str := self.NextNumberWord()
	return strconv.Atoi(str)
}

func (self *MsgCompiler) MatchChar(c rune) bool {
	self.SkipWhite()
	if self.PeekChar() != c {
		return false
	}
	self.NextChar()
	return true
}

func (self *MsgCompiler) ParseEnumField(msgname string, lastComments []string) (*EnumFieldSchema, error) {
	var err error
	field := NewEnumFieldSchema()
	if lastComments == nil {
		if field.Comments, err = self.SkipWhiteAndReturnComments(true); err != nil {
			return nil, err
		}
	} else {
		field.Comments = lastComments
	}
	field.FieldName = self.NextWord()

	if !self.MatchChar('=') {
		return nil, errors.New("Expect '=' for field definition: " + field.FieldName + ", msg = " + msgname)
	}
	if field.FieldValue, err = self.NextInteger(); err != nil {
		return nil, errors.New("Expect integer for field id: " + field.FieldName + ", msg = " + msgname)
	}
	if !self.MatchChar(';') {
		return nil, errors.New("expect ';' at end of field definition, field: " + field.FieldName + ", msg = " + msgname)
	}

	if field.SuffixComment, err = self.SkipSuffixComment(); err != nil {
		return nil, err
	}

	return field, nil
}

func (self *MsgCompiler) ParseField(msgname string, lastComments []string) (*FieldSchema, error) {
	var err error
	field := NewFieldSchema()
	if lastComments == nil {
		if field.Comments, err = self.SkipWhiteAndReturnComments(true); err != nil {
			return nil, err
		}
	} else {
		field.Comments = lastComments
	}
	field.TypeName = self.NextWord()
	if field.TypeName == "list" || field.TypeName == "set" {
		if !self.MatchChar('<') {
			return nil, errors.New("should be list<T> for field " + field.TypeName + ", msg = " + msgname)
		}
		item := self.NextWord()
		if !self.MatchChar('>') {
			return nil, errors.New("should be map<K,V> for field " + field.TypeName + ", msg = " + msgname)
		}
		field.TypeParams = []string{item}
	} else if field.TypeName == "map" {
		if !self.MatchChar('<') {
			return nil, errors.New("should be map<K,V> for field " + field.TypeName + ", msg = " + msgname)
		}
		keytype := self.NextWord()
		if !self.MatchChar(',') {
			return nil, errors.New("should be map<K,V> for field " + field.TypeName + ", msg = " + msgname)
		}
		valtype := self.NextWord()
		if !self.MatchChar('>') {
			return nil, errors.New("should be map<K,V> for field " + field.TypeName + ", msg = " + msgname)
		}
		field.TypeParams = []string{keytype, valtype}
	}
	//
	field.FieldName = self.NextWord()
	if !self.MatchChar('=') {
		return nil, errors.New("Expect '=' for field definition: " + field.TypeName + ", " + field.FieldName + ", msg = " + msgname)
	}

	if field.FieldID, err = self.NextInteger(); err != nil {
		return nil, errors.New("Expect integer for field id: " + field.TypeName + " " + field.FieldName + ", msg = " + msgname)
	}
	if !self.MatchChar(';') {
		return nil, errors.New("expect ';' at end of field definition, field: " + field.TypeName +
			" " + field.FieldName + ", msg = " + msgname)
	}

	if field.SuffixComment, err = self.SkipSuffixComment(); err != nil {
		return nil, err
	}

	return field, nil
}

func (self *MsgCompiler) SkipWhiteAndReturnComments(storeComments bool) ([]string, error) {
	var comments []string = nil
	// check comment
	for {
		if self.IsEOF() {
			return comments, nil
		}
		c := self.PeekChar()

		if unicode.IsSpace(c) {
			self.NextChar()
			continue
		} else if c == '/' {
			self.NextChar()
			if self.IsEOF() {
				return comments, nil
			}
			if self.PeekChar() != '/' {
				return nil, errors.New("Comment should start with '//'")
			}
			//
			var sb []rune = nil
			if storeComments {
				sb = append(sb, '/')
			}
			for {
				if self.IsEOF() {
					if storeComments {
						comments = append(comments, string(sb))
					}
					return comments, nil
				}
				c := self.NextChar()
				if c == '\n' || c == '\r' {
					break
				}
				if storeComments {
					sb = append(sb, c)
				}
			}
			if storeComments {
				comments = append(comments, string(sb))
			}
		} else {
			break
		}
	}

	return comments, nil
}

func (self *MsgCompiler) SkipSuffixComment() (string, error) {
	for {
		if self.IsEOF() {
			return "", nil
		}
		c := self.PeekChar()
		if unicode.IsSpace(c) {
			self.NextChar()
			if c == '\r' || c == '\n' {
				break
			}
		} else if c == '/' {
			self.NextChar()
			if self.IsEOF() {
				return "", nil
			}
			if self.PeekChar() != '/' {
				return "", errors.New("Comment should start with '//'")
			}
			sb := []rune{'/'}
			for {
				if self.IsEOF() {
					return string(sb), nil
				}
				c = self.NextChar()
				if c == '\n' || c == '\r' {
					break
				}
				sb = append(sb, c)
			}
			return string(sb), nil
		} else {
			break
		}
	}
	return "", nil
}

func (self *MsgCompiler) IsEnumType(typename string) bool {
	_, ok := self.EnumMap[typename]
	return ok
}

// template helpers

func CreateTemplate(name string, allTemplates []string, funcMap template.FuncMap) (*template.Template, error) {
	var (
		tmpl *template.Template
		err  error
	)

	leading := regexp.MustCompile("(\n)?[ \t]*[{]{2}[-][ ]*")
	trailing := regexp.MustCompile("[ ]*[-][}]{2}[ \t]*(\n)?")

	tmpl = template.New(name).Funcs(funcMap)
	for _, code := range allTemplates {
		code = leading.ReplaceAllString(code, "{{")
		code = trailing.ReplaceAllString(code, "}}")
		tmpl, err = tmpl.Parse(code)
		if err != nil {
			return nil, err
		}
	}

	return tmpl, nil
}

func findPosition(haystack []string, needle string) int {
	for i, str := range haystack {
		if str == needle {
			return i
		}
	}
	return -1
}

func ParseArgsFromErrMessage(val string) []string {
	args := make([]string, 0)
	runelist := []rune(val)
	runelen := len(runelist)
	for s := 0; s < runelen-1; {
		if runelist[s] == '$' && runelist[s+1] == '{' {
			e := s + 2
			for ; e < runelen; e++ {
				if runelist[e] == '}' {
					curr := string(runelist[s+2 : e])
					if findPosition(args, curr) < 0 {
						args = append(args, curr)
					}
					e = e + 1
					break
				}
			}
			s = e
		} else {
			s = s + 1
		}
	}
	return args
}

func trimComment(comment string) string {
	if len(comment) < 2 {
		return comment
	}
	return strings.TrimSpace(comment[2:])
}
