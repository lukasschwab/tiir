package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/lukasschwab/tiir/pkg/text"
)

type MigrateCommand struct {
	FromFilename string `short:"f" required:"" name:"from" help:"HTML file to migrate."`
}

func (command *MigrateCommand) Run(rt *runtime) error {
	f, err := os.Open(command.FromFilename)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()
	p, err := NewParser(f)
	if err != nil {
		return fmt.Errorf("initialize parser: %w", err)
	}

	p.Parse()
	log.Printf("Writing %v texts to app", len(p.parsed))
	for _, text := range p.parsed {
		created, err := rt.cfg.App.Create(text)
		if err != nil {
			return fmt.Errorf("create text: %w", err)
		}
		log.Printf("Created text %v: %+v", created.ID, created)
	}
	log.Printf("Wrote %v texts to app", len(p.parsed))
	return nil
}

// NewParser constructs a new BrittleParser for f without parsing it. See
// [BrittleParser.Parse].
func NewParser(f *os.File) (*BrittleParser, error) {
	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return nil, fmt.Errorf("error constructing reader: %w", err)
	}
	return &BrittleParser{Document: doc, parsed: []*text.Text{}}, nil
}

// BrittleParser for old (pre-Go) tir HTML format; not guaranteed to be
// forwards-compatible. See https://github.com/lukasschwab/tir for an example of
// the supported document structure.
type BrittleParser struct {
	*goquery.Document
	parsed []*text.Text
}

// Parse the full table row by row.
func (p *BrittleParser) Parse() {
	p.Find("table").Each(func(_ int, tablehtml *goquery.Selection) {
		tablehtml.Find("tr").Each(p.parseRow)
	})
}

// Parse a row in the table. This can yield a text (accumulated on p.parsed),
// yield nothing for a date-header row, or halt the caller process upon
// unparseable rows.
func (p *BrittleParser) parseRow(rowIndex int, rowhtml *goquery.Selection) {
	if rowIndex == 0 {
		log.Print("Skipping header row of table")
		return
	}

	children := rowhtml.Children().Nodes

	switch count := len(children); count {
	case 4:
		text := new(text.Text)

		linkTD, authorTD, noteTD, dateTD := children[0], children[1], children[2], children[3]

		text.Author = authorTD.FirstChild.Data
		text.Note = noteTD.FirstChild.Data

		timestamp, err := time.Parse("January 2, 2006", dateTD.FirstChild.Data)
		if err != nil {
			log.Fatalf("Unparseable date: %v", dateTD.FirstChild.Data)
		}
		text.Timestamp = timestamp

		linkA := linkTD.FirstChild
		text.Title = linkA.FirstChild.Data
		for _, attr := range linkA.Attr {
			if attr.Key == "href" {
				text.URL = attr.Val
			}
		}

		if err := text.Validate(); err != nil {
			log.Fatalf("Row %v is invalid text: %v", rowIndex, err)
		}
		p.parsed = append(p.parsed, text)
	case 1:
		log.Printf("Row %v is a date header; skipping", rowIndex)
	default:
		log.Fatalf("Row %v has unhandled format: has %v children", rowIndex, count)
	}
}
