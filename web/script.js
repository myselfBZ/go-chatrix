// Dear future me, i cannot express how sorry i am for writing this piece of sh*t
import { SERVERADDR } from "./constants.js"


const svgDelivered = `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#bbb" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
  <path d="M5 12l5 5L20 7"></path>
</svg>
`
const svgRead = `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#34c759" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
  <path d="M5 13l4 4L19 7"></path>
  <path d="M11 13l4 4L23 7"></path>
</svg>
`

const svgSending = `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#bbb" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
  <circle cx="12" cy="12" r="10"></circle>
  <path d="M12 6v6l4 2"></path>
</svg>
`

const addr = `${SERVERADDR}/ws`
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
        body:
            {
                to:state.to.username,
                to_id:state.to.id,
                content:msg,
                mark:mark
            }
    }
    ws.send(JSON.stringify(event))

    const viewableMsgContainer = document.createElement('div')
    const viewableMsg = document.createElement('p')
    const stateOfMsg = document.createElement('small')
    stateOfMsg.innerHTML =  svgSending

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
    document.getElementById('search').addEventListener('change', handleSearch)
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
                msgState.innerHTML = svgDelivered                
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
                        if(!msg.read && msg.user_id !== state.user.id){
                            readNewMessages.push(msg.id)
                        }
                        if(msg.user_id === state.user.id){
                            
                            
                            const viewableMsgContainer = document.createElement('div')
                            const viewableMsg = document.createElement('p')
                            const stateOfMsg = document.createElement('small')
                            
                            viewableMsg.innerText = msg.content
                            viewableMsg.className = "users-msgs"
                            stateOfMsg.id = msg.id    
                            if(!msg.read){
                                stateOfMsg.innerHTML = svgDelivered
                            }  
                            viewableMsgContainer.appendChild(viewableMsg)
                            viewableMsgContainer.appendChild(stateOfMsg)
                            msgContainer.appendChild(viewableMsgContainer)
                            if(msg.read){
                                stateOfMsg.innerHTML = svgRead
                            }
                            scrollToBottom()
                            continue
                            
                        }
                        const incomingMsg = document.createElement('p')
                        incomingMsg.id = msg.id
                        incomingMsg.textContent = `${msg.content}`
                        incomingMsg.className = "others-msgs"
                        

                        msgContainer.appendChild(incomingMsg)
                        scrollToBottom()
                    }
                    if(readMessages.length > 0){
                        console.log("i am sending messages...")
                        readMessages(readNewMessages)
                    }
                    break;
                case EVENT.MarkRead:
                    for(const id of data.body){
                        
                        
                        const msg = document.getElementById(id.toString())
                        msg.innerHTML = svgRead
                    }
                    break;
        }
    }
}

init()

function readMessages(msgsArr){
    const jsonData = JSON.stringify({
        type:EVENT.MarkRead,
        body:{
            to:state.to.username,
            message_ids:msgsArr
        }
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
        body:
            {
                user1_id:state.user.id,
                user2_id:userId
            }
       
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

    const emojiButton = document.createElement('button')
    emojiButton.innerText = "🙂"
    emojiButton.id = "emoji-toggler"
    sendButton.addEventListener('click', handleSend)
    emojiButton.addEventListener('click', handleEmojiToggle)
    container.appendChild(input)
    container.appendChild(sendButton)
    container.appendChild(emojiButton)
}

function handleEmojiToggle() {
    let pickerContainer = document.getElementById('picker-container');

    if (!pickerContainer) {
        // Create container
        pickerContainer = document.createElement('div');
        pickerContainer.id = 'picker-container';
        
        // Create emoji picker
        const picker = document.createElement('emoji-picker');
        picker.id = 'emoji-picker';
        picker.addEventListener("emoji-click", (e) => {
            const msgInput = document.getElementById('msg'); // Ensure input field ID is correct
            if (msgInput) {
                msgInput.value += e.detail.unicode;
            }
        })

        // Append to container
        pickerContainer.appendChild(picker);
        document.body.appendChild(pickerContainer);
    }

    // Toggle visibility
    pickerContainer.style.display = pickerContainer.style.display === 'none' ? 'block' : 'none';

    // Position it near the input field
    const msgInput = document.getElementById('message-input');
    const rect = msgInput.getBoundingClientRect();

    pickerContainer.style.top = `${rect.top + window.scrollY - 420}px`; // Position above input
    pickerContainer.style.left = `${rect.left + window.scrollX}px`; // Align with input left
}

function handleSearch(){
    const username = document.getElementById('search').value
    if(username == "") return;
    const request = JSON.stringify({
        type:EVENT.SearchByUsername,
        body:{
            username: username
        }
    })
    ws.send(request)
}
