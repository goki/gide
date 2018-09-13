// Copyright (c) 2018, The Gide Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gide

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/goki/gi"
	"github.com/goki/gi/oswin"
	"github.com/goki/ki"
	"github.com/goki/ki/kit"
)

// todo: really need a consolidated mimetype system too, that integrates with this

// currently use this as a list:
// https://github.com/alecthomas/chroma

// Lang defines properties associated with a given language or file type more
// generally (e.g., image files, data files, etc)
type Lang struct {
	Name         string   `desc:"name of this language / data / file type (must be unique)"`
	Desc         string   `desc:"<i>brief</i> description of it"`
	Exts         []string `desc:"associated file extensions -- if the filename itself is more diagnostic (e.g., Makefile), specify that -- if it doesn't start with a . then it will be treated as the start of the filename"`
	PostSaveCmds CmdNames `desc:"command(s) to run after a file of this type is saved"`
}

// Langs is a list of language types
type Langs []Lang

// LangName has an associated ValueView for selecting from the list of
// available language names, for use in preferences etc.
type LangName string

// LangNames is a list of language names
type LangNames []LangName

// ExtToLangMap is a compiled map of file extensions (always lowercased) and
// their associated language(s) -- there can be some ambiguity (e.g.,
// .h files), so multiple languages are allowed
var ExtToLangMap = map[string]Langs{}

var KiT_Langs = kit.Types.AddType(&Langs{}, LangsProps)

// AvailLangs is the current list of available languages defined -- can be
// loaded / saved / edited with preferences.  This is set to StdLangs at
// startup.
var AvailLangs Langs

func init() {
	AvailLangs.CopyFrom(StdLangs)
}

// LangByName returns a language and index by name -- returns false and emits a
// message to stdout if not found
func (lt *Langs) LangByName(name LangName) (*Lang, int, bool) {
	for i := range *lt {
		lr := &((*lt)[i])
		if lr.Name == string(name) {
			return lr, i, true
		}
	}
	fmt.Printf("gide.LangByName: language named: %v not found\n", name)
	return nil, -1, false
}

// PrefsLangsFileName is the name of the preferences file in App prefs
// directory for saving / loading the default AvailLangs languages list
var PrefsLangsFileName = "lang_prefs.json"

// OpenJSON opens languages from a JSON-formatted file.
func (lt *Langs) OpenJSON(filename gi.FileName) error {
	*lt = make(Langs, 0, 10) // reset
	b, err := ioutil.ReadFile(string(filename))
	if err != nil {
		// gi.PromptDialog(nil, gi.DlgOpts{Title: "File Not Found", Prompt: err.Error()}, true, false, nil, nil)
		// log.Println(err)
		return err
	}
	return json.Unmarshal(b, lt)
}

// SaveJSON saves languages to a JSON-formatted file.
func (lt *Langs) SaveJSON(filename gi.FileName) error {
	b, err := json.MarshalIndent(lt, "", "  ")
	if err != nil {
		log.Println(err) // unlikely
		return err
	}
	err = ioutil.WriteFile(string(filename), b, 0644)
	if err != nil {
		gi.PromptDialog(nil, gi.DlgOpts{Title: "Could not Save to File", Prompt: err.Error()}, true, false, nil, nil)
		log.Println(err)
	}
	return err
}

// OpenPrefs opens Langs from App standard prefs directory, using PrefsLangsFileName
func (lt *Langs) OpenPrefs() error {
	pdir := oswin.TheApp.AppPrefsDir()
	pnm := filepath.Join(pdir, PrefsLangsFileName)
	AvailLangsChanged = false
	return lt.OpenJSON(gi.FileName(pnm))
}

// SavePrefs saves Langs to App standard prefs directory, using PrefsLangsFileName
func (lt *Langs) SavePrefs() error {
	pdir := oswin.TheApp.AppPrefsDir()
	pnm := filepath.Join(pdir, PrefsLangsFileName)
	AvailLangsChanged = false
	return lt.SaveJSON(gi.FileName(pnm))
}

// CopyFrom copies languages from given other map
func (lt *Langs) CopyFrom(cp Langs) {
	*lt = make(Langs, 0, len(cp)) // reset
	b, err := json.Marshal(cp)
	if err != nil {
		fmt.Printf("json err: %v\n", err.Error())
	}
	json.Unmarshal(b, lt)
}

// RevertToStd reverts this map to using the StdLangs that are compiled into
// the program and have all the lastest standards.
func (lt *Langs) RevertToStd() {
	lt.CopyFrom(StdLangs)
	AvailLangsChanged = true
}

// ViewStd shows the standard types that are compiled into the program and have
// all the lastest standards.  Useful for comparing against custom lists.
func (lt *Langs) ViewStd() {
	LangsView(&StdLangs)
}

// AvailLangsChanged is used to update giv.LangsView toolbars via
// following menu, toolbar props update methods -- not accurate if editing any
// other map but works for now..
var AvailLangsChanged = false

// LangsProps define the ToolBar and MenuBar for TableView of Langs, e.g., giv.LangsView
var LangsProps = ki.Props{
	"MainMenu": ki.PropSlice{
		{"AppMenu", ki.BlankProp{}},
		{"File", ki.PropSlice{
			{"OpenPrefs", ki.Props{}},
			{"SavePrefs", ki.Props{
				"shortcut": "Command+S",
				"updtfunc": func(lti interface{}, act *gi.Action) {
					act.SetActiveState(AvailLangsChanged)
				},
			}},
			{"sep-file", ki.BlankProp{}},
			{"OpenJSON", ki.Props{
				"label":    "Open from file",
				"desc":     "You can save and open languages to / from files to share, experiment, transfer, etc",
				"shortcut": "Command+O",
				"Args": ki.PropSlice{
					{"File Name", ki.Props{
						"ext": ".json",
					}},
				},
			}},
			{"SaveJSON", ki.Props{
				"label": "Save to file",
				"desc":  "You can save and open languages to / from files to share, experiment, transfer, etc",
				"Args": ki.PropSlice{
					{"File Name", ki.Props{
						"ext": ".json",
					}},
				},
			}},
			{"RevertToStd", ki.Props{
				"desc":    "This reverts the languages to using the StdLangs that are compiled into the program and have all the lastest standards. <b>Your current edits will be lost if you proceed!</b>  Continue?",
				"confirm": true,
			}},
		}},
		{"Edit", "Copy Cut Paste Dupe"},
		{"Window", "Windows"},
	},
	"ToolBar": ki.PropSlice{
		{"SavePrefs", ki.Props{
			"desc": "saves Langs to App standard prefs directory, in file lang_prefs.json, which will be loaded automatically at startup if prefs SaveLangs is checked (should be if you're using custom languages)",
			"icon": "file-save",
			"updtfunc": func(lti interface{}, act *gi.Action) {
				act.SetActiveStateUpdt(AvailLangsChanged)
			},
		}},
		{"sep-file", ki.BlankProp{}},
		{"OpenJSON", ki.Props{
			"label": "Open from file",
			"icon":  "file-open",
			"desc":  "You can save and open languages to / from files to share, experiment, transfer, etc",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".json",
				}},
			},
		}},
		{"SaveJSON", ki.Props{
			"label": "Save to file",
			"icon":  "file-save",
			"desc":  "You can save and open languages to / from files to share, experiment, transfer, etc",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".json",
				}},
			},
		}},
		{"sep-std", ki.BlankProp{}},
		{"ViewStd", ki.Props{
			"desc":    "Shows the standard types that are compiled into the program and have all the latest changes.  Useful for comparing against custom types.",
			"confirm": true,
		}},
		{"RevertToStd", ki.Props{
			"icon":    "update",
			"desc":    "This reverts the languages to using the StdLangs that are compiled into the program and have all the lastest standards.  <b>Your current edits will be lost if you proceed!</b>  Continue?",
			"confirm": true,
		}},
	},
}

// StdLangs is the original compiled-in set of standard languages.
var StdLangs = Langs{
	{"Go", "Go code", []string{".go"}, CmdNames{"Go Imports File"}},
}
