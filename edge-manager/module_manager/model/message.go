package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type header struct {
	id        string
	parentId  string
	isSync    bool
	timestamp int64
}

type router struct {
	source      string
	destination string
	option      string
	resource    string
}

type Message struct {
	header  header
	router  router
	content interface{}
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
const messageIdVariantMask = 0xff

const messgeIdTimestampMask = 1e6

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
	g.buffer[messageIdVersionIndex] = (g.buffer[messageIdVersionIndex] & messageIdVersionMask) | (messageIdVersion << 4)
}

func (g *messageIdGenerator) setVariant() {
	g.buffer[messageIdVariantIndex] = g.buffer[messageIdVariantIndex]&(messageIdVariantMask>>2) | (0x02 << 6)
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
	return msg.header.id
}

// GetParentId get message parent id
func (msg *Message) GetParentId() string {
	return msg.header.parentId
}

// SetParentId set message parent id
func (msg *Message) SetParentId(parentId string) {
	msg.header.parentId = parentId
}

// GetIsSync get message to is sync or not
func (msg *Message) GetIsSync() bool {
	return msg.header.isSync
}

// SetIsSync set message to is sync or not
func (msg *Message) SetIsSync(isSync bool) {
	msg.header.isSync = isSync
}

// GetTimestamp get message timestamp
func (msg *Message) GetTimestamp() int64 {
	return msg.header.timestamp
}

// GetSource get message source
func (msg *Message) GetSource() string {
	return msg.router.source
}

// GetDestination get message destination
func (msg *Message) GetDestination() string {
	return msg.router.destination
}

// GetOption get message option
func (msg *Message) GetOption() string {
	return msg.router.option
}

// GetResource get message resource
func (msg *Message) GetResource() string {
	return msg.router.resource
}

// SetRouter set the message router
func (msg *Message) SetRouter(source, destination, option, resource string) {
	msg.router.source = source
	msg.router.destination = destination
	msg.router.option = option
	msg.router.resource = resource
}

// GetContent get the message content
func (msg *Message) GetContent() interface{} {
	return msg.content
}

// FillContent fill the message content
func (msg *Message) FillContent(content interface{}) {
	msg.content = content
}

// NewResponse create new inner respone
func (msg *Message) NewResponse() (*Message, error) {
	var respMsg *Message
	var err error

	if respMsg, err = NewMessage(); err != nil || respMsg == nil {
		return nil, fmt.Errorf("create new message fail when create new response")
	}
	respMsg.header.parentId = msg.header.id
	respMsg.header.isSync = msg.header.isSync

	respMsg.router.source = msg.router.destination
	respMsg.router.destination = msg.router.source
	respMsg.router.resource = msg.router.resource

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
	msg.header.id = msgId
	msg.header.timestamp = time.Now().UnixNano() / messgeIdTimestampMask

	return &msg, nil
}
