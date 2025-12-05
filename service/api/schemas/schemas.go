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
	StatusSeen      Status = "seen"
	StatusDeleted   Status = "deleted"
	StatusSending   Status = "sending"
)

type User struct {
	UserId    UserId    `json:"userId"`
	UserName  string    `json:"userName"`
	PFPURL    *string   `json:"pfp,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type Message struct {
	MessageId MessageId `json:"messageId"`
	Sender    UserId    `json:"sender"`
	Status    Status    `json:"status"`
	Kind      Kind      `json:"kind"`
	Time      time.Time `json:"time"`
	Text      *string   `json:"text,omitempty"`
	Media     *Media    `json:"media,omitempty"`
}

type Comment struct {
	Message
}

// ForwardedMessage implementa lo schema ForwardedMessage (usa allOf: [Message])
// Embedding per ereditare tutti i campi di Message.
type ForwardedMessage struct {
	Message
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
	ID             ConversationId `json:"convId"`
	ParticipantIDs []UserId       `json:"participants"`
	LastMessage    *Message       `json:"lastMessage,omitempty"`
	Messages       []Message      `json:"messages,omitempty"`
	Kind           string
}
type Group struct {
	Conversation
	Name  string  `json:"groupName"`
	Photo *string `json:"groupPhoto,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
}
