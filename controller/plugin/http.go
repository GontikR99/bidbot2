package plugin

import (
	lua "github.com/yuin/gopher-lua"
	"time"
)

func fetchHttp(state *lua.LState) int {
	lv := state.CheckString(1)
	if lv == "" {
		panic("Specify a page")
	}
	text, err := guildPlugin(state).web.FetchHTTP(lv, 5*time.Minute)
	if err != nil {
		panic(err)
	}
	state.Push(lua.LString(text))
	return 1
}

var httpExports = map[string]lua.LGFunction{
	"get": fetchHttp,
}

func httpLoader(state *lua.LState) int {
	mod := state.SetFuncs(state.NewTable(), httpExports)
	state.Push(mod)
	return 1
}
