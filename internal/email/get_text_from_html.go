package email

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

func getTextFromHtml(htmlString string) string {
	result := ""

	tokenizer := html.NewTokenizer(strings.NewReader(htmlString))
	prevToken := tokenizer.Token()

domLoop:
	for {
		token := tokenizer.Next()

		switch token {
		case html.ErrorToken:
			break domLoop

		case html.StartTagToken:
			prevToken = tokenizer.Token()

		case html.TextToken:
			if prevToken.Data == "script" || prevToken.Data == "style" {
				continue
			}

			txt := strings.TrimSpace(html.UnescapeString(string(tokenizer.Text())))

			if len(txt) <= 0 {
				continue
			}

			result = fmt.Sprintf("%s%s\n", result, txt)
		}
	}

	return result
}
