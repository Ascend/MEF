// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package model to start module_manager model
package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type header struct {
	Id        string `json:"id"`
	ParentId  string `json:"parentId"`
	IsSync    bool   `json:"isSync"`
	Timestamp int64  `json:"timestamp"`
	Version   string `json:"version"`
	NodeId    string `json:"nodeId"`
}

type router struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Option      string `json:"option"`
	Resource    string `json:"resource"`
}

// Message struct
type Message struct {
	Header  header      `json:"header"`
	Router  router      `json:"router"`
	Content interface{} `json:"content"`
}

const messageIdSize = 16
const messageIdVersion = 4
const messageIdStringBufferSize = 36
const messageIdVersionIndex = 6
const messageIdVariantIndex = 8

const messageIdHexFirstPartBegin = 0
const messageIdHexFirstPartEnd = 8
const messageIdHexSecondPartBegin = 9
const messageIdHexSecondPartEnd = 13
const messageIdHexThirdPartBegin = 14
const messageIdHexThirdPartEnd = 18
const messageIdHexFourthPartBegin = 19
const messageIdHexFourthPartEnd = 23
const messageIdHexFifthPartBegin = 24

const messageIdRandomFirstPartBegin = 0
const messageIdRandomSecondPartBegin = 4
const messageIdRandomThirdPartBegin = 6
const messageIdRandomFourthPartBegin = 8
const messageIdRandomFifthPartBegin = 10

const messageIdVersionMask = 0x0f
const messageIdVersionOffset = 4
const messageIdVariantMask = 0xff
const messageIdVariantMaskOffset = 2
const messageIdVariantConst = 0x02
const messageIdVariantConstOffset = 6

const messageIdTimestampMask = 1e6

type messageIdGenerator struct {
	buffer [messageIdSize]byte
}

func (g *messageIdGenerator) random() error {
	if _, err := rand.Read(g.buffer[:]); err != nil {
		return err
	}
	return nil
}

func (g *messageIdGenerator) setVersion() {
	g.buffer[messageIdVersionIndex] = (g.buffer[messageIdVersionIndex] & messageIdVersionMask) |
		(messageIdVersion << messageIdVersionOffset)
}

func (g *messageIdGenerator) setVariant() {
	g.buffer[messageIdVariantIndex] =
		g.buffer[messageIdVariantIndex]&(messageIdVariantMask>>messageIdVariantMaskOffset) |
			(messageIdVariantConst << messageIdVariantConstOffset)
}

func (g *messageIdGenerator) toString() string {
	buf := make([]byte, messageIdStringBufferSize)

	hex.Encode(buf[messageIdHexFirstPartBegin:messageIdHexFirstPartEnd],
		g.buffer[messageIdRandomFirstPartBegin:messageIdRandomSecondPartBegin])
	buf[messageIdHexFirstPartEnd] = '-'
	hex.Encode(buf[messageIdHexSecondPartBegin:messageIdHexSecondPartEnd],
		g.buffer[messageIdRandomSecondPartBegin:messageIdRandomThirdPartBegin])
	buf[messageIdHexSecondPartEnd] = '-'
	hex.Encode(buf[messageIdHexThirdPartBegin:messageIdHexThirdPartEnd],
		g.buffer[messageIdRandomThirdPartBegin:messageIdRandomFourthPartBegin])
	buf[messageIdHexThirdPartEnd] = '-'
	hex.Encode(buf[messageIdHexFourthPartBegin:messageIdHexFourthPartEnd],
		g.buffer[messageIdRandomFourthPartBegin:messageIdRandomFifthPartBegin])
	buf[messageIdHexFourthPartEnd] = '-'
	hex.Encode(buf[messageIdHexFifthPartBegin:], g.buffer[messageIdRandomFifthPartBegin:])

	return string(buf)
}

// String get message id string
func (g *messageIdGenerator) String() (string, error) {
	if err := g.random(); err != nil {
		return "", err
	}
	g.setVersion()
	g.setVariant()
	return g.toString(), nil
}

// GetId get message id
func (msg *Message) GetId() string {
	return msg.Header.Id
}

// GetNodeId get message node id
func (msg *Message) GetNodeId() string {
	return msg.Header.NodeId
}

// SetNodeId set message node id
func (msg *Message) SetNodeId(nodeId string) {
	msg.Header.NodeId = nodeId
}

// GetParentId get message parent id
func (msg *Message) GetParentId() string {
	return msg.Header.ParentId
}

// SetParentId set message parent id
func (msg *Message) SetParentId(parentId string) {
	msg.Header.ParentId = parentId
}

// GetIsSync get message to is sync or not
func (msg *Message) GetIsSync() bool {
	return msg.Header.IsSync
}

// SetIsSync set message to is sync or not
func (msg *Message) SetIsSync(isSync bool) {
	msg.Header.IsSync = isSync
}

// GetTimestamp get message timestamp
func (msg *Message) GetTimestamp() int64 {
	return msg.Header.Timestamp
}

// GetSource get message source
func (msg *Message) GetSource() string {
	return msg.Router.Source
}

// GetDestination get message destination
func (msg *Message) GetDestination() string {
	return msg.Router.Destination
}

// GetOption get message option
func (msg *Message) GetOption() string {
	return msg.Router.Option
}

// GetResource get message resource
func (msg *Message) GetResource() string {
	return msg.Router.Resource
}

// SetRouter set the message router
func (msg *Message) SetRouter(source, destination, option, resource string) {
	msg.Router.Source = source
	msg.Router.Destination = destination
	msg.Router.Option = option
	msg.Router.Resource = resource
}

// GetContent get the message content
func (msg *Message) GetContent() interface{} {
	return msg.Content
}

// FillContent fill the message content
func (msg *Message) FillContent(content interface{}) {
	msg.Content = content
}

// NewResponse create new inner response
func (msg *Message) NewResponse() (*Message, error) {
	var respMsg *Message
	var err error

	if respMsg, err = NewMessage(); err != nil || respMsg == nil {
		return nil, fmt.Errorf("create new message fail when create new response")
	}
	respMsg.Header.ParentId = msg.Header.Id
	respMsg.Header.IsSync = msg.Header.IsSync

	respMsg.Router.Source = msg.Router.Destination
	respMsg.Router.Destination = msg.Router.Source
	respMsg.Router.Option = msg.Router.Option
	respMsg.Router.Resource = msg.Router.Resource

	return respMsg, nil
}

// NewMessage create new inner message
func NewMessage() (*Message, error) {
	var msgId string
	var err error
	generator := messageIdGenerator{}

	if msgId, err = generator.String(); err != nil {
		return nil, err
	}

	msg := Message{}
	msg.Header.Id = msgId
	msg.Header.Timestamp = time.Now().UnixNano() / messageIdTimestampMask
	msg.Header.Version = "1.0"

	return &msg, nil
}
