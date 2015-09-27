package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	Q "github.com/PuerkitoBio/goquery"
	"github.com/olekukonko/tablewriter"
	"github.com/peterh/liner"
	"github.com/vrischmann/shlex"
)

func makeURL(name string, searchPage int) string {
	if searchPage == 0 {
		return fmt.Sprintf("https://kat.cr/usearch/%s", name)
	}

	return fmt.Sprintf("https://kat.cr/usearch/%s/%d", name, searchPage)
}

var (
	links = make(map[int]string)
)

func search(args []string) (err error) {
	var searchPage int

	name := args[0]
	if len(args) > 1 {
		searchPage, err = strconv.Atoi(args[1])
		if err != nil {
			return
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "Name", "Size", "Age", "Seeds"})
	table.SetColWidth(200)

	url := makeURL(name, searchPage)

	var doc *Q.Document
	{
		doc, err = Q.NewDocument(url)
		if err != nil {
			return
		}
	}

	doc.Find(".torrentname").Each(func(i int, s *Q.Selection) {
		row := s.Parent().Parent()

		cell := row.Find(".cellMainLink")

		{
			link, _ := cell.Attr("href")
			link = "https://kat.cr" + link
			links[i] = link
		}

		name := cell.Text()

		size := row.Find("td:nth-child(2)").Text()
		age := row.Find("td:nth-child(4)").Text()
		seeds := row.Find("td:nth-child(5)").Text()

		table.Append([]string{fmt.Sprintf("%d", i), name, size, age, seeds})
	})
	if err != nil {
		return
	}

	table.Render()

	return
}

var (
	errMagnetTakesANumber = errors.New("magnet takes a single argument which is the number of the torrent")
)

func magnet(args []string) (err error) {
	if len(args) < 1 {
		return errMagnetTakesANumber
	}

	i, err := strconv.Atoi(args[0])
	if err != nil {
		return errMagnetTakesANumber
	}

	link, ok := links[i]
	if !ok {
		return fmt.Errorf("torrent %d does not exist", i)
	}

	doc, err := Q.NewDocument(link)
	if err != nil {
		return err
	}

	magnet, ok := doc.Find(".magnetlinkButton").Attr("href")
	if !ok {
		return errors.New("the magnet link does not exist")
	}

	fmt.Println(magnet)

	return
}

func main() {
	flag.Parse()

	line := liner.NewLiner()
	line.SetCtrlCAborts(true)
	defer line.Close()

	for {
		cmd, err := line.Prompt("> ")
		if err == liner.ErrPromptAborted || err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		args := shlex.Parse(cmd)

		switch strings.ToLower(args[0]) {
		case "search":
			if err := search(args[1:]); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
		case "magnet":
			if err := magnet(args[1:]); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
		}
	}
}
