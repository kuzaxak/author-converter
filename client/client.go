package client

import (
	"encoding/json"
	"fmt"
	"github.com/bmaupin/go-epub"
	"github.com/cheggaaa/pb/v3"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"time"
)

type client struct {
	client  *http.Client
	cookie  *http.Cookie
	baseUrl string
}

type chapter struct {
	TextLength      int    `json:"textLength"`
	Id              int    `json:"id"`
	Title           string `json:"title"`
	PublishTime     string `json:"publishTime"`
	AutoPublishTime string `json:"autoPublishTime"`
}

type readerIndex struct {
	WorkId          string    `json:"workId"`
	WorkTitle       string    `json:"workTitle"`
	WorkForm        string    `json:"workForm"`
	AuthorName      string    `json:"authorName"`
	Protection      bool      `json:"protection"`
	ChapterId       string    `json:"chapterId"`
	Chapters        []chapter `json:"chapters"`
	ChapterProgress int       `json:"chapterProgress"`
	SessionId       string    `json:"sessionId"`
}

type chapterResponse struct {
	IsSuccessful bool              `json:"isSuccessful"`
	IsWarning    bool              `json:"isWarning"`
	Messages     string            `json:"messages"`
	Data         map[string]string `json:"data"`
}

func NewClient(baseUrl string) *client {
	cookieJar, _ := cookiejar.New(nil)

	return &client{
		client: &http.Client{
			Jar:     cookieJar,
			Timeout: time.Second * 20,
		},
		baseUrl: baseUrl,
	}
}

func (a *client) newRequest(method, url string, body io.Reader) (*http.Response, []byte) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", a.baseUrl, url), body)
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}

	a.setHeaders(req)
	a.addAuthCookie(req) // TODO add login command

	// Send request
	resp, err := a.client.Do(req)
	if err != nil {
		log.Fatal("Error reading response. ", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading body. ", err)
	}
	return resp, respBody
}

func (a *client) parseChaptersJson(body []byte) *[]chapter {
	var re = regexp.MustCompile(`(?ms)"readerIndex"(.*chapters:(\s+\[.*?\]),.*?})\);`)

	jsonString := re.FindStringSubmatch(string(body))

	if len(jsonString) < 3 {
		log.Fatal("Error parsing body. ")
		return nil
	}

	var chapters []chapter
	err := json.Unmarshal([]byte(jsonString[2]), &chapters)

	if err != nil {
		log.Fatal("Error json.Unmarshal body. ", err)
	}

	return &chapters
}

func (a *client) parseWorkTitle(body []byte) string {
	var re = regexp.MustCompile(`(?ms)workTitle:\s+['"](.*?)['"],`)
	parsed := re.FindStringSubmatch(string(body))

	if len(parsed) < 2 {
		log.Fatal("Error parsing body. ")
		return ""
	}

	return parsed[1]
}

func (a *client) parseAuthor(body []byte) string {
	var re = regexp.MustCompile(`(?ms)authorName:\s+['"](.*?)['"],`)
	parsed := re.FindStringSubmatch(string(body))

	if len(parsed) < 2 {
		log.Fatal("Error parsing body. ")
		return ""
	}

	return parsed[1]
}

func (a *client) parseUserId(body []byte) string {
	var re = regexp.MustCompile(`(?ms)userId:\s+(.*?),`)
	parsed := re.FindStringSubmatch(string(body))

	if len(parsed) < 2 {
		log.Fatal("Error parsing body. ")
		return ""
	}

	return parsed[1]
}

func (a *client) GetChapters(bookId string) *epub.Epub {
	rand.Seed(time.Now().UnixNano())

	_, respBody := a.newRequest("GET", fmt.Sprintf("reader/%s", bookId), nil)
	chapters := a.parseChaptersJson(respBody)

	e := epub.NewEpub(a.parseWorkTitle(respBody))
	e.SetAuthor(a.parseAuthor(respBody))

	userId := a.parseUserId(respBody)

	var brRegexp = regexp.MustCompile(`(?ms)(<br>\s?(<\/br>)?)`)
	var colorRegexp = regexp.MustCompile(`(?ms)(color:#)([0-9a-fA-F]+)([;"'])`)
	var textRegexp = regexp.MustCompile(`(?ms)"text":\s+"(.*)"\s+}\s+}`)

	bar := pb.StartNew(len(*chapters))

	for _, chapter := range *chapters {
		time.Sleep(time.Duration(rand.Intn(20)+5) * time.Second)
		resp, chapterBody := a.newRequest("GET", fmt.Sprintf("reader/%s/chapter?id=%d", bookId, chapter.Id), nil)
		reText := textRegexp.FindSubmatch(chapterBody)
		if len(reText) < 2 {
			log.Fatal("Error parse body. ")
		}

		eText, err := decrypt(reText[1], decryptKey(resp.Header.Get("Reader-Secret"), userId))
		if err != nil {
			log.Fatal("Error decrypting. ", err)
		}

		// fix unclosed br's
		eText = brRegexp.ReplaceAll(eText, []byte(""))
		// remove custom text colors
		eText = colorRegexp.ReplaceAll(eText, []byte("$1ffffff$3"))

		_, err = e.AddSection(fmt.Sprintf("<h1>%s</h1>%s", chapter.Title, string(eText)), chapter.Title, "", "")
		if err != nil {
			log.Fatal("Error compiling to the book. ", err)
		}

		bar.Increment()
	}

	bar.Finish()

	return e
}
