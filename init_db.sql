CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT UNIQUE NOT NULL,
	password TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS posts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	content TEXT NOT NULL,
	author_id INTEGER NOT NULL,
	created_at TIMESTAMP NOT NULL,
	category_id INTEGER NOT NULL,
	FOREIGN KEY (author_id) REFERENCES users (id)
	FOREIGN KEY (category_id) REFERENCES categories (id)
	
);

CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content TEXT NOT NULL,
    author_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (author_id) REFERENCES users (id),
    FOREIGN KEY (post_id) REFERENCES posts (id)
);

CREATE TABLE IF NOT EXISTS categories (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS posts_categories (
    post_id INTEGER NOT NULL,
    category TEXT NOT NULL,
    FOREIGN KEY (post_id) REFERENCES posts (id),
    FOREIGN KEY (category) REFERENCES categories (name),
    PRIMARY KEY (post_id, category)
);

CREATE TABLE IF NOT EXISTS post_likes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL,
    liked BOOLEAN NOT NULL,
    disliked BOOLEAN NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (post_id) REFERENCES posts (id),
    UNIQUE (user_id, post_id)
);

CREATE TABLE IF NOT EXISTS comment_likes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    comment_id INTEGER NOT NULL,
    liked BOOLEAN NOT NULL,
    disliked BOOLEAN NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (comment_id) REFERENCES comments (id),
    UNIQUE (user_id, comment_id)
);

