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

const (
	datestr = "2006-01-02 15:04"
)

var (
	daysOfTheWeek = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
)

func getTime(timestr string) (time.Time, error) {
	t := time.Now()
	return time.ParseInLocation(datestr, timestr, t.Location())
}

func getUnixTime(timestr string) int64 {
	t, _ := getTime(timestr)
	return t.Unix()
}

type cronentry struct {
	time  *time.Time
	daily bool
	day   string
	text  string
	hash  string
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

	f, _ := os.OpenFile(c.dir+"/hash", os.O_APPEND|os.O_WRONLY, 0600)
	defer f.Close()
	f.WriteString(hash(card.String()) + "\n")
}

func (c *Cron) isWritten(card pb.Card) bool {
	file, _ := os.Open(c.dir + "/hash")
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text() == hash(card.String()) {
			return true
		}
	}

	return false
}

//GetCards gets the cards between the specified times
func (c *Cron) GetCards(ts time.Time, te time.Time) []*pb.Card {
	newindex := 0
	var cards []*pb.Card
	for i, entry := range c.crons {
		if entry.time != nil {
			if (entry.time.Before(te) || entry.time.Equal(te)) && entry.time.After(ts) {
				newindex = i + 1
				card := pb.Card{Text: entry.text, Action: pb.Card_DISMISS, ApplicationDate: entry.time.Unix(), Priority: -1, Hash: entry.hash}
				if !c.isWritten(card) {
					cards = append(cards, &card)
					c.writehash(card)
				}
			}
		} else if entry.day != "" {
			// Hack central
			stime := ts
			stime = stime.Add(-time.Hour * time.Duration(stime.Hour()))
			stime = stime.Add(-time.Minute * time.Duration(stime.Minute()))
			stime = stime.Add(-time.Second * time.Duration(stime.Second()))

			count := 1
			for stime.Before(te) {
				if stime.Format("Mon") == entry.day {
					card := pb.Card{Text: entry.text, Action: pb.Card_DISMISS, ApplicationDate: stime.Unix(), Priority: -1, Hash: entry.hash}
					if !c.isWritten(card) {
						cards = append(cards, &card)
						c.writehash(card)
					}
				}

				count++
				stime = stime.Add(time.Hour * 24)
			}
		} else if entry.daily {
			// Hack central
			stime := ts
			stime = stime.Add(-time.Hour * time.Duration(stime.Hour()))
			stime = stime.Add(-time.Minute * time.Duration(stime.Minute()))
			stime = stime.Add(-time.Second * time.Duration(stime.Second()))
			stime = stime.Add(time.Hour * time.Duration(5))

			count := 1
			for stime.Before(te) {
				card := pb.Card{Text: entry.text, Action: pb.Card_DISMISS, ApplicationDate: stime.Unix(), Priority: -1, Hash: entry.hash}
				if !c.isWritten(card) {
					log.Printf("%v - %v", stime, entry.hash)
					cards = append(cards, &card)
					c.writehash(card)
				}

				count++
				stime = stime.Add(time.Hour * 24)
			}
		}
	}
	c.crons = c.crons[newindex:]
	return cards
}

func (c *Cron) logd() {
	log.Printf("LEN = %v", len(c.crons))
}

func matches(s string, strs []string) bool {
	for _, str := range strs {
		if s == str {
			return true
		}
	}
	return false
}

func (c *Cron) loadline(line string) {
	elems := strings.Split(line, "~")
	entry := cronentry{}
	if matches(elems[0], daysOfTheWeek) {
		entry.day = elems[0]
	} else if elems[0] == "Daily" {
		entry.daily = true
	} else {
		t, _ := getTime(elems[0])
		entry.time = &t
	}
	entry.text = elems[2] + "|" + elems[3]
	entry.hash = elems[1] + "-" + elems[4]

	c.crons = append(c.crons, entry)
}
