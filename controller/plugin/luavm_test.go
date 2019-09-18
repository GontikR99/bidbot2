package plugin

import (
	"context"
	"github.com/gontikr99/bidbot2/controller/everquest"
	"log"
	"testing"
	"time"
)

type constRecord struct {
	records map[string]*everquest.GuildRecord
}

func (cr *constRecord) GuildRecords() (gr map[string]*everquest.GuildRecord, err error) {
	gr = cr.records
	return
}

type dummyWebCache struct {
}

func (*dummyWebCache) FetchHTTP(string, time.Duration) (text string, err error) {
	text = "<a href=\"\">Jephine</a></td><td ><span class='dkp_current'>500.00</span></td></tr>" +
		"<a href=\"\">Dalamin</a></td><td ><span class='dkp_current'>501.00</span></td></tr>" +
		"<a href=\"\">Piddles</a></td><td ><span class='dkp_current'>502.00</span></td></tr>"
	return
}

func buildDummyGuildPlugin(t *testing.T) (*GuildPlugin, func()) {
	ctx, done := context.WithCancel(context.Background())
	vm, err := NewGuildPlugin(ctx, &dummyWebCache{}, "../../plugins/modusgelidus.lua")
	if err != nil {
		t.Fatal(t)
	}
	cr := &constRecord{make(map[string]*everquest.GuildRecord)}

	cr.records["dalamin"] = &everquest.GuildRecord{
		Level:     65,
		Class:     "cleric",
		Rank:      "leader",
		IsAlt:     false,
		GuildNote: "The Boss",
	}
	cr.records["larryy"] = &everquest.GuildRecord{
		Level:     65,
		Class:     "bard",
		Rank:      "officer alt/box",
		IsAlt:     true,
		GuildNote: "Dalamin",
	}
	cr.records["piddles"] = &everquest.GuildRecord{
		Level:     65,
		Class:     "warrior",
		Rank:      "officer",
		IsAlt:     false,
		GuildNote: "Loot Team",
	}
	cr.records["widdles"] = &everquest.GuildRecord{
		Level:     65,
		Class:     "bard",
		Rank:      "officer box/alt",
		IsAlt:     true,
		GuildNote: "Piddles",
	}
	cr.records["jephine"] = &everquest.GuildRecord{
		Level:     65,
		Class:     "shadow knight",
		Rank:      "officer",
		IsAlt:     false,
		GuildNote: "Loot Team",
	}
	cr.records["joramar"] = &everquest.GuildRecord{
		Level:     65,
		Class:     "beastlord",
		Rank:      "officer box/alt",
		IsAlt:     true,
		GuildNote: "Jephine",
	}
	cr.records["geoffrey"] = &everquest.GuildRecord{
		Level:     65,
		Class:     "shaman",
		Rank:      "officer box/alt",
		IsAlt:     true,
		GuildNote: "Jephine",
	}
	cr.records["suuloti"] = &everquest.GuildRecord{
		Level:     65,
		Class:     "rogue",
		Rank:      "officer box/alt",
		IsAlt:     false,
		GuildNote: "Jephine",
	}
	vm.guildRecordReader = cr

	return vm, done
}

func TestGuildPlugin_GetMain(t *testing.T) {
	vm, done := buildDummyGuildPlugin(t)
	defer done()
	if result, err := vm.GetMain("Jephine"); err != nil || result != "jephine" {
		t.Fatalf("Expected jephine, got %v", result)
	}
	if result, err := vm.GetMain("Joramar"); err != nil || result != "jephine" {
		t.Fatalf("Expected jephine, got %v", result)
	}
	if result, err := vm.GetMain("Geoffrey"); err != nil || result != "jephine" {
		t.Fatalf("Expected jephine, got %v", result)
	}
	if result, err := vm.GetMain("Suuloti"); err != nil || result != "jephine" {
		t.Fatalf("Expected jephine, got %v", result)
	}
	if result, err := vm.GetMain("Piddles"); err != nil || result != "piddles" {
		t.Fatalf("Expected piddles, got %v", result)
	}
	if result, err := vm.GetMain("Widdles"); err != nil || result != "piddles" {
		t.Fatalf("Expected piddles, got %v", result)
	}
}

func TestGuildPlugin_GetDKP(t *testing.T) {
	vm, done := buildDummyGuildPlugin(t)
	defer done()

	if result, err := vm.GetDKP("Joramar"); err != nil || result != 500 {
		t.Fatalf("Expected 500, got %v", result)
	}
	if result, err := vm.GetDKP("Dalamin"); err != nil || result != 501 {
		t.Fatalf("Expected 501, got %v", result)
	}
	if result, err := vm.GetDKP("Widdles"); err != nil || result != 502 {
		t.Fatalf("Expected 502, got %v", result)
	}
}

func TestGuildPlugin_SortBids(t *testing.T) {
	vm, done := buildDummyGuildPlugin(t)
	defer done()

	bids := map[string]float64{
		"Jephine": 101,
		"Joramar": 101,
	}
	price, winners, _, err := vm.SortBids(bids, 1)
	if err != nil {
		t.Fatal(err)
	}
	if price != 101 {
		t.Fatalf("Expected price of 101, got price of %v", price)
	}
	if len(winners) != 1 || winners[0] != "jephine" {
		t.Fatalf("Expected Jephine to win")
	}

	bids = map[string]float64{
		"Dalamin": 200,
		"Jephine": 150,
	}
	price, winners, _, err = vm.SortBids(bids, 1)
	if err != nil {
		t.Fatal(err)
	}
	if price != 151 {
		t.Fatalf("Expected price of 101, got price of %v", price)
	}
	if len(winners) != 1 || winners[0] != "dalamin" {
		t.Fatalf("Expected Dalamin to win")
	}

	bids = map[string]float64{
		"Dalamin": 200,
		"Jephine": 200,
		"Piddles": 100,
	}
	price, winners, _, err = vm.SortBids(bids, 1)
	if err != nil {
		t.Fatal(err)
	}
	if price != 200 {
		t.Fatalf("Expected price of 200, got price of %v", price)
	}
	if len(winners) != 2 || winners[0] != "dalamin" || winners[1] != "jephine" {
		t.Fatalf("Expected Dalamin+Jephine to win")
	}

	bids = map[string]float64{
		"Larryy":  200,
		"Joramar": 110,
		"Piddles": 75,
	}
	price, winners, _, err = vm.SortBids(bids, 1)
	if err != nil {
		t.Fatal(err)
	}
	if price != 111 {
		t.Fatalf("Expected price of 111, got price of %v", price)
	}
	if len(winners) != 1 || winners[0] != "larryy" {
		t.Fatalf("Expected Larryy to win")
	}

	bids = map[string]float64{
		"Geoffrey": 200,
	}
	price, winners, _, err = vm.SortBids(bids, 1)
	if err != nil {
		t.Fatal(err)
	}
	if price != 5 {
		t.Fatalf("Expected price of 5, got price of %v", price)
	}
	if len(winners) != 1 || winners[0] != "geoffrey" {
		t.Fatalf("Expected Larryy to win")
	}

	bids = map[string]float64{
		"Larryy":  200,
		"Joramar": 110,
		"Piddles": 75,
	}
	var display []BidDesc
	price, winners, display, err = vm.SortBids(bids, 2)
	log.Println(price, winners, display)
	if err != nil {
		t.Fatal(err)
	}
	if price != 76 {
		t.Fatalf("Expected price of 76, got price of %v", price)
	}
	if len(winners) != 2 || winners[0] != "larryy" || winners[1] != "joramar" {
		t.Fatalf("Expected Joramar+Larryy to win")
	}
}
