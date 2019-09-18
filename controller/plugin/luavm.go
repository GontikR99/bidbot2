package plugin

import (
	"context"
	"errors"
	"fmt"
	"github.com/gontikr99/bidbot2/controller/discord"
	"github.com/gontikr99/bidbot2/controller/everquest"
	"github.com/gontikr99/bidbot2/controller/storage"
	"github.com/yuin/gluare"
	lua "github.com/yuin/gopher-lua"
	"log"
	"math"
	"strings"
)

type GuildPlugin struct {
	state             *lua.LState
	context           context.Context
	web               storage.WebCache
	guildRecordReader everquest.GuildRecordsReader
	discord           *discord.Client

	dkpFunc         lua.LValue
	mainFunc        lua.LValue
	validateBidFunc lua.LValue
	sortBidsFunc    lua.LValue
	solicitFunc     lua.LValue
	actions         chan<- luaRequest
}

const guildPluginKey = "guildPlugin"

func guildPlugin(state *lua.LState) *GuildPlugin {
	regTable := state.Get(lua.RegistryIndex)
	ud := state.GetField(regTable, guildPluginKey).(*lua.LUserData)
	return ud.Value.(*GuildPlugin)
}

func NewGuildPlugin(ctx context.Context, web storage.WebCache, sourcePath string) (gp *GuildPlugin, err error) {
	log.Println("Compiling plugin " + sourcePath)
	return newGuildPlugin(ctx, web, func(state *lua.LState) error {
		return state.DoFile(sourcePath)
	})
}

func newGuildPlugin(ctx context.Context, web storage.WebCache, sourceRunner func(state *lua.LState) error) (gp *GuildPlugin, err error) {
	result := &GuildPlugin{}
	result.context = ctx
	result.web = web
	result.state = lua.NewState()

	regTable := result.state.Get(lua.RegistryIndex)
	gpUserData := result.state.NewUserData()
	gpUserData.Value = result
	result.state.SetField(regTable, guildPluginKey, gpUserData)

	result.state.PreloadModule("re", gluare.Loader)
	result.state.PreloadModule("http", httpLoader)
	result.state.PreloadModule("everquest", eqLoader)
	result.state.SetGlobal("print", result.state.NewFunction(logPrint))

	err = sourceRunner(result.state)
	if err != nil {
		return
	}

	result.dkpFunc = result.state.GetGlobal("getdkp")
	if result.dkpFunc == nil {
		err = errors.New("The plugin doesn't define a function named 'getdkp'")
		result.state.Close()
		return
	}
	result.mainFunc = result.state.GetGlobal("getmain")
	if result.mainFunc == nil {
		err = errors.New("The plugin doesn't define a function named 'getmain'")
		result.state.Close()
		return
	}
	result.validateBidFunc = result.state.GetGlobal("validatebid")
	if result.validateBidFunc == nil {
		err = errors.New("The plugin doesn't define a function named 'validatebid'")
		result.state.Close()
		return
	}

	result.sortBidsFunc = result.state.GetGlobal("sortbids")
	if result.sortBidsFunc == nil {
		err = errors.New("The plugin doesn't define a function named 'sortbids'")
		result.state.Close()
		return
	}

	result.solicitFunc = result.state.GetGlobal("solicit")
	if result.sortBidsFunc == nil {
		err = errors.New("The plugin doesn't define a function named 'solicit'")
		result.state.Close()
		return
	}

	result.state.SetContext(ctx)
	actChan := make(chan luaRequest)
	result.actions = actChan

	log.Println("Starting plugin loop")
	go result.luaLoop(actChan)

	gp = result
	return
}

func (gp *GuildPlugin) submit(callback func() (lua.LValue, error)) (lua.LValue, error) {
	respChan := make(chan luaResponse)
	req := luaRequest{callback, respChan}
	gp.actions <- req
	select {
	case <-gp.context.Done():
		return nil, errors.New("Closed Lua before action finished")
	case resp := <-respChan:
		return resp.value, resp.err
	}
}

func (gp *GuildPlugin) SetEqClient(everquest *everquest.Client) {
	_, _ = gp.submit(func() (lua.LValue, error) {
		gp.guildRecordReader = everquest
		return nil, nil
	})
}

func (gp *GuildPlugin) SetDiscordClient(discord *discord.Client) {
	_, _ = gp.submit(func() (lua.LValue, error) {
		gp.discord = discord
		return nil, nil
	})
}

func (gp *GuildPlugin) GetMain(charname string) (string, error) {
	value, err := gp.submit(func() (lua.LValue, error) {
		err := gp.state.CallByParam(lua.P{
			Fn:      gp.mainFunc,
			NRet:    1,
			Protect: true,
		}, lua.LString(strings.ToLower(charname)))
		if err != nil {
			return nil, err
		}
		ret := gp.state.Get(-1)
		gp.state.Pop(1)
		return ret, nil
	})
	if err != nil {
		return "", err
	}
	if value.Type() == lua.LTString {
		return strings.ToLower(lua.LVAsString(value)), nil
	} else if value.Type() == lua.LTNil {
		return "", nil
	} else {
		return "", errors.New("getmain function didn't return a string or nil")
	}
}

func (gp *GuildPlugin) GetDKP(charname string) (float64, error) {
	value, err := gp.submit(func() (lua.LValue, error) {
		err := gp.state.CallByParam(lua.P{
			Fn:      gp.dkpFunc,
			NRet:    1,
			Protect: true,
		}, lua.LString(strings.ToLower(charname)))
		if err != nil {
			return nil, err
		}
		ret := gp.state.Get(-1)
		gp.state.Pop(1)
		return ret, nil
	})
	if err != nil {
		return 0, err
	}

	if value.Type() == lua.LTNumber {
		return float64(lua.LVAsNumber(value)), nil
	} else if value.Type() == lua.LTNil {
		return math.NaN(), nil
	} else {
		log.Printf("Expected getdkp to return a number or nil, but it returned a %v", value.Type())
		return 0, errors.New("getdkp function didn't return a number or nil")
	}
}

func (gp *GuildPlugin) ValidateBid(charname string, bid float64) (string, error) {
	value, err := gp.submit(func() (lua.LValue, error) {
		err := gp.state.CallByParam(lua.P{
			Fn:      gp.validateBidFunc,
			NRet:    1,
			Protect: true,
		}, lua.LString(strings.ToLower(charname)), lua.LNumber(bid))
		if err != nil {
			return nil, err
		}
		ret := gp.state.Get(-1)
		gp.state.Pop(1)
		return ret, nil
	})
	if err != nil {
		return "", err
	}
	if value.Type() == lua.LTNil {
		return "", nil
	} else if value.Type() == lua.LTString {
		return lua.LVAsString(value), nil
	} else {
		return "", errors.New("validatebid function didn't return a string or nil")
	}
}

func (gp *GuildPlugin) Solicit(itemName string) (string, error) {
	value, err := gp.submit(func() (lua.LValue, error) {
		err := gp.state.CallByParam(lua.P{
			Fn:      gp.solicitFunc,
			NRet:    1,
			Protect: true,
		}, lua.LString(itemName))
		if err != nil {
			return nil, err
		}
		ret := gp.state.Get(-1)
		gp.state.Pop(1)
		return ret, nil
	})
	if err != nil {
		return "", err
	}
	if value.Type() == lua.LTString {
		return strings.ToLower(lua.LVAsString(value)), nil
	} else {
		return "", errors.New("solicit function didn't return a string")
	}
}

type BidDesc struct {
	BidderDesc string
	BidDesc    string
}

func (gp *GuildPlugin) SortBids(rawBids map[string]float64, count int) (price float64, winners []string, displayBids []BidDesc, err error) {
	_, _ = gp.submit(func() (lua.LValue, error) {
		rawBidsTable := gp.state.NewTable()
		for bidder, bid := range rawBids {
			gp.state.SetField(rawBidsTable, strings.ToLower(bidder), lua.LNumber(bid))
		}
		err = gp.state.CallByParam(lua.P{
			Fn:      gp.sortBidsFunc,
			NRet:    3,
			Protect: true,
		}, rawBidsTable, lua.LNumber(count))
		if err != nil {
			return nil, err
		}
		var winTable, displayTable lua.LValue
		price, winTable, displayTable = float64(lua.LVAsNumber(gp.state.Get(-3))), gp.state.Get(-2), gp.state.Get(-1)
		gp.state.Pop(3)

		if winTable.Type() != lua.LTTable {
			return nil, fmt.Errorf("SortBids expected second return value to be a table, but it was %v", displayTable.Type())
		}
		if displayTable.Type() != lua.LTTable {
			return nil, fmt.Errorf("SortBids expected third return value to be a table, but it was %v", displayTable.Type())
		}

		winners = make([]string, 0)
		for i := 1; ; i++ {
			winner := gp.state.GetTable(winTable, lua.LNumber(i))
			if winner == lua.LNil {
				break
			}
			winners = append(winners, strings.ToLower(lua.LVAsString(winner)))
		}

		displayBids = make([]BidDesc, 0)
		for i := 1; ; i++ {
			displayEntry := gp.state.GetTable(displayTable, lua.LNumber(i))
			if displayEntry == lua.LNil {
				break
			}
			if displayEntry.Type() != lua.LTTable {
				return nil, fmt.Errorf("SortBids expected third return value to contain tables, but found a %v", displayEntry.Type())
			}
			bidDescEntry := BidDesc{
				BidderDesc: lua.LVAsString(gp.state.GetTable(displayEntry, lua.LNumber(1))),
				BidDesc:    lua.LVAsString(gp.state.GetTable(displayEntry, lua.LNumber(2))),
			}
			displayBids = append(displayBids, bidDescEntry)
		}
		return nil, nil
	})
	return
}

func (gp *GuildPlugin) luaLoop(requests <-chan luaRequest) {
	for {
		select {
		case <-gp.context.Done():
			gp.state.Close()
			log.Printf("Ending plugin loop")
			return
		case req := <-requests:
			value, err := req.action()
			req.respChan <- luaResponse{value, err}
		}
	}
}

type luaRequest struct {
	action   func() (lua.LValue, error)
	respChan chan<- luaResponse
}

type luaResponse struct {
	value lua.LValue
	err   error
}

func logPrint(state *lua.LState) int {
	text := make([]byte, 0)
	top := state.GetTop()
	for i := 1; i <= top; i++ {
		text = append(text, []byte(state.ToStringMeta(state.Get(i)).String())...)
		if i != top {
			text = append(text, '\t')
		}
	}
	log.Println(string(text))
	return 0
}
