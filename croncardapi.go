package main

import (
	"flag"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	pbc "github.com/brotherlogic/cardserver/card"
	pbdi "github.com/brotherlogic/discovery/proto"
)

func getIP(servername string, ip string, port int) (string, int) {
	conn, _ := grpc.Dial(ip+":"+strconv.Itoa(port), grpc.WithInsecure())
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceClient(conn)
	entry := pbdi.RegistryEntry{Name: servername}
	r, _ := registry.Discover(context.Background(), &entry)
	return r.Ip, int(r.Port)
}

func main() {
	c := InitFromFile("crontstore", "cron")
	dryRun := flag.Bool("dry_run", false, "Don't write anything.")
	quiet := flag.Bool("quiet", true, "Don't log owt.")
	flag.Parse()

	if *quiet {
		log.SetOutput(ioutil.Discard)
		grpclog.SetLogger(log.New(ioutil.Discard, "", -1))
	}

	cards := c.GetCards(c.last, time.Now())

	if *dryRun {
		log.Printf("Would write: %v", cards)
	} else {
		var host = flag.String("host", "192.168.86.34", "Hostname of server.")
		var port = flag.Int("port", 50055, "Port number of server")

		cServer, cPort := getIP("cardserver", *host, *port)
		conn, err := grpc.Dial(cServer+":"+strconv.Itoa(cPort), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Failure to dial cardserver (%v)", err)
		}
		defer conn.Close()
		client := pbc.NewCardServiceClient(conn)

		_, err = client.AddCards(context.Background(), &pbc.CardList{Cards: cards})
		if err != nil {
			log.Fatalf("Failure to add cards: %v", err)
		}
	}
}
