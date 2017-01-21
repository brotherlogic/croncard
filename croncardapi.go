package main

import (
	"flag"
	"log"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

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
	cards := c.GetCards(c.last, time.Now())

	var host = flag.String("host", "10.0.1.17", "Hostname of server.")
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
