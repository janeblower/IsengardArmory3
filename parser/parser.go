package parser

import (
	"fmt"
	"net/http"
	"net/url"
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
			return -1 // удаляем
		}
		return r
	}, s)
}

type Character struct {
	ID       int
	Name     string
	Login    string
	Class    string
	Race     string
	Guild    string
	LVL      uint
	Kills    uint
	GS       uint
	AP       uint
	ExpireAt time.Time `bson:"expireAt"`
}

func ParseCharacters(st int, cookie string) ([]Character, int, bool) {

	baseURL := "https://ezwow.org/index.php?app=isengard&module=core&tab=armory&section=characters&realm=1&sort%5Bkey%5D=playtime&sort%5Border%5D=desc&st="
	URL := baseURL + strconv.Itoa(st)

	req, _ := http.NewRequest("GET", URL, nil)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("ParseCharacters Status:", resp.Status)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	found := false

	var chars []Character

	doc.Find("table.ipb_table>tbody>tr.character").Each(func(i int, html *goquery.Selection) {
		idHref, _ := html.Find("td>span.character-name>span>a").Attr("href")
		if idHref == "" {
			idHref, _ = html.Find("td").First().Find("a").Attr("href")
		}
		parsedURL, _ := url.Parse(idHref)
		id, _ := strconv.Atoi(parsedURL.Query().Get("character"))
		name := html.Find("td>span.character-name>span>a").Text()
		if name == "" {
			name = html.Find("td").First().Find("a").Text()
		}
		class, _ := html.Find("td>span.character-icons>img.character-class").Attr("title")
		if start := strings.Index(class, "("); start != -1 {
			if end := strings.Index(class, ")"); end != -1 && start < end {
				class = class[start+1 : end]
			}
		}
		race, _ := html.Find("td>span.character-icons>img.character-race").Attr("title")
		guild := html.Find("td>span.character-name>span.desc>span>a").Text()
		login := html.Find("td>span.member>a>span").Text()
		if login == "" {
			login = strings.TrimSpace(html.Find("span.member").First().Clone().Children().Remove().End().Text())
		}
		lvl, _ := strconv.Atoi(RemoveSpaces(html.Find("td.short").Eq(0).Text()))
		kills, _ := strconv.Atoi(RemoveSpaces(html.Find("td.short").Eq(1).Text()))
		gs, _ := strconv.Atoi(RemoveSpaces(html.Find("td.short>span.gearscore>span").Text()))
		ap, _ := strconv.Atoi(RemoveSpaces(html.Find("td.short").Eq(4).Text()))

		char := Character{
			ID:       id,
			Name:     name,
			Login:    login,
			Class:    class,
			Race:     race,
			Guild:    guild,
			LVL:      uint(lvl),
			Kills:    uint(kills),
			GS:       uint(gs),
			AP:       uint(ap),
			ExpireAt: time.Now().Add(18 * time.Hour),
		}
		chars = append(chars, char)
		found = true
	})

	lastPage := doc.Find(`a[title="Перейти к последней странице"]`).First()
	href, _ := lastPage.Attr("href")

	parsedURL, _ := url.Parse(href)
	stMax, _ := strconv.Atoi(parsedURL.Query().Get("st"))

	return chars, stMax, found

}
