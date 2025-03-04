# Events 
----------
Client should send
Event {
    Type: 0 (0 == Text)
    Body: "
    {
        "to":"username",
        "mark":int
        "content":"message"
    }
    "
}

Server Forwards

ServerMessage {
    Type: 0 (0 == Text)
    Body:{
        from:"username",
        content:"message"
        timestamp:time since epoch
    }
}

