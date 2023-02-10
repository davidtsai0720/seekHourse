package parser

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/hourse"
	pw "github.com/playwright-community/playwright-go"
)

// https://buy.yungching.com.tw/region/台北市-_c/1000-3000_price/_rm/?pg=2
type ParseYungChing struct {
	PageSize    int
	CurrentPage int
	TotalPage   int
	City        string
	MaxPrice    int
	MinPrice    int
	Selectors   struct {
		ListItem QuerySelector
		Link     QuerySelector
		Detail   QuerySelector
		Total    QuerySelector
		Address  QuerySelector
		Price    QuerySelector
	}
}

func NewParseYungChing(city string) hourse.Parser {
	yc := new(ParseYungChing)
	yc.PageSize = 30
	yc.CurrentPage = 1
	yc.TotalPage = 0
	yc.City = city
	yc.MaxPrice = 3000
	yc.MinPrice = 500
	yc.Selectors.ListItem = QuerySelector{ClassName: []string{"m-list-item"}, TagName: "li"}
	yc.Selectors.Link = QuerySelector{ClassName: []string{"item-img", "ga_click_trace"}, TagName: "a"}
	yc.Selectors.Total = QuerySelector{ClassName: []string{"list-filter", "is-first", "active", "ng-isolate-scope"}, NextTagName: []string{"span"}, TagName: "a"}
	yc.Selectors.Address = QuerySelector{ClassName: []string{"item-description"}, TagName: "div", NextTagName: []string{"span"}}
	yc.Selectors.Detail = QuerySelector{ClassName: []string{"item-info-detail"}, TagName: "ul"}
	yc.Selectors.Price = QuerySelector{ClassName: []string{"price-num"}, TagName: "span"}
	return yc
}

func (yc ParseYungChing) URL() string {
	return fmt.Sprintf(
		"https://buy.yungching.com.tw/region/%s-_c/%d-%d_price/_rm/?pg=%d",
		yc.City, yc.MinPrice, yc.MaxPrice, yc.CurrentPage)
}

func (yc ParseYungChing) HasNext() bool {
	return yc.TotalPage == 0 || yc.TotalPage > yc.CurrentPage
}

func (yc ParseYungChing) ItemQuerySelector() string {
	return yc.Selectors.ListItem.Build()
}

func (yc *ParseYungChing) UpdateCurrentPage() {
	yc.CurrentPage++
}

func (yc *ParseYungChing) SetTotalRow(ctx context.Context, f func(ctx context.Context, qs string) (int, error)) error {
	if yc.TotalPage != 0 {
		return nil
	}

	rows, err := f(ctx, yc.Selectors.Total.Build())
	if err != nil {
		return err
	}

	count := 0
	if rows%yc.PageSize != 0 {
		count++
	}

	yc.TotalPage = rows/yc.PageSize + count
	return nil
}

func (yc ParseYungChing) Link(handle pw.ElementHandle) (string, error) {
	var path *url.URL
	var host *url.URL

	if element, err := handle.QuerySelector(yc.Selectors.Link.Build()); err != nil {
		return "", err
	} else if element == nil {
		return "", nil
	} else if link, err := element.GetAttribute("href"); err != nil {
		return "", err
	} else if strings.HasPrefix(link, "https") {
		return "", errors.New("")
	} else if path, err = url.Parse(link); err != nil {
		return "", err
	} else if host, err = url.Parse("https://buy.yungching.com.tw"); err != nil {
		return "", err
	}

	return host.ResolveReference(path).String(), nil
}

func (yc ParseYungChing) Price(handle pw.ElementHandle) (int, error) {
	return Price(handle, yc.Selectors.Price.Build())
}

func (yc ParseYungChing) Address(item pw.ElementHandle, in *hourse.UpsertHourseRequest) error {
	var err error
	var element pw.ElementHandle

	if element, err = item.QuerySelector(yc.Selectors.Address.Build()); err != nil || element == nil {
		return errors.New("")
	}

	var address string
	address, err = element.TextContent()
	if err != nil {
		return err
	}

	address = strings.Replace(address, yc.City, "", 1)

	var sb strings.Builder
	for _, char := range address {
		sb.WriteRune(char)
		if char == '鄉' || char == '鎮' || char == '市' || char == '區' {
			break
		}
	}

	in.Section = sb.String()
	in.Address = strings.Replace(address, in.Section, "", 1)
	return nil
}

func (yc ParseYungChing) FetchItem(item pw.ElementHandle) (hourse.UpsertHourseRequest, error) {
	var result hourse.UpsertHourseRequest
	if item == nil {
		return result, nil
	}
	var err error

	result.City = yc.City

	result.Link, err = yc.Link(item)
	if err != nil {
		return result, err
	}

	result.Price, err = yc.Price(item)
	if err != nil {
		return result, err
	}

	if err = yc.Address(item, &result); err != nil {
		return result, err
	}

	var detail []pw.ElementHandle
	if detailElement, err := item.QuerySelector(yc.Selectors.Detail.Build()); err != nil {
		return result, err
	} else if detail, err = detailElement.QuerySelectorAll("li"); err != nil {
		return result, err
	} else if len(detail) != 9 {
		return result, errors.New("")
	}

	UpdateField := func(idx int, field *string) error {
		text, err := detail[idx].TextContent()
		if err != nil {
			return err
		}
		text = strings.TrimSpace(text)
		*field = strings.ReplaceAll(text, " ", "")
		return nil
	}

	UpdateField(0, &result.Shape)
	UpdateField(1, &result.Age)
	UpdateField(2, &result.Floor)
	UpdateField(4, &result.Mainarea)
	UpdateField(5, &result.Area)
	UpdateField(6, &result.Layout)

	for _, num := range []int{3, 7, 8} {
		text, err := detail[num].TextContent()
		if err != nil || strings.TrimSpace(text) == "" {
			continue
		}

		result.Others = append(result.Others, strings.TrimSpace(text))
	}

	floor := strings.Split(result.Floor, "~")
	result.Floor = floor[len(floor)-1]
	return result, nil
}