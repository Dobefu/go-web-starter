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

	for {
		token := tokenizer.Next()

		switch token {
		case html.ErrorToken:
			return result

		case html.StartTagToken:
			prevToken = tokenizer.Token()

		case html.TextToken:
			if prevToken.Data == "script" ||
				prevToken.Data == "style" ||
				prevToken.Data == "head" ||
				prevToken.Data == "html" {
				continue
			}

			txt := strings.TrimSpace(html.UnescapeString(string(tokenizer.Text())))

			if len(txt) <= 0 {
				continue
			}

			if prevToken.Data == "a" {
				href := getAttributeValue(prevToken.Attr, "href")

				if len(href) > 0 {
					txt = fmt.Sprintf("%s (%s)", txt, href)
				}
			}

			if strings.Contains(prevToken.String(), "footer") {
				result = fmt.Sprintf("%s\n\n", result)
			}

			result = fmt.Sprintf("%s%s\n", result, txt)

			if strings.Contains(prevToken.String(), "text-xl") {
				result = fmt.Sprintf("%s%s\n\n\n", result, strings.Repeat("-", len(txt)))
			}
		}
	}
}

func getAttributeValue(attributes []html.Attribute, attributeName string) string {
	for _, attr := range attributes {
		if attr.Key != "href" {
			continue
		}

		return attr.Val
	}

	return ""
}
