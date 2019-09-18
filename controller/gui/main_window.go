package gui

import (
	"context"
	"encoding/json"
	"github.com/gontikr99/bidbot2/controller/everquest"
	storage2 "github.com/gontikr99/bidbot2/controller/storage"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

type CharModel struct {
	walk.ListModelBase
	items []string
}

func NewCharModel() *CharModel { return &CharModel{items: []string{}} }

func (m *CharModel) ItemCount() int              { return len(m.items) }
func (m *CharModel) Value(index int) interface{} { return m.items[index] }
func (m *CharModel) Len() int                    { return m.ItemCount() }
func (m *CharModel) Less(i, j int) bool          { return strings.Compare(m.items[i], m.items[j]) < 0 }
func (m *CharModel) Swap(i, j int)               { m.items[i], m.items[j] = m.items[j], m.items[i] }

type AnnounceChanItem struct {
	ChanCmd     string
	DisplayName string
}

type AnnounceChanModel struct {
	walk.ListModelBase
	items []AnnounceChanItem
}

func (acm *AnnounceChanModel) ItemCount() int              { return len(acm.items) }
func (acm *AnnounceChanModel) Value(index int) interface{} { return acm.items[index].DisplayName }

var announceChannels = &AnnounceChanModel{
	items: []AnnounceChanItem{
		AnnounceChanItem{"/gu", "Guild"},
		AnnounceChanItem{"/say", "Say"},
		AnnounceChanItem{"/rsay", "Raid"},
		AnnounceChanItem{"/auc", "Auction"},
		AnnounceChanItem{"/shout", "Shout"},
		AnnounceChanItem{"/ooc", "Out of Character"},
		AnnounceChanItem{"/g", "Group"},
	},
}

var (
	uiFilePattern  = regexp.MustCompile("^UI_([A-Za-z]+_[A-Za-z]+).ini$")
	channelPattern = regexp.MustCompile("^[a-zA-Z0-9]+:[^ ]+$")
)

func validateEqDir(directory string) bool {
	if _, err := os.Stat(directory + "\\eqclient.ini"); err != nil {
		return false
	}
	if _, err := os.Stat(directory + "\\eqgame.exe"); err != nil {
		return false
	}
	return true
}

func validChanText(text string) bool {
	return channelPattern.MatchString(text)
}

func validToken(text string) bool {
	return len(text) > 1
}

func validCred(filename string) bool {
	fd, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer fd.Close()
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return false
	}
	var jsonValue interface{}
	err = json.Unmarshal(data, &jsonValue)
	if err != nil {
		log.Printf("Invalid credentials file: %v", err)
		return false
	}
	return true
}

func validLua(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func enumerateCharacters(directory string) *CharModel {
	cm := NewCharModel()
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return cm
	}
	for _, fileinfo := range files {
		if parts := uiFilePattern.FindStringSubmatch(fileinfo.Name()); parts != nil {
			cm.items = append(cm.items, parts[1])
		}
	}
	sort.Sort(cm)
	return cm
}

type mainWindowModel struct {
	mainWindow *walk.MainWindow

	dirEdit      *walk.LineEdit
	dirBrowse    *walk.PushButton
	announceChan *walk.ComboBox
	useLinks     *walk.CheckBox
	charBox      *walk.ComboBox
	chanEdit     *walk.LineEdit

	tokenEdit  *walk.LineEdit
	credEdit   *walk.LineEdit
	credBrowse *walk.PushButton

	luaEdit   *walk.LineEdit
	luaBrowse *walk.PushButton

	prepareButton *walk.PushButton
	startButton   *walk.PushButton
	started       bool
}

// Figure out what portions of the GUI should be enabled, and enable them
func (mwm *mainWindowModel) shade() {
	if mwm.started {
		mwm.dirEdit.SetEnabled(false)
		mwm.dirBrowse.SetEnabled(false)
		mwm.charBox.SetEnabled(false)
		mwm.chanEdit.SetEnabled(false)
		mwm.tokenEdit.SetEnabled(false)
		mwm.credEdit.SetEnabled(false)
		mwm.credBrowse.SetEnabled(false)
		mwm.luaEdit.SetEnabled(false)
		mwm.luaBrowse.SetEnabled(false)
		mwm.prepareButton.SetEnabled(false)
		mwm.useLinks.SetEnabled(false)
		mwm.startButton.SetEnabled(true)
		mwm.announceChan.SetEnabled(false)
		return
	} else {
		mwm.dirEdit.SetEnabled(true)
		mwm.dirBrowse.SetEnabled(true)
		mwm.chanEdit.SetEnabled(true)
		mwm.tokenEdit.SetEnabled(true)
		mwm.credEdit.SetEnabled(true)
		mwm.credBrowse.SetEnabled(true)
		mwm.luaEdit.SetEnabled(true)
		mwm.luaBrowse.SetEnabled(true)
		mwm.useLinks.SetEnabled(true)
		mwm.announceChan.SetEnabled(true)
	}
	selectedDir := mwm.dirEdit.Text()
	validDir := validateEqDir(selectedDir)
	if validDir {
		mwm.charBox.SetEnabled(true)
	} else {
		mwm.charBox.SetEnabled(false)
		mwm.charBox.SetCurrentIndex(-1)
	}

	charSelected := mwm.charBox.CurrentIndex() != -1
	if charSelected {
		mwm.prepareButton.SetEnabled(true)
	} else {
		mwm.prepareButton.SetEnabled(false)
	}

	useLinks := mwm.useLinks.Checked()
	validChannel := validChanText(mwm.chanEdit.Text())
	validToken := validToken(mwm.tokenEdit.Text())
	validCred := validCred(mwm.credEdit.Text())
	validLua := validLua(mwm.luaEdit.Text())

	if !useLinks {
		mwm.charBox.SetEnabled(false)
		mwm.prepareButton.SetEnabled(false)
	}

	if validDir && (!useLinks || charSelected) && validChannel && validToken && validCred && validLua {
		mwm.startButton.SetEnabled(true)
	} else {
		mwm.startButton.SetEnabled(false)
	}
}

func RunMainWindow(config storage2.ControllerConfig, start func(context.Context, storage2.ControllerConfig)) {
	model := &mainWindowModel{}
	var doneFunc func()

	err := MainWindow{
		Title:    "BidBot controller",
		AssignTo: &model.mainWindow,
		MinSize:  Size{720, 480},
		Layout:   VBox{},
		Children: []Widget{
			GroupBox{
				Layout: Grid{Columns: 3},
				Title:  "EverQuest settings",
				Children: []Widget{
					Label{
						Text:          "EverQuest directory",
						TextAlignment: AlignFar,
					},
					LineEdit{
						AssignTo: &model.dirEdit,
						OnTextChanged: func() {
							dirChoice := model.dirEdit.Text()
							cm := enumerateCharacters(dirChoice)
							model.charBox.SetModel(cm)
							model.charBox.SetCurrentIndex(-1)
							if validateEqDir(dirChoice) {
								config.SetEverQuestDirectory(dirChoice)
							}
							curchar := config.SelectedCharacter()
							if len(curchar) != 0 {
								for index, char_server := range cm.items {
									if strings.EqualFold(char_server, curchar) {
										model.charBox.SetCurrentIndex(index)
										break
									}
								}
							}
							model.shade()
						},
					},
					PushButton{
						Text:     "Browse...",
						AssignTo: &model.dirBrowse,
						OnClicked: func() {
							dialog := &walk.FileDialog{
								Title:    "Select EverQuest directory",
								FilePath: config.EverQuestDirectory(),
							}
							choose, err := dialog.ShowBrowseFolder(model.mainWindow)
							if err != nil {
								log.Println("Failed to show file dialog: %v", err)
								return
							}
							if choose {
								model.dirEdit.SetText(dialog.FilePath)
							}
						},
					},
					Label{
						Text:          "Annoucement Channel",
						TextAlignment: AlignFar,
					},
					ComboBox{
						AssignTo:   &model.announceChan,
						Model:      announceChannels,
						ColumnSpan: 2,
						Editable:   false,
						OnCurrentIndexChanged: func() {
							config.SetAnnounceChannel(announceChannels.items[model.announceChan.CurrentIndex()].ChanCmd)
						},
					},
					Label{
						Text:          "Control channel:password",
						TextAlignment: AlignFar,
					},
					LineEdit{
						AssignTo:   &model.chanEdit,
						ColumnSpan: 2,
						OnTextChanged: func() {
							chanText := model.chanEdit.Text()
							if strings.Compare(chanText, config.ChatChannelAndPassword()) != 0 {
								config.SetChannelImage(nil)
							}
							if validChanText(chanText) {
								config.SetChatChannelAndPassword(chanText)
							}
							model.shade()
						},
					},
					VSplitter{
						Children: []Widget{
							Label{
								Text:          "Link items during auction",
								TextAlignment: AlignFar,
							},
							Label{
								Text:          "(READ DOCUMENTATIION BEFORE USING)",
								TextAlignment: AlignFar,
							},
						},
					},
					CheckBox{
						AssignTo:   &model.useLinks,
						ColumnSpan: 2,
						OnCheckedChanged: func() {
							config.SetUseLinks(model.useLinks.Checked())
							model.shade()
						},
					},
					Label{
						Text:          "Bot character",
						TextAlignment: AlignFar,
					},
					ComboBox{
						Model:    NewCharModel(),
						AssignTo: &model.charBox,
						Editable: false,
						Enabled:  false,
						OnCurrentIndexChanged: func() {
							if model.charBox.CurrentIndex() >= 0 {
								config.SetSelectedCharacter(model.charBox.Model().(*CharModel).items[model.charBox.CurrentIndex()])
							}
							model.shade()
						},
					},
					PushButton{
						Text:     "Write window layout",
						Enabled:  false,
						AssignTo: &model.prepareButton,
						OnClicked: func() {
							everquest.WriteUIFile(config)
						},
					},
				},
			},
			GroupBox{
				Layout: Grid{Columns: 3},
				Title:  "Discord settings",
				Children: []Widget{
					Label{
						Text:          "Discord Token",
						TextAlignment: AlignFar,
					},
					LineEdit{
						AssignTo:   &model.tokenEdit,
						ColumnSpan: 2,
						OnTextChanged: func() {
							tokenText := model.tokenEdit.Text()
							if validToken(tokenText) {
								config.SetDiscordToken(tokenText)
							}
							model.shade()
						},
					},
					Label{
						Text:          "Google Cloud TTS credentials",
						TextAlignment: AlignFar,
					},
					LineEdit{
						AssignTo: &model.credEdit,
						OnTextChanged: func() {
							config.SetCloudTTSCredPath(model.credEdit.Text())
							model.shade()
						},
					},
					PushButton{
						Text:     "Browse...",
						AssignTo: &model.credBrowse,
						OnClicked: func() {
							dialog := &walk.FileDialog{
								Title:    "Select Google Cloud TTS credentials",
								FilePath: model.credEdit.Text(),
								Filter:   "JSON files (*.json)",
							}
							choose, err := dialog.ShowOpen(model.mainWindow)
							if err != nil {
								log.Println("Failed to show file dialog: %v", err)
								return
							}
							if choose {
								model.credEdit.SetText(dialog.FilePath)
							}
						},
					},
				},
			},
			GroupBox{
				Layout: Grid{Columns: 3},
				Title:  "Rules",
				Children: []Widget{
					Label{
						Text:          "Rules script",
						TextAlignment: AlignFar,
					},
					LineEdit{
						AssignTo: &model.luaEdit,
						OnTextChanged: func() {
							config.SetRulesLua(model.luaEdit.Text())
							model.shade()
						},
					},
					PushButton{
						Text:     "Browse...",
						AssignTo: &model.luaBrowse,
						OnClicked: func() {
							dialog := &walk.FileDialog{
								Title:    "Select Rules file",
								FilePath: model.luaEdit.Text(),
								Filter:   "LUA files (*.lua)",
							}
							choose, err := dialog.ShowOpen(model.mainWindow)
							if err != nil {
								log.Println("Failed to show file dialog: %v", err)
								return
							}
							if choose {
								model.luaEdit.SetText(dialog.FilePath)
							}
						},
					},
				},
			},
			HSplitter{
				Children: []Widget{
					PushButton{
						Text:     "Start",
						Enabled:  false,
						AssignTo: &model.startButton,
						OnClicked: func() {
							if model.started {
								if doneFunc != nil {
									doneFunc()
									doneFunc = nil
								}
								model.started = false
								model.startButton.SetText("Start")
								model.shade()
							} else {
								model.started = true
								model.startButton.SetText("Stop")
								model.shade()
								var ctx context.Context
								ctx, doneFunc = context.WithCancel(context.Background())
								go func() {
									start(ctx, config)
									model.mainWindow.Synchronize(func() {
										if doneFunc != nil {
											doneFunc()
											doneFunc = nil
										}
										model.started = false
										model.startButton.SetText("Start")
										model.shade()
									})
								}()
							}
						},
					},
				},
			},
		},
	}.Create()
	if err != nil {
		panic(err)
	}
	lv, _ := NewLogView(model.mainWindow)
	lv.PostAppendText("")
	log.SetOutput(lv)

	model.dirEdit.SetText(config.EverQuestDirectory())
	model.chanEdit.SetText(config.ChatChannelAndPassword())
	model.tokenEdit.SetText(config.DiscordToken())
	model.credEdit.SetText(config.CloudTTSCredPath())
	model.luaEdit.SetText(config.RulesLua())
	model.useLinks.SetChecked(config.UseLinks())
	curAnnounceChan := config.AnnounceChannel()
	for idx, ac := range announceChannels.items {
		if ac.ChanCmd == curAnnounceChan {
			model.announceChan.SetCurrentIndex(idx)
			break
		}
	}
	model.shade()

	model.mainWindow.Run()
}
