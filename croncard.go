package main

import (
	"bufio"
	"hash/fnv"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	pb "github.com/brotherlogic/cardserver/card"
)

type cronentry struct {
	time time.Time
	text string
	hash string
}

// Cron the main cronentry holder
type Cron struct {
	crons []cronentry
	last  time.Time
	dir   string
}

// Init prepares a cron for use
func Init(dir string) *Cron {
	c := &Cron{}
	c.last = time.Unix(0, 0)
	c.dir = dir

	if _, err := os.Stat(c.dir); os.IsNotExist(err) {
		os.MkdirAll(c.dir, 0777)
	}

	return c
}

// InitFromFile loads a cron from a given file
func InitFromFile(dir string, filename string) *Cron {
	c := Init(dir)

	file, _ := os.Open(filename)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		c.loadline(scanner.Text())
	}

	//if err := scanner.Err(); err != nil {
	//	log.Fatal(err)
	//}

	return c
}

func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return strconv.Itoa(int(h.Sum32()))
}

func (c *Cron) clearhash() {
	os.Remove(c.dir + "/hash")
}

func (c *Cron) writehash(card pb.Card) {
	if _, err := os.Stat(c.dir + "/hash"); os.IsNotExist(err) {
		os.Create(c.dir + "/hash")
	}

	f, err := os.OpenFile(c.dir+"/hash", os.O_APPEND|os.O_WRONLY, 0600)
	log.Printf("ERROR = %v", err)
	defer f.Close()
	f.WriteString(hash(card.String()) + "\n")
}

func (c *Cron) isWritten(card pb.Card) bool {
	file, _ := os.Open(c.dir + "/hash")
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		log.Printf("COMP %v and %v", scanner.Text(), hash(card.String()))
		if scanner.Text() == hash(card.String()) {
			return true
		}
	}

	return false
}

//GetCards gets the cards up to the specified time
func (c *Cron) GetCards(t time.Time) []*pb.Card {
	newindex := 0
	var cards []*pb.Card
	for i, entry := range c.crons {
		if entry.time.Before(t) {
			newindex = i + 1
			card := pb.Card{Text: entry.text, Action: pb.Card_DISMISS, ApplicationDate: entry.time.Unix(), Priority: -1, Hash: entry.hash}
			if !c.isWritten(card) {
				cards = append(cards, &card)
				c.writehash(card)
			}
		}
	}
	c.crons = c.crons[newindex:]
	return cards
}

func (c *Cron) logd() {
	log.Printf("LEN = %v", len(c.crons))
}

func (c *Cron) loadline(line string) {
	elems := strings.Split(line, "~")
	entry := cronentry{}
	entry.time, _ = time.Parse("2006-01-02 15:04", elems[0])
	entry.text = elems[2] + "|" + elems[3]
	entry.hash = elems[1] + "-" + elems[4]

	c.crons = append(c.crons, entry)
}
