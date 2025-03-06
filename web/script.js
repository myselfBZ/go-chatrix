// Dear future me, i cannot express how sorry i am for writing this piece of sh*t


const addr = "http://192.168.1.22:6969/ws"
const ws = new WebSocket(addr)

let state = {
    to:{
        username:"",
        id:null,
    },
    user:{
        username:"",
        id:null,
    },
    //...
}


let pendingMessages = {};


const EVENT = {
    Text:0,
    Delivered:1,
    MarkRead:2,
    Err:3,
    ProfileInfo: 4,
    Chatprview:5,
    SearchByUsername:6,
    SearchByUsernameResponse:7,

    LoadChatHistoryRequest:8,
    LoadChatHistoryResponse:9
}

function handleSend(){
    if(state.to.username === "" || state.to.username === state.user.username){
        const errorMessage = document.createElement('p')
        errorMessage.style.color = 'red'
        errorMessage.textContent = "first choose someone to text okay?"
        
        document.body.appendChild(errorMessage)
        setTimeout(() => {
            document.body.removeChild(errorMessage)
        }, 3000)
        document.body.scrollTop = document.body.scrollHeight
        return
    }
    const msg = document.getElementById("msg").value
    const mark = Date.now()
    const event = {
        type:EVENT.Text,
        body:JSON.stringify(
            {
                to:state.to.username,
                to_id:state.to.id,
                content:msg,
                mark:mark
            }
        )
    }
    ws.send(JSON.stringify(event))

    const viewableMsgContainer = document.createElement('div')
    const viewableMsg = document.createElement('p')
    const stateOfMsg = document.createElement('small')
    stateOfMsg.innerText = "sending..."

    viewableMsg.innerText = msg
    viewableMsg.className = "users-msgs"
    
    viewableMsgContainer.appendChild(viewableMsg)
    viewableMsgContainer.appendChild(stateOfMsg)

    pendingMessages[mark] = viewableMsgContainer
    document.getElementById('msgs').appendChild(viewableMsgContainer)

    document.getElementById("msg").value = ""
    scrollToBottom()
}

function init(){
    const token = localStorage.getItem('token')
    if (token == null) {
        window.location.href = "/login.html"
    }
    ws.onclose = () => {
        console.log("disconected")
    }
    ws.onopen = () => {
        ws.send(JSON.stringify({token}))
    }
    ws.onmessage = (e) => {
        const data = JSON.parse(e.data)
        
        switch (data.type){
            case EVENT.Text:
                // UI shit
                if(data.body.from !== state.to.username) {
                    // this means that if the user sending this message isn't
                    // the user that current user is having conversation with
                    // then users just gets notified
                    alert(`you got a new message from ${data.body.from}`)
                    return
                };
                const msgs = document.getElementById('msgs')
                const incomingMsg = document.createElement('p')
                incomingMsg.textContent = `${data.body.content}`
                incomingMsg.className = "others-msgs"
                msgs.appendChild(incomingMsg)
                console.log("i got a message with id: ", data.body.msg_id)
                readMessages([data.body.msg_id])
                scrollToBottom()

                break;


            case EVENT.ProfileInfo:
                state.user = data.body
                const h1 = document.getElementById('welcome')
                h1.innerHTML += ` ${state.user.name}`
                break


            case EVENT.Delivered:
                const htmlElement = pendingMessages[data.body.mark]
                const msgState = htmlElement.getElementsByTagName('small')[0]
                msgState.innerText = "sent"                
                msgState.id = data.body.message_id
                
                
                delete pendingMessages[data.body.mark]
                break


            case EVENT.Chatprview:

                const div = document.getElementById('chats')
                for(const i of data.body){
                    const chatCard = document.createElement('h3')
                    
                    chatCard.addEventListener("click", () => {
                        if (state.to.username === chatCard.textContent) return;
                        state.to.username = chatCard.textContent
                        state.to.id = i.id
                        
                        const textingTo = document.getElementById('texting-to')
                        textingTo.textContent = `${state.to.username}`
                        handleLoadingChatHistory(i.id)
                        showMessageInput()

                    })
                    
                    chatCard.textContent = i.username
                    chatCard.className = "chat-cards"
                    div.appendChild(chatCard)
                }
                break;


                case EVENT.SearchByUsernameResponse:

                const container = document.getElementById('searched-users')
                if(container.innerHTML != "") container.innerHTML = "";
                for (const user of data.body){
                    const result = document.createElement('p')

                    result.textContent = user.username

                    result.addEventListener('click', () => {
                        if(state.to.username === result.textContent) return;
                        
                        state.to.username = result.textContent
                        state.to.id = user.id
                        const textingTo = document.getElementById('texting-to')

                        textingTo.textContent =   `${result.textContent}`
                        
                        
                        handleLoadingChatHistory(user.id)
                        showMessageInput()
                    })
                    
                    container.appendChild(result)
                }
                break;


                case EVENT.LoadChatHistoryResponse:
                    const msgContainer = document.getElementById('msgs')
                    msgContainer.innerHTML = ""
                    
                    if(data.body === null) {
                        return
                    };

                    data.body.sort((a, b) => { new Date(a.created_at) - new Date(b.created_at)})
                    

                    
                    let readNewMessages = [];

                    for(const msg of data.body){
                        if(!msg.read){
                            readNewMessages.push(msg.id)
                        }
                        if(msg.user_id === state.user.id){
                            console.log(msg.read)
                            
                            const viewableMsgContainer = document.createElement('div')
                            const viewableMsg = document.createElement('p')
                            const stateOfMsg = document.createElement('small')
                            
                            viewableMsg.innerText = msg.content
                            viewableMsg.className = "users-msgs"
                            stateOfMsg.id = msg.id      
                            viewableMsgContainer.appendChild(viewableMsg)
                            viewableMsgContainer.appendChild(stateOfMsg)
                            msgContainer.appendChild(viewableMsgContainer)
                            if(msg.read){
                                stateOfMsg.innerText = "read"
                            }
                            continue
                            
                        }
                        const incomingMsg = document.createElement('p')
                        incomingMsg.id = msg.id
                        incomingMsg.textContent = `${msg.content}`
                        incomingMsg.className = "others-msgs"
                        

                        msgContainer.appendChild(incomingMsg)
                    }
                    readMessages(readNewMessages)
                    break;
                case EVENT.MarkRead:
                    for(const id of data.body){
                        
                        
                        const msg = document.getElementById(id.toString())
                        msg.innerText = "read"
                    }
                    break;
        }
    }
}

init()

function readMessages(msgsArr){
    const jsonData = JSON.stringify({
        type:EVENT.MarkRead,
        body:JSON.stringify({
            to:state.to.username,
            message_ids:msgsArr
        })
    })
    ws.send(jsonData)
}

function scrollToBottom() {
    const msgContainer = document.getElementById('msgs');
    msgContainer.scrollTop = msgContainer.scrollHeight;
}

function handleLoadingChatHistory(userId){
    const event = {
        type:EVENT.LoadChatHistoryRequest,
        body:JSON.stringify(
            {
                user1_id:state.user.id,
                user2_id:userId
            }
        )
    }
    
    ws.send(JSON.stringify(event))
}

function showMessageInput(){
    const container = document.getElementById('message-input')
    if(container.innerHTML !== "") return;
    const input = document.createElement('input')
    input.type = "text"
    input.id = "msg"
    input.placeholder = "Message"
    
    const sendButton = document.createElement('button')
    sendButton.textContent = "send"
    sendButton.addEventListener('click', handleSend)
    container.appendChild(input)
    container.appendChild(sendButton)
}


function handleSearch(){
    const username = document.getElementById('search').value
    if(username == "") return;
    const request = JSON.stringify({
        type:EVENT.SearchByUsername,
        body:JSON.stringify({
            username: username
        })
    })
    ws.send(request)
}
