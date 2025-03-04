package main 

type EventType int

const(
    TEXT EventType = iota
    DELIVERED 
    MARK_READ
    ERR
    PROFILE_INFO
    CHATPREVIEWS

    SearchUserRequest
    SearchUserResponse
)

type ServerMessage struct{
    Type EventType `json:"type"`    
    Body any        `json:"body"`
}

type Event struct{
    Type EventType  `json:"type"`
    From string     `json:"-"`
    FromID int      `json:"-"`
    Body  string `json:"body"`
}


type IncomingMessagePayload struct{
    To string `json:"to"`
    Content string `json:"content"`
    MessageMark int `json:"mark"`
    From    string `json:"-"`

    ToUserID int `json:"to_id"`
    FromUserID int `json:"-"`
}

type OutGoingMessage struct{
    From string `json:"from"`
    Content string `json:"content"`
    Timestamp int64 `json:"timestamp"`
}

type Delivered struct{
    Mark int   `json:"mark"`
    TimeStamp int64 `json:"timestamp"`
}

type SearchUserPayload struct{
    Username string `json:"username"`
    From string `json:"-"`
}

// not now baby
type MarkRead struct{
    MessageId int   `json:"message_id"`
    // the time message was read
    At int64

    To string 
    From string
}
