CREATE TABLE quiz_sessions (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    quiz_uuid VARCHAR(36) NOT NULL,
    started_at DATETIME NOT NULL,
    ended_at DATETIME,
    participants JSON,
    final_scores JSON,
    FOREIGN KEY (quiz_uuid) REFERENCES quizzes(uuid) ON DELETE CASCADE
);