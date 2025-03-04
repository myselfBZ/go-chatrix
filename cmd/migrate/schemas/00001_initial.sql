CREATE TABLE IF NOT EXISTS users(
    id  SERIAL PRIMARY KEY NOT NULL,
    username VARCHAR(255),
    password VARCHAR(255),
    name VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS chats (
    id SERIAL PRIMARY KEY,       
    user_1_id INT NOT NULL,     
    user_2_id INT NOT NULL,    
    created_at TIMESTAMP DEFAULT NOW(), 

    CONSTRAINT unique_chat_users UNIQUE (user_1_id, user_2_id),

    FOREIGN KEY (user_1_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (user_2_id) REFERENCES users(id) ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY NOT NULL,
    user_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    chat_id INT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE
);

