package main

import (
	"testing"
	"time"

	"github.com/golang/protobuf/proto"

	pb "github.com/brotherlogic/cardserver/card"
)

var testdata = []struct {
	cronline string
	card     []pb.Card
}{
	{"2017-02-03 00:00~githubissueadd~Made Up Title~Made Up Test~component", []pb.Card{pb.Card{Text: "Made Up Title|Made Up Test", Action: pb.Card_DISMISS, ApplicationDate: 1486080000, Priority: -1, Hash: "githubissueadd-component"}}},
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
	end, _ := time.Parse("2006-01-02 15:04", "2018-01-01 00:00")
	cards := c.GetCards(end)
	if len(cards) != 1 {
		t.Errorf("Failure to pull correct number of cards: %v", cards)
	}

	c2 := Init(".testreload")
	c2.loadline(testdata[0].cronline)
	cards = c2.GetCards(end)
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
		curr, _ := time.Parse("2006-01-02 15:04", "2017-01-01 00:00")
		end, _ := time.Parse("2006-01-02 15:04", "2018-01-01 00:00")

		var cards []*pb.Card
		for curr.Before(end) {
			cards = append(cards, c.GetCards(curr)...)

			curr = curr.Add(time.Minute * 59)
		}

		if len(cards) != len(test.card) {
			t.Errorf("Too many cards written %v vs %v", len(cards), len(test.card))
		} else {
			for i := range cards {
				if !proto.Equal(&test.card[i], cards[i]) {
					t.Errorf("Mismatch in cards %v vs %v", cards[i], test.card[i])
				}
			}
		}
	}
}
