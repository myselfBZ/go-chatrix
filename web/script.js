// Dear future me, i cannot express how sorry i am for writing this piece of sh*t


const addr = "http://localhost:6969/ws"
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
    pendingMessages:{},
    unreadMessages:[],
    //...
}


let pendingMessages = {};
let unreadMessages = {};

const EVENT = {
    Text:0,
    Delivered:1,
    // MarkRead:2,
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

    const viewableMsg = document.createElement('p')
    viewableMsg.innerText = msg
    viewableMsg.className = "users-msgs"
    pendingMessages[mark] = viewableMsg
    document.getElementById('msgs').appendChild(viewableMsg)

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
                const msgs = document.getElementById('msgs')
                const incomingMsg = document.createElement('p')
                incomingMsg.textContent = `${data.body.from}:  ${data.body.content}`
                msgs.appendChild(incomingMsg)
                scrollToBottom()

                // send read Event!

                break;
            case EVENT.ProfileInfo:
                state.user = data.body
                const h1 = document.getElementById('welcome')
                h1.innerHTML += ` ${state.user.name}`
                break
            case EVENT.Delivered:
                const htmlElement = pendingMessages[data.body.mark]
                const msgStatus = document.createElement('small')
                msgStatus.textContent = " ✅"
                htmlElement.appendChild(msgStatus)
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

                        if(textingTo === null){
                            const newTextingTo = document.createElement('p')
                            newTextingTo.textContent = `You are texting ${state.to.username}`
                            newTextingTo.id = 'texting-to'
                            document.body.appendChild(newTextingTo)
                            handleLoadingChatHistory(i.id)
                            return
                        }
                        textingTo.textContent = `You are texting ${state.to.username}`
                        handleLoadingChatHistory(i.id)

                    })
                    console.log(i)
                    chatCard.textContent = i.username
                    chatCard.className = "chat-cards"
                    div.appendChild(chatCard)
                }
                break;
                case EVENT.SearchByUsernameResponse:
                const container = document.getElementById('searched-users')
                if(container.innerHTML != "") container.innerHTML = "";
                for (user of data.body){
                    const result = document.createElement('p')
                    result.textContent = user.username

                    result.addEventListener('click', () => {
                        if(state.to.username === result.textContent) return;
                        
                        state.to.username = result.textContent
                        
                        const textingTo = document.getElementById('texting-to')
                        if(textingTo === null) {
                            const newTextingTo = document.createElement('p')
                            newTextingTo.id = 'texting-to'
                            newTextingTo.textContent = `You are texting ${result.textContent}`
                            document.body.appendChild(newTextingTo)
                            return
                        }

                        textingTo.textContent =   `You are texting to ${result.textContent}`
                    })
                    
                    container.appendChild(result)
                }
                break;
                case EVENT.LoadChatHistoryResponse:
                    const msgContainer = document.getElementById('msgs')
                    msgContainer.innerHTML = ""
                    data.body.sort((a, b) => { new Date(a.created_at) - new Date(b.created_at)})
                    for(const msg of data.body){
                        const incomingMsg = document.createElement('p')
                        incomingMsg.textContent = `${msg.content}`
                        if(msg.user_id === state.user.id){
                            incomingMsg.textContent = `${msg.content}`
                            incomingMsg.className = "users-msgs"
                        }
                        

                        msgContainer.appendChild(incomingMsg)
                    }
        }
    }
}

init()

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
    console.log("Hello??");
    
    ws.send(JSON.stringify(event))
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
