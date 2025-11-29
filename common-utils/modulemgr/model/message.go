// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package model to start module_manager model
package model

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// Not good, just for compatible with both mef msg and kubeede msg
type header struct {
	Id              string      `json:"id"`
	ID              string      `json:"msg_id"` // kubeedge format
	ParentId        string      `json:"parentId"`
	ParentID        string      `json:"parent_msg_id"` // kubeedge format
	IsSync          bool        `json:"isSync"`
	Sync            bool        `json:"sync"` // kubeedge format
	Timestamp       int64       `json:"timestamp"`
	Version         string      `json:"version"`
	ResourceVersion string      `json:"resourceversion"` // kubeedge format
	NodeId          string      `json:"nodeId"`
	PeerInfo        MsgPeerInfo `json:"peerInfo"`
}

type router struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Option      string `json:"option"`
	Resource    string `json:"resource"`
}

// MessageRoute kubeedge message route
type MessageRoute struct {
	Source    string `json:"source"`    // kubeedge format
	Group     string `json:"group"`     // kubeedge format
	Operation string `json:"operation"` // kubeedge format
	Resource  string `json:"resource"`  // kubeedge format
}

// Message struct
type Message struct {
	Header         header       `json:"header"`
	Router         router       `json:"router"`
	KubeEdgeRouter MessageRoute `json:"route"`
	Content        RawMessage   `json:"content"`
}

// RawMessage is the type of content used for message passing
// It realizes Marshaler and Unmashaler interface.
type RawMessage []byte

// MarshalJSON returns rawmessage itself if it fits the json-format.
// Otherwise, it will convert it into json-formatted string.
func (m RawMessage) MarshalJSON() ([]byte, error) {
	if json.Valid(m) {
		return m, nil
	}

	return FormatMsg(m), nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *RawMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("unmarshalJSON on nil pointer")
	}

	*m = append((*m)[0:0], data...)

	return nil
}

// FormatMsg is used to change a byte slice into json-format
func FormatMsg(content []byte) []byte {
	return []byte(strconv.QuoteToASCII(string(content)))
}

// MsgPeerInfo is the struct to save peer info in a msg
type MsgPeerInfo struct {
	Ip string `json:"ip,omitempty"`
	Sn string `json:"serialNumber,omitempty"`
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

// GetPeerInfo get the peer info from msg
func (msg *Message) GetPeerInfo() MsgPeerInfo {
	return msg.Header.PeerInfo
}

// SetPeerInfo set the peer info into msg
func (msg *Message) SetPeerInfo(peerInfo MsgPeerInfo) {
	msg.Header.PeerInfo = peerInfo
}

// SetRouter set the message router
func (msg *Message) SetRouter(source, destination, option, resource string) {
	msg.Router.Source = source
	msg.Router.Destination = destination
	msg.Router.Option = option
	msg.Router.Resource = resource
}

// SetKubeEdgeRouter set kubeedge message router
func (msg *Message) SetKubeEdgeRouter(source, group, operation, resource string) {
	msg.KubeEdgeRouter.Source = source
	msg.KubeEdgeRouter.Group = group
	msg.KubeEdgeRouter.Operation = operation
	msg.KubeEdgeRouter.Resource = resource
}

// ParseContent parse the message content to a specified type
// v must be addressable
func (msg *Message) ParseContent(v interface{}) error {
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if !value.CanSet() {
		return errors.New("value is not addressable")
	}

	switch value.Kind() {
	case reflect.Slice:
		if value.Type().Elem().Kind() == reflect.Uint8 {
			value.SetBytes(UnformatMsg(msg.Content))
			return nil
		}
		return json.Unmarshal(UnformatMsg(msg.Content), v)
	case reflect.String:
		value.SetString(string(UnformatMsg(msg.Content)))
		return nil
	default:
		return json.Unmarshal(UnformatMsg(msg.Content), v)
	}
}

// UnformatMsg return origin data if the data is not a JSON-formatted string.
// Otherwise, it returns the string after deserializing it from JSON format
func UnformatMsg(data []byte) []byte {
	if len(data) <= 1 {
		return data
	}

	if data[0] != '"' || data[len(data)-1] != '"' {
		return data
	}

	var strRet string
	err := json.Unmarshal(data, &strRet)
	if err != nil {
		return []byte{}
	}
	return []byte(strRet)
}

// FillContent fill the message content
// For the content of type []byte, it is directly filled in.
// For the content of type string, it is converted into []byte and then filled in.
// For other types, they are serialized and then filled in.
// The param transferStructIntoStr is an optional parameter that indicates whether the struct needs to be serialized
// twice. It converts a structure into a final Json-formatted string.
// It mainly used for integrating with certain edgecore/FD messages.
// eg: {"name":"someone"} is the final content if the input is a struct and  transferStructIntoStr is false.
//     "{\"name\":\"someone\"}" is the final content if the input is a struct and  transferStructIntoStr is true.
func (msg *Message) FillContent(content interface{}, transferStructIntoStr ...bool) error {
	if bytes, ok := content.([]byte); ok {
		msg.Content = bytes
		return nil
	}

	if str, ok := content.(string); ok {
		msg.Content = RawMessage(str)
		return nil
	}

	marshaledContent, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("marshal content failed: %v", err)
	}
	msg.Content = marshaledContent

	if reflect.TypeOf(content) == nil {
		return fmt.Errorf("content is nil interface")
	}

	if len(transferStructIntoStr) == 0 || !transferStructIntoStr[0] || reflect.TypeOf(content).Kind() != reflect.
		Struct {
		return nil
	}

	marshaledStrContent, err := json.Marshal(string(marshaledContent))
	if err != nil {
		return fmt.Errorf("marshal transformed content failed: %v", err)
	}
	msg.Content = marshaledStrContent

	return nil
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
	respMsg.KubeEdgeRouter.Group = msg.KubeEdgeRouter.Group

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
