const addr = "http://localhost:6969/ws"
const ws = new WebSocket(addr)

let state = {
    to:{
        username:"",
        id:null,
    },
    user:{},
    pendingMessages:{},
    unreadMessages:[],
    //...
}


let user;

let pendingMessages = {};
let unreadMessages = {};

const EVENT = {
    Text:0,
    Delivered:1,
    Err:3,
    ProfileInfo: 4,
    Chatprview:5,
    SearchByUsername:6,
    SearchByUsernameResponse:7
}

function handleSend(){
    if(state.to == ""){
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
    viewableMsg.innerText = "You: " + msg
    
    pendingMessages[mark] = viewableMsg
    document.getElementById('msgs').appendChild(viewableMsg)

    document.getElementById("msg").value = ""
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

                // send read Event!

                break;
            case EVENT.ProfileInfo:
                user = data.body
                const h1 = document.getElementById('welcome')
                h1.innerHTML += ` ${user.name}`
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
                for(i of data.body){
                    const chatCard = document.createElement('h3')
                    
                    chatCard.addEventListener("click", () => {
                        if (state.to == chatCard.textContent) return;
                        state.to.username = chatCard.textContent
                        state.to.id = i.id
                        const textingTo = document.createElement('p')
                        textingTo.textContent = `You are texting ${state.to.username}`
                        document.body.appendChild(textingTo)
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
                    // HASH T MEL
                    const result = document.createElement('p')
                    result.textContent = user.username

                    result.addEventListener('click', () => {
                        if (state.to == result.textContent) return;
                        state.to.username = result.textContent
                        state.to.id = user.id
                        const textingTo = document.createElement('p')
                        textingTo.textContent = `You are texting ${state.to.username}`
                        document.body.appendChild(textingTo)
                    })
                    
                    container.appendChild(result)
                }
                break
        }
    }
}

init()

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
