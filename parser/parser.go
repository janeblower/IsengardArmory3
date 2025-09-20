package parser

import (
	"ezserver/utils"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

type Config struct {
	URL     string `json:"url"`
	Cookies string `json:"cookies"`
}

func RemoveSpaces(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

type Character struct {
	ID       int
	Name     string
	Login    string
	Class    int
	Race     int
	Guild    string
	LVL      int
	Kills    int
	GS       int
	AP       int
	ExpireAt time.Time `bson:"expireAt"`
}

var classMap = map[string]int{
	"3.png": 0, "9.png": 1, "5.png": 2, "2.png": 3, "8.png": 4,
	"4.png": 5, "11.png": 6, "7.png": 7, "1.png": 8, "6.png": 9,
}

var raceMap = map[string]int{
	"1-0.png": 0, "1-1.png": 0, "3-0.png": 1, "3-1.png": 1,
	"4-0.png": 2, "4-1.png": 2, "7-0.png": 3, "7-1.png": 3,
	"11-0.png": 4, "11-1.png": 4, "2-0.png": 5, "2-1.png": 5,
	"5-0.png": 6, "5-1.png": 6, "6-0.png": 7, "6-1.png": 7,
	"8-0.png": 8, "8-1.png": 8, "10-0.png": 9, "10-1.png": 9,
}

func ParseCharacters(st int, cookie string) ([]Character, int, bool) {
	baseURL := "https://ezwow.org/index.php?app=isengard&module=core&tab=armory&section=characters&realm=1&sort%5Bkey%5D=playtime&sort%5Border%5D=desc&st="
	URL := baseURL + strconv.Itoa(st)

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Cookie", cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	log.Println("ParseCharacters Status:", resp.Status)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	var chars []Character
	found := false

	stMax := parseLastPage(doc)

	doc.Find("table.ipb_table>tbody>tr.character").Each(func(i int, html *goquery.Selection) {
		id := parseCharacterID(html)
		name := parseCharacterName(html)
		class := classMap[path.Base(getAttr(html, "td>span.character-icons>img.character-class", "src"))]
		race := raceMap[path.Base(getAttr(html, "td>span.character-icons>img.character-race", "src"))]
		guild := html.Find("td>span.character-name>span.desc>span>a").Text()
		login := parseLogin(html)
		lvl := utils.ParseInt(RemoveSpaces(html.Find("td.short").Eq(0).Text()))
		kills := utils.ParseInt(RemoveSpaces(html.Find("td.short").Eq(1).Text()))
		gs := utils.ParseInt(RemoveSpaces(html.Find("td.short>span.gearscore>span").Text()))
		ap := utils.ParseInt(RemoveSpaces(html.Find("td.short").Eq(4).Text()))

		char := Character{
			ID:       id,
			Name:     name,
			Login:    login,
			Class:    class,
			Race:     race,
			Guild:    guild,
			LVL:      lvl,
			Kills:    kills,
			GS:       gs,
			AP:       ap,
			ExpireAt: time.Now().Add(20 * time.Hour),
		}
		chars = append(chars, char)
		found = true
	})

	return chars, stMax, found
}

// parseCharacterID извлекает ID персонажа из html
func parseCharacterID(html *goquery.Selection) int {
	idHref := getAttr(html, "td>span.character-name>span>a", "href")
	if idHref == "" {
		idHref = getAttr(html, "td>a", "href")
	}
	parsedURL, _ := url.Parse(idHref)
	id, _ := strconv.Atoi(parsedURL.Query().Get("character"))
	return id
}

// parseCharacterName извлекает имя персонажа из html
func parseCharacterName(html *goquery.Selection) string {
	name := html.Find("td>span.character-name>span>a").Text()
	if name == "" {
		name = html.Find("td").First().Find("a").Text()
	}
	return name
}

// parseLogin извлекает логин персонажа из html
func parseLogin(html *goquery.Selection) string {
	login := html.Find("td>span.member>a>span").Text()
	if login == "" {
		login = strings.TrimSpace(html.Find("span.member").First().Clone().Children().Remove().End().Text())
	}
	return login
}

// getAttr безопасно получает атрибут из селектора
func getAttr(html *goquery.Selection, selector, attr string) string {
	val, _ := html.Find(selector).Attr(attr)
	return val
}

// parseLastPage извлекает максимальное значение st для пагинации
func parseLastPage(doc *goquery.Document) int {
	lastPage := doc.Find("ul.pages>li.page").Last()
	var stMax int
	if lastPage.HasClass("active") {
		stMax = utils.ParseInt(lastPage.Text())
	} else {
		stMax = utils.ParseInt(lastPage.Find("a").Text())
	}
	log.Println("Max page (stMax):", stMax)
	return stMax * 20
}
