package schemas

import "time"

// Type definitions for IDs
type UserId string
type ConversationId string
type MessageId string

// Enum constants
const (
	// Message Kind
	MsgKindNormal    = "normal"
	MsgKindForwarded = "forwarded"

	// Message Status
	MsgStatusSent      = "sent"
	MsgStatusDelivered = "delivered"
	MsgStatusSeen      = "seen"
	MsgStatusDeleted   = "deleted"

	// Conversation Type
	ConvTypePrivate = "private"
	ConvTypeGroup   = "group"
)

// schemas/User
type User struct {
	ID        UserId    `json:"userId"`
	Name      string    `json:"userName"`
	PFPURL    string    `json:"pfp,omitempty"` // Profile Picture URL
	CreatedAt time.Time `json:"createdAt"`
}

// /schemas/Media
type Media struct {
	URL      string `json:"url"`
	Filename string `json:"filename,omitempty"`
	MimeType string `json:"mimeType"`
	Size     int    `json:"size,omitempty"`
}

// /schemas/Reaction
type Reaction struct {
	UserId UserId `json:"userId"`
	Emoji  string `json:"emoji"`
}

// /schemas/Message
type Message struct {
	ID        MessageId  `json:"messageId"`
	Sender    UserId     `json:"sender"`
	Status    string     `json:"status"`
	Kind      string     `json:"kind"`
	Time      time.Time  `json:"time"`
	Text      string     `json:"text,omitempty"`
	MediaId   string     `json:"mediaId,omitempty"`
	Media     *Media     `json:"media,omitempty"`
	Reactions []Reaction `json:"reactions"`
	ReplyToId *MessageId `json:"replyToId,omitempty"`
}

// /schemas/Conversation

type PrivateConversation struct {
	ConvId       ConversationId `json:"convId"`
	Type         string         `json:"type"`
	LastMessage  *Message       `json:"lastMessage,omitempty"`
	UnreadCount  int            `json:"unreadCount"`
	Participants []UserId       `json:"participants"`
}

type Group struct {
	ConvId       ConversationId `json:"convId"`
	Type         string         `json:"type"`
	LastMessage  *Message       `json:"lastMessage,omitempty"`
	UnreadCount  int            `json:"unreadCount"`
	Participants []UserId       `json:"participants"`
	GroupName    string         `json:"groupName"`
	GroupPhoto   string         `json:"groupPhoto,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
}

type GenericResponse struct {
	Message string `json:"message"`
}
