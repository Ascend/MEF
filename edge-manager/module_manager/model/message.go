package model

type header struct {
	id       string
	parentId string
	isSync   bool
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

func (msg *Message) GetSource() string {
	return msg.router.source
}

func (msg *Message) GetDestination() string {
	return msg.router.destination
}

func (msg *Message) GetId() string {
	return msg.header.id
}

func (msg *Message) SetParentId(parentId string) {
	msg.header.parentId = parentId
}

func (msg *Message) GetParentId() string {
	return msg.header.parentId
}

