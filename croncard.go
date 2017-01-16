package main

import (
	"log"
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
}

// Init prepares a cron for use
func Init() *Cron {
	c := &Cron{}
	c.last = time.Unix(0, 0)
	return c
}

//GetCards gets the cards up to the specified time
func (c *Cron) GetCards(t time.Time) []pb.Card {
	newindex := 0
	var cards []pb.Card
	for i, entry := range c.crons {
		if entry.time.Before(t) {
			newindex = i + 1
			cards = append(cards, pb.Card{Text: entry.text, Action: pb.Card_DISMISS, ApplicationDate: entry.time.Unix(), Priority: -1, Hash: entry.hash})
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
	entry.text = elems[2] + "\n" + elems[3]
	entry.hash = elems[1] + "-" + elems[4]

	c.crons = append(c.crons, entry)
}
