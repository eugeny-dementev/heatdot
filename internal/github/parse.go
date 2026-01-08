package github

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	ariaDateRe  = regexp.MustCompile(`\bon ([A-Za-z]+ \d{1,2}, \d{4})`)
	ariaCountRe = regexp.MustCompile(`^(\d+|No) contributions?`)
)

func ParseCountForDate(html []byte, date time.Time) (int, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return 0, err
	}

	dateStr := date.Format("2006-01-02")
	if count, ok := findCountByDataAttrs(doc, dateStr); ok {
		return count, nil
	}

	if count, ok := findCountByTooltip(doc, dateStr); ok {
		return count, nil
	}

	if count, ok := findCountByAriaLabel(doc, date); ok {
		return count, nil
	}

	return 0, fmt.Errorf("no contribution count found for %s", dateStr)
}

func findCountByDataAttrs(doc *goquery.Document, dateStr string) (int, bool) {
	if count, ok := findCountInSelection(doc.Find("rect[data-date][data-count]"), dateStr); ok {
		return count, true
	}

	return findCountInSelection(doc.Find("[data-date][data-count]"), dateStr)
}

func findCountInSelection(sel *goquery.Selection, dateStr string) (int, bool) {
	var (
		count int
		ok    bool
	)

	sel.EachWithBreak(func(_ int, s *goquery.Selection) bool {
		dateVal, exists := s.Attr("data-date")
		if !exists || dateVal != dateStr {
			return true
		}
		countVal, exists := s.Attr("data-count")
		if !exists {
			return true
		}
		parsed, err := strconv.Atoi(strings.TrimSpace(countVal))
		if err != nil {
			return true
		}
		count = parsed
		ok = true
		return false
	})

	return count, ok
}

func findCountByAriaLabel(doc *goquery.Document, date time.Time) (int, bool) {
	date = date.In(date.Location())
	var (
		count int
		ok    bool
	)

	doc.Find("[aria-label]").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		label, exists := s.Attr("aria-label")
		if !exists {
			return true
		}
		label = strings.TrimSpace(label)
		if !strings.Contains(strings.ToLower(label), "contribution") {
			return true
		}
		parsedCount, parsedDate, parsed := parseAriaLabel(label, date.Location())
		if !parsed {
			return true
		}
		if sameDay(parsedDate, date) {
			count = parsedCount
			ok = true
			return false
		}
		return true
	})

	return count, ok
}

func findCountByTooltip(doc *goquery.Document, dateStr string) (int, bool) {
	counts := map[string]int{}

	doc.Find("tool-tip[for]").Each(func(_ int, s *goquery.Selection) {
		target, exists := s.Attr("for")
		if !exists {
			return
		}
		text := strings.TrimSpace(s.Text())
		if text == "" {
			return
		}
		if count, ok := parseTooltipCount(text); ok {
			counts[target] = count
		}
	})

	if len(counts) == 0 {
		return 0, false
	}

	var (
		count int
		ok    bool
	)

	doc.Find("[data-date][id]").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		dateVal, exists := s.Attr("data-date")
		if !exists || dateVal != dateStr {
			return true
		}
		id, exists := s.Attr("id")
		if !exists {
			return true
		}
		mapped, exists := counts[id]
		if !exists {
			return true
		}
		count = mapped
		ok = true
		return false
	})

	return count, ok
}

func parseTooltipCount(text string) (int, bool) {
	countMatch := ariaCountRe.FindStringSubmatch(text)
	if len(countMatch) < 2 {
		return 0, false
	}
	countStr := strings.ToLower(countMatch[1])
	if countStr == "no" {
		return 0, true
	}
	count, err := strconv.Atoi(countMatch[1])
	if err != nil {
		return 0, false
	}
	return count, true
}

func parseAriaLabel(label string, loc *time.Location) (int, time.Time, bool) {
	dateMatch := ariaDateRe.FindStringSubmatch(label)
	if len(dateMatch) < 2 {
		return 0, time.Time{}, false
	}

	dateVal := dateMatch[1]
	parsedDate, err := time.ParseInLocation("January 2, 2006", dateVal, loc)
	if err != nil {
		return 0, time.Time{}, false
	}

	countMatch := ariaCountRe.FindStringSubmatch(label)
	if len(countMatch) < 2 {
		return 0, time.Time{}, false
	}

	countStr := strings.ToLower(countMatch[1])
	if countStr == "no" {
		return 0, parsedDate, true
	}
	count, err := strconv.Atoi(countMatch[1])
	if err != nil {
		return 0, time.Time{}, false
	}

	return count, parsedDate, true
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
