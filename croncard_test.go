package main

import (
	"log"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"

	pb "github.com/brotherlogic/cardserver/card"
)

var testdata = []struct {
	cronline string
	card     []pb.Card
}{
	{"2017-02-03 00:00~githubissueadd~Made Up Title~Made Up Test~component", []pb.Card{pb.Card{Text: "Made Up Title|Made Up Test", Action: pb.Card_DISMISS, ApplicationDate: getUnixTime("2017-02-03 00:00"), Priority: -1, Hash: "githubissueadd-Made Up Title-component"}}},
}

var testcounts = []struct {
	cronline []string
	count    int
}{
	{[]string{"Wed~githubissueadd~Made Up Title~Made Up Test~component"}, 52},
	{[]string{"Daily~githubissueadd~Made Up Title~Made Up Test~component"}, 365},
	{[]string{"Daily~addgithubissue~Record Check~Bedroom~home", "Daily~addgithubissue~Clear Email~Bedroom~home"}, 365 * 2},
}

func TestCounts(t *testing.T) {
	for _, test := range testcounts {
		c := Init(".testcronforcounting")
		c.clearhash()
		for _, line := range test.cronline {
			c.loadline(line)
		}
		c.logd()

		//Run through the whole of 2017, at random
		curr, _ := getTime("2017-01-01 00:01")
		prev, _ := getTime("2017-01-01 00:00")
		end, _ := getTime("2018-01-01 00:00")

		var cards []*pb.Card
		for curr.Before(end) {
			cards = append(cards, c.GetCards(prev, curr)...)

			prev = curr
			curr = curr.Add(time.Minute * 59)
		}

		if len(cards) != test.count {
			t.Errorf("Wrong number of cards written %v vs %v", len(cards), test.count)
		}
	}
}

func TestTimeParse(t *testing.T) {
	t1 := time.Now()
	log.Printf("Now = %v", t1)
	t1 = t1.Round(time.Minute)
	tstr := t1.Format(datestr)
	t2, err := getTime(tstr)

	if err != nil {
		t.Errorf("Parsing failed: %v", err)
	}
	if !t1.Equal(t2) {
		t.Errorf("Parsing time has failed: %v vs %v", t1, t2)
	}
}

func TestCronLoad(t *testing.T) {
	c := InitFromFile(".testcronload", "crontest.txt")
	c.clearhash()
	if len(c.crons) != 1 {
		t.Errorf("Init has failed %v", c.crons)
	}
}

func TestNoDoubleOnReload(t *testing.T) {
	c := Init(".testreload")
	c.clearhash()
	c.loadline(testdata[0].cronline)
	start, _ := getTime("2017-01-01 00:00")
	end, _ := getTime("2018-01-01 00:00")
	cards := c.GetCards(start, end)
	if len(cards) != 1 {
		t.Errorf("Failure to pull correct number of cards: %v", cards)
	}

	c2 := Init(".testreload")
	c2.loadline(testdata[0].cronline)
	cards = c2.GetCards(start, end)
	if len(cards) != 0 {
		t.Errorf("Failure to pull correct number of cards on reload: %v", cards)
	}
}

func TestNoDoubleOnReloadWithDaily(t *testing.T) {
	c := Init(".testreload2")
	c.clearhash()
	c.loadline(testcounts[1].cronline[0])
	start, _ := getTime("2017-01-01 00:00")
	end, _ := getTime("2017-01-01 23:00")
	cards := c.GetCards(start, end)
	log.Printf("HERE1 = %v", cards)
	if len(cards) != 1 {
		t.Errorf("Failure to pull correct number of cards: %v", cards)
	}

	c2 := Init(".testreload2")
	c2.last = time.Unix(0, 0)
	c2.loadline(testcounts[1].cronline[0])
	cards = c2.GetCards(start, end)
	log.Printf("HERE2 = %v", cards)
	if len(cards) != 0 {
		t.Errorf("Failure to pull correct number of cards on reload: %v", cards)
	}
}

func TestCron(t *testing.T) {
	for _, test := range testdata {
		c := Init(".testcron")
		c.clearhash()
		c.loadline(test.cronline)
		c.logd()

		//Run through the whole of 2017, at random
		curr, _ := getTime("2017-01-01 00:01")
		prev, _ := getTime("2017-01-01 00:00")
		end, _ := getTime("2018-01-01 00:00")

		var cards []*pb.Card
		for curr.Before(end) {
			cards = append(cards, c.GetCards(prev, curr)...)

			prev = curr
			curr = curr.Add(time.Minute * 59)
		}

		if len(cards) != len(test.card) {
			t.Errorf("Wrong nubmer of cards written %v vs %v", len(cards), len(test.card))
		} else {
			for i := range cards {
				if !proto.Equal(&test.card[i], cards[i]) {
					t.Errorf("Mismatch in cards %v vs %v", cards[i], test.card[i])
				}
			}
		}
	}
}
