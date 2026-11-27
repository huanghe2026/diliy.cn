// Package store 封装 geoyuan 准入服务的本地 SQLite 存储。
//
// 表：
//   - users：注册用户（SSH 公钥指纹绑定）
//   - attempts：每次答题记录（SSH 端）
//   - web_passes：Web 端通过记录，可免 SSH 重测（带 token）
package store

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"database/sql"

	_ "modernc.org/sqlite" // 纯 Go SQLite 驱动
)

// 用户准入状态
const (
	StatusPending  = "pending"  // 已注册未通过
	StatusApproved = "approved" // 已通过，加入 team
)

// User 对应 users 表（SSH 注册用户）
type User struct {
	ID                int64
	Username          string
	DisplayName       string
	Tagline           string
	Lang              string
	PubKeyFingerprint string
	Status            string
	CreatedAt         string
}

// Attempt 对应 attempts 表
type Attempt struct {
	ID        int64
	UserID    int64
	Score     int
	Total     int
	Passed    bool
	CreatedAt string
}

// WebPass 对应 web_passes 表（Web 端通过记录）
type WebPass struct {
	ID        int64
	Username  string
	Email     string
	Token     string
	Score     int
	Total     int
	Passed    bool
	Used      bool // 是否已绑定到 SSH 用户
	CreatedAt string
}

// Store 持有 SQLite 连接
type Store struct {
	db *sql.DB
}

// Open 打开（或创建）SQLite 数据库，初始化表结构并开启 WAL。
func Open(path string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}
	// 迁移：新增 display_name / tagline / lang 字段（兼容已有 DB）
	for _, m := range []string{
		"ALTER TABLE users ADD COLUMN display_name TEXT",
		"ALTER TABLE users ADD COLUMN tagline TEXT",
		"ALTER TABLE users ADD COLUMN lang TEXT DEFAULT 'en'",
	} {
		// SQLite 没有 IF NOT EXISTS，用错误忽略方式
		_, _ = db.Exec(m)
	}
	return &Store{db: db}, nil
}

const schemaSQL = `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	display_name TEXT,
	tagline TEXT,
	lang TEXT DEFAULT 'en',
	pubkey_fingerprint TEXT NOT NULL UNIQUE,
	status TEXT NOT NULL DEFAULT 'pending',
	created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS attempts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	score INTEGER NOT NULL,
	total INTEGER NOT NULL,
	passed INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE TABLE IF NOT EXISTS web_passes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL,
	email TEXT NOT NULL,
	token TEXT NOT NULL UNIQUE,
	score INTEGER NOT NULL,
	total INTEGER NOT NULL,
	passed INTEGER NOT NULL,
	used INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_attempts_user ON attempts(user_id);
CREATE INDEX IF NOT EXISTS idx_web_passes_token ON web_passes(token);
`

// Close 关闭数据库
func (s *Store) Close() error { return s.db.Close() }

// DB 返回底层 *sql.DB（供复杂查询/事务绕过封装）
func (s *Store) DB() *sql.DB { return s.db }

// ─── User 查询 ───

// GetUserByFingerprint 按公钥指纹查用户；不存在返回 (nil, nil)
func (s *Store) GetUserByFingerprint(fp string) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(
		`SELECT id, username, COALESCE(display_name,''), COALESCE(tagline,''), COALESCE(lang,'en'), pubkey_fingerprint, status, created_at FROM users WHERE pubkey_fingerprint = ?`,
		fp,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.Tagline, &u.Lang, &u.PubKeyFingerprint, &u.Status, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserByUsername 按用户名查用户；不存在返回 (nil, nil)
func (s *Store) GetUserByUsername(name string) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(
		`SELECT id, username, COALESCE(display_name,''), COALESCE(tagline,''), COALESCE(lang,'en'), pubkey_fingerprint, status, created_at FROM users WHERE username = ?`,
		name,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.Tagline, &u.Lang, &u.PubKeyFingerprint, &u.Status, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserByID 按用户 ID 查用户；不存在返回 (nil, nil)
func (s *Store) GetUserByID(id int64) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(
		`SELECT id, username, COALESCE(display_name,''), COALESCE(tagline,''), COALESCE(lang,'en'), pubkey_fingerprint, status, created_at FROM users WHERE id = ?`,
		id,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.Tagline, &u.Lang, &u.PubKeyFingerprint, &u.Status, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// ListApprovedUsers 列出所有已通过准入的用户（个人主页反查 owner → 用户用）
func (s *Store) ListApprovedUsers() ([]User, error) {
	rows, err := s.db.Query(
		`SELECT id, username, COALESCE(display_name,''), COALESCE(tagline,''), COALESCE(lang,'en'), pubkey_fingerprint, status, created_at FROM users WHERE status = ? ORDER BY id`,
		StatusApproved,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &u.Tagline, &u.Lang, &u.PubKeyFingerprint, &u.Status, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// CreateUser 新建用户（pending 状态），返回 user id
func (s *Store) CreateUser(username, fingerprint, displayName, tagline, lang string) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	if lang == "" {
		lang = "en"
	}
	res, err := s.db.Exec(
		`INSERT INTO users (username, display_name, tagline, lang, pubkey_fingerprint, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		username, displayName, tagline, lang, fingerprint, StatusPending, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// DeleteUser 删除用户及其答题记录（级联）— 用于 Switch User 替换旧身份
func (s *Store) DeleteUser(userID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 级联删 attempts
	if _, err := tx.Exec(`DELETE FROM quiz_attempts WHERE user_id = ?`, userID); err != nil {
		return err
	}
	// 删用户
	if _, err := tx.Exec(`DELETE FROM users WHERE id = ?`, userID); err != nil {
		return err
	}
	return tx.Commit()
}

// RecordAttempt 记录一次答题结果，若通过则同步标记用户 approved
func (s *Store) RecordAttempt(userID int64, score, total int, passed bool) error {
	now := time.Now().UTC().Format(time.RFC3339)
	passedInt := 0
	if passed {
		passedInt = 1
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(
		`INSERT INTO attempts (user_id, score, total, passed, created_at) VALUES (?, ?, ?, ?, ?)`,
		userID, score, total, passedInt, now,
	); err != nil {
		return err
	}
	if passed {
		if _, err := tx.Exec(`UPDATE users SET status = ? WHERE id = ?`, StatusApproved, userID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// CountApproved 返回已通过人数
func (s *Store) CountApproved() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users WHERE status = ?`, StatusApproved).Scan(&n)
	return n, err
}

// CountAttemptsForUser 返回某用户答题次数
func (s *Store) CountAttemptsForUser(userID int64) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM attempts WHERE user_id = ?`, userID).Scan(&n)
	return n, err
}

// ─── Web Pass （Web 端答题通过） ───

// CreateWebPass 创建一条 Web 端通过记录，返回 token
func (s *Store) CreateWebPass(username, email string, score, total int, passed bool) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	passedInt := 0
	if passed {
		passedInt = 1
	}
	_, err = s.db.Exec(
		`INSERT INTO web_passes (username, email, token, score, total, passed, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		username, email, token, score, total, passedInt, now,
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

// GetWebPassByToken 按 token 查 Web 通过记录
func (s *Store) GetWebPassByToken(token string) (*WebPass, error) {
	w := &WebPass{}
	var passedInt, usedInt int
	err := s.db.QueryRow(
		`SELECT id, username, email, token, score, total, passed, used, created_at FROM web_passes WHERE token = ?`,
		token,
	).Scan(&w.ID, &w.Username, &w.Email, &w.Token, &w.Score, &w.Total, &passedInt, &usedInt, &w.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	w.Passed = passedInt == 1
	w.Used = usedInt == 1
	return w, nil
}

// MarkWebPassUsed 标记 Web pass 已绑定 SSH
func (s *Store) MarkWebPassUsed(token string) error {
	_, err := s.db.Exec(`UPDATE web_passes SET used = 1 WHERE token = ?`, token)
	return err
}

// generateToken 生成 32 字符 hex 随机 token
func generateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
