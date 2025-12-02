package schemas

import "time"

type UserId string
type ConversationId string
type MessageId string
type CommentedId string
type Kind string
type Status string

const (
	MessageKindNormal    Kind = "normal"
	MessageKindComment   Kind = "comment"
	MessageKindForwarded Kind = "forwarded"
	MessageKindDeleted   Kind = "deleted"

	StatusSent      Status = "sent"
	StatusDelivered Status = "delivered"
	StatusRead      Status = "seen"
	StatusDeleted   Status = "deleted"
	StatusSending   Status = "sending"
)

type User struct {
	ID        UserId    `json:"userId"`
	Name      string    `json:"name"`
	UserName  string    `json:"userName,omitempty"`
	PfpURL    *string   `json:"pfp,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type Message struct {
	ID       MessageId `json:"messageId"`
	SenderID UserId    `json:"senderId"`
	Status   Status    `json:"status"`
	Text     string    `json:"text"`
	Media    Media     `json:"media,omitempty"`
	Kind     Kind      `json:"kind"`
	SentAt   time.Time `json:"time"`
}

type Media struct {
	URL      string `json:"url"`
	MimeType string `json:"mimeType"`
	Size     int    `json:"size"`
	Filename string `json:"filename"`
}

type NotFound struct {
	Message string `json:"message"`
}

type Conversation struct {
	ID             ConversationId `json:"ConversationId"`
	ParticipantIDs []UserId       `json:"participantIds"`
	CreatedAt      time.Time      `json:"createdAt"`
	LastMessage    *Message       `json:"lastMessage,omitempty"`
	Messages       []Message      `json:"messages,omitempty"`
}
type Group struct {
	ID             ConversationId `json:"conversationId"`
	Name           string         `json:"groupName"`
	Photo          *string        `json:"groupPhoto,omitempty"`
	ParticipantIDs []UserId       `json:"participants"`
	CreatedAt      time.Time      `json:"createdAt"`
	LastMessage    *Message       `json:"lastMessage,omitempty"`
	Messages       []Message      `json:"messages,omitempty"`
}
