CREATE TABLE quiz_answers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    quiz_session_id INT UNSIGNED NOT NULL,
    question_id INT UNSIGNED NOT NULL,
    user_id INT UNSIGNED NOT NULL,
    answer VARCHAR(255) NOT NULL,
    is_correct BOOLEAN NOT NULL,
    submitted_at DATETIME NOT NULL,
    FOREIGN KEY (quiz_session_id) REFERENCES quiz_sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (question_id) REFERENCES questions(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);