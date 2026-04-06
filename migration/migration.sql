CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    monthly_limit NUMERIC(15,2) DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS saving_goals (
                                            id SERIAL PRIMARY KEY,
                                            user_id INTEGER NOT NULL,
                                            name VARCHAR(50) NOT NULL,
    amount NUMERIC(15,2) NOT NULL,
    deadline DATE NOT NULL,
    status VARCHAR(50) NOT NULL,
    description VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT fk_users_goals
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE
    );

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    goal_id INTEGER,
    amount BIGINT NOT NULL,
    category VARCHAR(50) NOT NULL,
    description VARCHAR(50) NOT NULL,
    transaction_type VARCHAR(50) NOT NULL,
    transaction_source VARCHAR(50) NOT NULL,
    transaction_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT fk_transaction_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_transaction_goals
        FOREIGN KEY (goal_id)
        REFERENCES saving_goals(id)
        ON DELETE CASCADE
);


