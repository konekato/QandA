package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"kone-server/db"
	"kone-server/pkg/log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

func to400(c echo.Context) error {
	return c.Redirect(http.StatusSeeOther, os.Getenv("URI")+"/400/")
}

func to500(c echo.Context) error {
	return c.Redirect(http.StatusSeeOther, os.Getenv("URI")+"/500/")
}

func contact(c echo.Context) error {
	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		log.Error("Input error: 'name' is empty.")
		return to400(c)
	}

	mail := strings.TrimSpace(c.FormValue("mail"))
	if mail == "" {
		log.Error("Input error: 'mail' is empty.")
		return to400(c)
	}

	phone := strings.TrimSpace(c.FormValue("phone"))

	kind := c.FormValue("kind")
	var kindStr string
	const (
		kindOpinion    = "ご意見/ご質問"
		kindScout      = "スカウト/ご依頼"
		kindImpression = "本ページのご感想"
	)
	switch kind {
	case "opinion":
		kindStr = kindOpinion
	case "scout":
		kindStr = kindScout
	case "impression":
		kindStr = kindImpression
	default:
		log.Error("Input error: 'kind' is invalid.")
		return to400(c)
	}

	content := strings.TrimSpace(c.FormValue("content"))
	if content == "" {
		log.Error("Input error: 'content' is empty.")
		return to400(c)
	}

	if err := updateSpreadSheet(&contactFormValues{
		Name:    name,
		Mail:    mail,
		Phone:   phone,
		Kind:    kindStr,
		Content: content,
	}); err != nil {
		log.Error("Error under the updating spread sheet: ", err)
		return to500(c)
	}

	return c.Redirect(http.StatusSeeOther, os.Getenv("URI")+"/contact/success/")
}

type contactFormValues struct {
	Name    string
	Mail    string
	Phone   string
	Kind    string
	Content string
}

func updateSpreadSheet(v *contactFormValues) error {
	secret, err := ioutil.ReadFile("secret.json")
	if err != nil {
		return err
	}
	conf, err := google.JWTConfigFromJSON(secret, sheets.SpreadsheetsScope)
	if err != nil {
		return err
	}
	client := conf.Client(context.Background())
	srv, err := sheets.New(client)
	if err != nil {
		return err
	}
	spreadsheetID := os.Getenv("SHEET_ID")

	// 最終行の取得
	readRange := "シート1!A:B"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return err
	}
	l := len(resp.Values)
	if l == 0 {
		return fmt.Errorf("No data found")
	}

	// Update
	var vr sheets.ValueRange
	now := time.Now().Format("2006/01/02 15:04:05")
	vr.Values = append(vr.Values, []interface{}{l, v.Name, v.Mail, v.Phone, v.Kind, v.Content, now})
	updateRange := fmt.Sprintf("シート1!A%d:G%d", l+1, l+1)
	if _, err = srv.Spreadsheets.Values.Update(spreadsheetID, updateRange, &vr).ValueInputOption("RAW").Do(); err != nil {
		return err
	}

	return nil
}

func qAnda(c echo.Context) error {
	cookieID, err := fetchCookieID(c)
	if err != nil {
		log.Error(err)
	}

	var isNg bool
	if c.QueryParam("ng") == "true" {
		isNg = true
	}

	const pageLimit = 6
	pageParam := c.QueryParam("page")
	page, err := strconv.Atoi(pageParam)
	if err != nil || page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageLimit

	orderBy := c.QueryParam("order_by")
	if orderBy != "count" {
		orderBy = "id"
	}

	sortBy := c.QueryParam("sort_by")
	if sortBy != "asc" {
		sortBy = "desc"
	}

	onlyAnswered := c.QueryParam("only_answered")
	if onlyAnswered != "true" {
		onlyAnswered = "false"
	}

	QAndAs, err := db.FetchQAndAsByPage(orderBy, sortBy, onlyAnswered, cookieID, pageLimit+1, offset)
	if err != nil {
		log.Error(err)
		return to500(c)
	}

	// Pagenation
	nextPage := page
	if len(QAndAs) == pageLimit+1 {
		nextPage++
		QAndAs = QAndAs[:pageLimit]
	}
	prevPage := page
	if page != 1 {
		prevPage--
	}

	// Show target data
	var target db.QAndA
	qIDParam := c.QueryParam("target")
	qID, err := strconv.ParseInt(qIDParam, 10, 64)
	if qIDParam == "" {
		qID = 0
	} else if err != nil {
		log.Info(err)
		qID = 0
	}
	for _, qa := range QAndAs {
		if qID == qa.ID {
			target = qa
			break
		}
	}

	return c.Render(http.StatusOK, "qanda", map[string]interface{}{
		"QAndAs":       QAndAs,
		"page":         page,
		"nextPage":     nextPage,
		"prevPage":     prevPage,
		"orderBy":      orderBy,
		"sortBy":       sortBy,
		"onlyAnswered": onlyAnswered,
		"isNg":         isNg,
		"target":       target,
		"cookieID":     cookieID,
	})
}

func qAndaQCreate(c echo.Context) error {
	content := strings.TrimSpace(c.FormValue("q"))
	content = template.HTMLEscapeString(content)

	uri := "/githubio/qanda"
	var isHidden int

	// 改行チェック、xss
	// TODO NGワード検出/NGワードが検出された可能性が有ります。管理者が確認致しますので、もうしばらくお待ちください。
	if content == "" {
		uri += "?ng=true"
		isHidden = 1
		log.Infof("This question saved hidden: [%s]", content)
	}

	if err := db.InsertQuestion(content, isHidden); err != nil {
		log.Error(err)
		return to500(c)
	}

	return c.Redirect(http.StatusSeeOther, uri)
}

func qAndaQuestionGoodCount(c echo.Context) error {
	cID, err := fetchCookieID(c)
	if err != nil {
		log.Error(err)
		//TODO
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{})
	}

	qIDParam := c.Param("question_id")
	qID, err := strconv.ParseInt(qIDParam, 10, 64)
	if err != nil {
		log.Error(err)
		//TODO
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{})
	}

	qg, err := db.FetchQuestionGoodByQIDAndCID(qID, cID)
	if err == sql.ErrNoRows {
		// no rows の場合は insert
		if err := db.InsertQuestionGoodByQIDAndCID(qID, cID); err != nil {
			log.Error(err)
			//TODO
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{})
		}
	} else if err != nil {
		log.Error(err)
		//TODO
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{})
	} else {
		// 存在している場合は delete
		if err := db.DeleteQuestionGoodByQIDAndCID(qg.QID, qg.CID); err != nil {
			log.Error(err)
			//TODO
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{})
}

func qAndaCookie(c echo.Context) error {
	return c.Render(http.StatusOK, "qanda_cookie", map[string]interface{}{})
}

func qAndaCookieAccept(c echo.Context) error {
	cID, err := createCookieID()
	if err != nil {
		log.Error(err)
		return to500(c)
	}

	cookie := &http.Cookie{
		Path:     "/githubio/qanda",
		Name:     "cookie_id",
		Value:    cID,
		Expires:  time.Date(2035, time.December, 31, 23, 59, 59, 0, time.UTC),
		HttpOnly: true,
		// Secure: true,
	}

	http.SetCookie(c.Response().Writer, cookie)

	return c.Redirect(http.StatusSeeOther, "/githubio/qanda")
}

func fetchCookieID(c echo.Context) (string, error) {
	cookie, err := c.Request().Cookie("cookie_id")
	if err != nil {
		if err == http.ErrNoCookie {
			return "", nil
		}
		return "", err
	}

	cID, err := db.FetchCookieById(cookie.Value)
	if err != nil {
		return "", err
	}

	if cID.Count <= 0 {
		return "", fmt.Errorf("No match cookie_id: %s", cookie.Value)
	}

	return cookie.Value, nil
}

func createCookieID() (string, error) {
	const secretLength = 10
	s, err := makeRandomStr(secretLength)
	if err != nil {
		return "", err
	}
	if err := db.InsertCookie(s); err != nil {
		return "", err
	}
	log.Infof("Created and inserted CookieID: %s", s)
	return s, nil
}

func makeRandomStr(digit uint) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, digit)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	var result string
	for _, v := range b {
		result += string(letters[int(v)%len(letters)])
	}
	return result, nil
}
