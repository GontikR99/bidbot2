package plugin

import lua "github.com/yuin/gopher-lua"

func getMembers(state *lua.LState) int {
	gp := guildPlugin(state)
	if gp.guildRecordReader == nil {
		panic("EverQuest is not yet present")
	}
	grd, err := gp.guildRecordReader.GuildRecords()
	if err != nil {
		panic(err)
	}
	grMap := state.NewTable()
	for name, gr := range grd {
		grTable := state.NewTable()
		state.SetField(grTable, "alt", lua.LBool(gr.IsAlt))
		state.SetField(grTable, "level", lua.LNumber(gr.Level))
		state.SetField(grTable, "note", lua.LString(gr.GuildNote))
		state.SetField(grTable, "class", lua.LString(gr.Class))
		state.SetField(grTable, "rank", lua.LString(gr.Rank))
		state.SetField(grTable, "lastonline", lua.LString(gr.LastOnline))
		state.SetField(grMap, name, grTable)
	}
	state.Push(grMap)
	return 1
}

var eqExports = map[string]lua.LGFunction{
	"guildmembers": getMembers,
}

func eqLoader(state *lua.LState) int {
	mod := state.SetFuncs(state.NewTable(), eqExports)
	state.Push(mod)
	return 1
}
