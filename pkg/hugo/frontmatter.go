package hugo

import (
	"os"
	"strings"
	"time"

	"github.com/gohugoio/hugo/parser/pageparser"
)

const (
	fmTitle      = "title"
	fmAuthor     = "author"
	fmCategories = "categories"
	fmTags       = "tags"

	fmDate        = "date"        // priority high
	fmLastmod     = "lastmod"     // priority middle
	fmPublishDate = "publishDate" // priority low
)

type FrontMatter struct {
	Title    string
	Author   string
	Category string
	Tags     []string
	Date     time.Time
}

// ParseFrontMatter parses the frontmatter of the specified Hugo content.
func ParseFrontMatter(filename string) (*FrontMatter, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfm, err := pageparser.ParseFrontMatterAndContent(file)
	if err != nil {
		return nil, err
	}

	fm := &FrontMatter{}
	if fm.Title, err = getString(&cfm, fmTitle); err != nil {
		return nil, err
	}
	if fm.Author, err = getFirstStringItem(&cfm, fmAuthor); err != nil {
		return nil, err
	}
	if fm.Category, err = getFirstStringItem(&cfm, fmCategories); err != nil {
		return nil, err
	}
	if fm.Tags, err = getAllStringItems(&cfm, fmTags); err != nil {
		return nil, err
	}
	if fm.Date, err = getContentDate(&cfm); err != nil {
		return nil, err
	}

	return fm, nil
}

func getContentDate(cfm *pageparser.ContentFrontMatter) (time.Time, error) {
	for _, key := range []string{fmDate, fmLastmod, fmPublishDate} {
		t, err := getTime(cfm, key)
		if err != nil {
			switch err.(type) {
			case *FMNotExistError:
				continue
			}
		}
		return t, err
	}
	return time.Now(), NewFMNotExistError(
		strings.Join([]string{fmDate, fmLastmod, fmPublishDate}, ", "))
}

func getTime(cfm *pageparser.ContentFrontMatter, fmKey string) (time.Time, error) {
	v, ok := cfm.FrontMatter[fmKey]
	if !ok {
		return time.Now(), NewFMNotExistError(fmKey)
	}
	switch t := v.(type) {
	case string:
		return time.Parse(time.RFC3339, t)
	case time.Time:
		return t, nil
	default:
		return time.Now(), NewFMInvalidTypeError(fmKey, "time.Time or string", t)
	}
}

func getString(cfm *pageparser.ContentFrontMatter, fmKey string) (string, error) {
	v, ok := cfm.FrontMatter[fmKey]
	if !ok {
		return "", NewFMNotExistError(fmKey)
	}

	switch s := v.(type) {
	case string:
		return s, nil
	default:
		return "", NewFMInvalidTypeError(fmKey, "string", s)
	}
}

func getAllStringItems(cfm *pageparser.ContentFrontMatter, fmKey string) ([]string, error) {
	v, ok := cfm.FrontMatter[fmKey]
	if !ok {
		return nil, NewFMNotExistError(fmKey)
	}

	switch arr := v.(type) {
	case []interface{}:
		if len(arr) < 1 {
			return nil, NewFMNotExistError(fmKey)
		}

		var strarr []string
		for _, item := range arr {
			switch s := item.(type) {
			case string:
				strarr = append(strarr, s)
			default:
				return nil, NewFMInvalidTypeError(fmKey, "string", s)
			}
		}
		return strarr, nil

	default:
		return nil, NewFMInvalidTypeError(fmKey, "[]interface{}", arr)
	}
}

func getFirstStringItem(cfm *pageparser.ContentFrontMatter, fmKey string) (string, error) {
	arr, err := getAllStringItems(cfm, fmKey)
	return arr[0], err
}
