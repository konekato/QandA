package db

import "time"

// Question questionテーブル
type Question struct {
	ID       int64  `db:"id"`
	Content  string `db:"q_content"`
	IsHidden int64  `db:"hide"`
	Created  int64  `db:"q_created"`
}

// Answer answerテーブル
type Answer struct {
	QuestionID int64  `db:"a_question_id"`
	Content    string `db:"a_content"`
	Created    int64  `db:"a_created"`
}

// QuestionGood question_goodテーブル
type QuestionGood struct {
	QID int64  `db:"question_id"`
	CID string `db:"cookie_id"`

	Count int64 `db:"count"`
}

// CookieID cookie_idテーブル
type CookieID struct {
	ID    string `db:"id"`
	Count int64  `db:"count"`
}

type QAndA struct {
	Question
	Answer
	QuestionGood
}

func FetchQAndAsByPage(orderBy, sortBy, onlyAnswered, cookieID string, page, offset int) ([]QAndA, error) {
	// Selects
	q := `SELECT` +
		` q.id AS id, q.content AS q_content, COALESCE(q.hide, 1) AS hide, COALESCE(q.created, 0) AS q_created,` +
		` COALESCE(a.question_id, 0) AS a_question_id, COALESCE(a.content, '') AS a_content, COALESCE(a.created, 0) AS a_created,` +
		` COALESCE(qg.count, 0) AS count,` +
		` COALESCE(qg2.cookie_id, '') AS cookie_id` +
		` FROM question AS q` +
		` LEFT JOIN answer AS a` +
		` ON a.question_id = q.id` +
		` LEFT JOIN` +
		` (SELECT question_id, count(*) AS count FROM question_good GROUP BY question_id) AS qg` +
		` ON qg.question_id = q.id` +
		` LEFT JOIN` +
		` (SELECT question_id, cookie_id FROM question_good WHERE cookie_id=$1) AS qg2` +
		` ON qg2.question_id = q.id` +
		` WHERE q.hide = 0`
	if onlyAnswered == "true" {
		q += ` AND a.question_id > 0`
	}
	q += ` ORDER BY ` + orderBy + ` ` + sortBy + ` ` +
		` LIMIT $2` +
		` OFFSET $3`
	var QAndAs []QAndA
	if err := db.Select(&QAndAs, q, cookieID, page, offset); err != nil {
		return nil, err
	}
	return QAndAs, nil
}

func InsertQuestion(content string, isHidden int) error {
	q := `INSERT INTO question (content, hide, created) VALUES ($1, $2, $3)`
	if _, err := db.Exec(q, content, isHidden, time.Now().Unix()); err != nil {
		return err
	}
	return nil
}

func FetchQuestionGoodByQIDAndCID(qID int64, cID string) (QuestionGood, error) {
	q := `SELECT question_id, cookie_id FROM question_good WHERE question_id = $1 AND cookie_id = $2`
	var qg QuestionGood
	if err := db.Get(&qg, q, qID, cID); err != nil {
		return QuestionGood{}, err
	}
	return qg, nil
}

func InsertQuestionGoodByQIDAndCID(qID int64, cID string) error {
	q := `INSERT INTO question_good VALUES ($1, $2)`
	if _, err := db.Exec(q, qID, cID); err != nil {
		return err
	}
	return nil
}

func DeleteQuestionGoodByQIDAndCID(qID int64, cID string) error {
	q := `DELETE FROM question_good WHERE question_id = $1 AND cookie_id = $2`
	if _, err := db.Exec(q, qID, cID); err != nil {
		return err
	}
	return nil
}

func FetchCookieById(id string) (CookieID, error) {
	q := `SELECT COUNT(*) AS count FROM cookie_id WHERE id = $1`
	var cID CookieID
	if err := db.Get(&cID, q, id); err != nil {
		return CookieID{}, err
	}
	return cID, nil
}

func InsertCookie(secret string) error {
	q := `INSERT INTO cookie_id VALUES ($1)`
	if _, err := db.Exec(q, secret); err != nil {
		return err
	}
	return nil
}
