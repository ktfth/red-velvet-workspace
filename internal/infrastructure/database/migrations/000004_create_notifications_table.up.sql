CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

CREATE INDEX idx_notifications_account_id ON notifications(account_id);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
