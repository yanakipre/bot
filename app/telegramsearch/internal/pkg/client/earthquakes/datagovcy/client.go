package datagovcy

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/earthquakes"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const DateTimeTz = time.DateTime + " MST"

var _ earthquakes.Earthquaker = (*client)(nil)

type client struct {
	httpClient *http.Client
	cfg        earthquakes.Config
}

// NOTE: #5 Extract to common package?
type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	XMLName     xml.Name     `xml:"channel"`
	Title       string       `xml:"title"`
	Link        string       `xml:"link"`
	PubDate     *rfc1123Time `xml:"pubDate"`
	Description string       `xml:"description"`
	Items       []rssItem    `xml:"item"`
}
type rssItem struct {
	XMLName xml.Name `xml:"item"`
	Title   *string  `xml:"title"`
	Link    *string  `xml:"link"`
	// Description might contain plain text or some markup format,
	// bytes slice is for parsing readiness
	Description []byte       `xml:"description"`
	PubDate     *rfc1123Time `xml:"pubDate"`
	Guid        *string      `xml:"guid"`
}

type rfc1123Time struct {
	time.Time
}

func (t *rfc1123Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	err := d.DecodeElement(&s, &start)
	if err != nil {
		return fmt.Errorf("decoding xml time string: %w", err)
	}
	if len(s) == 0 {
		return nil
	}

	t.Time, err = time.Parse(time.RFC1123, s)
	if err != nil {
		return fmt.Errorf("parsing xml time: %w", err)
	}
	return nil
}

func NewClient(cfg earthquakes.Config) *client {
	client := &client{
		httpClient: &http.Client{
			Transport: cfg.HTTPTransport.Resolve(),
		},
		cfg: cfg,
	}

	return client
}

var ErrNegativeEarthquakes = errors.New("n cannot be negative")

func (c *client) LatestNEarthquakes(ctx context.Context, n int, minMagnitude float32) ([]earthquakes.Earthquake, error) {
	if n < 0 {
		return nil, ErrNegativeEarthquakes
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.cfg.ApiURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get earthquakes, unexpected status code: %d", res.StatusCode)
	}

	var feed rssFeed
	err = xml.NewDecoder(res.Body).Decode(&feed)
	if err != nil {
		return nil, fmt.Errorf("decoding xml response body: %w", err)
	}

	eqs := make([]earthquakes.Earthquake, 0, n)

	for i := 0; len(eqs) < n && i < len(feed.Channel.Items); i++ {
		v := feed.Channel.Items[i]
		eq, err := parseDesc(v.Description)
		if err != nil {
			return nil, fmt.Errorf("unmarshal datagovcy description: %w", err)
		}
		if eq.Magnitude < minMagnitude {
			continue
		}

		eqs = append(eqs, eq)
	}

	return eqs, nil
}

func parseDesc(data []byte) (earthquakes.Earthquake, error) {
	data = bytes.ReplaceAll(data, []byte{10}, []byte{})
	tree, err := html.ParseFragment(bytes.NewReader(data), &html.Node{
		Type:     html.ElementNode,
		Data:     "table",
		DataAtom: atom.Table,
	})
	if err != nil {
		return earthquakes.Earthquake{}, fmt.Errorf("parsing html doc: %w", err)
	}

	// Finds first occurence of a node
	var findNode func(n *html.Node, elem atom.Atom) *html.Node
	findNode = func(n *html.Node, elem atom.Atom) *html.Node {
		if n == nil {
			return nil
		}
		if n.Type == html.ElementNode && n.DataAtom == elem {
			return n
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if node := findNode(c, elem); node != nil {
				return node
			}
		}

		return nil
	}

	var eq earthquakes.Earthquake

	// Parses values from HTML <table> to Earthquake struct
	var f func(*html.Node) error
	f = func(n *html.Node) error {
		if n == nil {
			return nil
		}

		if n.Type == html.ElementNode && n.DataAtom == atom.Tr {
			th := findNode(n, atom.Th)
			td := findNode(n, atom.Td)
			if th == nil || th.FirstChild == nil || td == nil || td.FirstChild == nil {
				return nil
			}

			// Lowers all letters and removes non-letter characters
			thData := strings.TrimFunc(strings.ToLower(th.FirstChild.Data), func(r rune) bool {
				return !unicode.IsLetter(r)
			})
			tdData := td.FirstChild.Data

			switch thData {
			case "time":
				eq.When, err = time.Parse(DateTimeTz, tdData)
				if err != nil {
				}
			case "position":
				n, err := fmt.Sscanf(tdData, "%f°N, %f°E", &eq.Position.Latitude, &eq.Position.Longitude)
				if err != nil {
					return fmt.Errorf("scanning coordinate data: %w", err)
				}
				if n != 2 {
					return errors.New("invalid format of scanned coordinate data")
				}

			case "place":
			case "major place":
				eq.Location = tdData
			case "depth":
			case "magnitude":
				n, err := fmt.Sscanf(tdData, "%f", &eq.Magnitude)
				if err != nil {
					return fmt.Errorf("scanning magnitude data: %w", err)
				}
				if n != 1 {
					return errors.New("invalid format of scanned magnitude data")
				}
			}

		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			err := f(c)
			if err != nil {
				return err
			}
		}

		return nil
	}

	for _, v := range tree {
		err = f(v)
		if err != nil {
			return earthquakes.Earthquake{}, err
		}
	}

	return eq, nil
}
