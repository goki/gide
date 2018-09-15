// Copyright (c) 2018, The Gide Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gide

import (
	"fmt"
	"reflect"

	"github.com/goki/gi"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/oswin"
	"github.com/goki/gi/units"
	"github.com/goki/ki"
	"github.com/goki/ki/kit"
)

//////////////////////////////////////////////////////////////////////////////////////
//  PrefsView

// PrefsView opens a view of user preferences, returns structview and window
func PrefsView(pf *Preferences) (*giv.StructView, *gi.Window) {
	winm := "gide-prefs"
	if w, ok := gi.MainWindows.FindName(winm); ok {
		w.OSWin.Raise()
		return nil, nil
	}

	width := 800
	height := 800
	win := gi.NewWindow2D(winm, "Gide Preferences", width, height, true)

	vp := win.WinViewport2D()
	updt := vp.UpdateStart()

	mfr := win.SetMainFrame()
	mfr.Lay = gi.LayoutVert

	sv := mfr.AddNewChild(giv.KiT_StructView, "sv").(*giv.StructView)
	sv.Viewport = vp
	sv.SetStruct(pf, nil)
	sv.SetStretchMaxWidth()
	sv.SetStretchMaxHeight()

	mmen := win.MainMenu
	giv.MainMenuView(pf, win, mmen)

	inClosePrompt := false
	win.OSWin.SetCloseReqFunc(func(w oswin.Window) {
		if pf.Changed {
			if !inClosePrompt {
				gi.ChoiceDialog(vp, gi.DlgOpts{Title: "Save Prefs Before Closing?",
					Prompt: "Do you want to save any changes to preferences before closing?"},
					[]string{"Save and Close", "Discard and Close", "Cancel"},
					win.This, func(recv, send ki.Ki, sig int64, data interface{}) {
						switch sig {
						case 0:
							pf.Save()
							fmt.Println("Preferences Saved to prefs.json")
							w.Close()
						case 1:
							pf.Open() // if we don't do this, then it actually remains in edited state
							w.Close()
						case 2:
							inClosePrompt = false
							// default is to do nothing, i.e., cancel
						}
					})
			}
		} else {
			w.Close()
		}
	})

	win.MainMenuUpdated()

	vp.UpdateEndNoSig(updt)
	win.GoStartEventLoop()
	return sv, win
}

//////////////////////////////////////////////////////////////////////////////////////
//  ProjPrefsView

// ProjPrefsView opens a view of project preferences (settings), returns structview and window
func ProjPrefsView(pf *ProjPrefs) (*giv.StructView, *gi.Window) {
	winm := "gide-proj-prefs"

	width := 800
	height := 800
	win := gi.NewWindow2D(winm, "Gide Project Preferences", width, height, true)

	vp := win.WinViewport2D()
	updt := vp.UpdateStart()

	mfr := win.SetMainFrame()
	mfr.Lay = gi.LayoutVert

	title := mfr.AddNewChild(gi.KiT_Label, "title").(*gi.Label)
	title.SetText("Project preferences are saved in the project .gide file, along with other current state (open directories, splitter settings, etc) -- do Save Project to save.")
	title.SetProp("word-wrap", true)

	sv := mfr.AddNewChild(giv.KiT_StructView, "sv").(*giv.StructView)
	sv.Viewport = vp
	sv.SetStruct(pf, nil)
	sv.SetStretchMaxWidth()
	sv.SetStretchMaxHeight()

	mmen := win.MainMenu
	giv.MainMenuView(pf, win, mmen)

	win.MainMenuUpdated()

	vp.UpdateEndNoSig(updt)
	win.GoStartEventLoop()
	return sv, win
}

//////////////////////////////////////////////////////////////////////////////////////
//  KeyMapsView

// KeyMapsView opens a view of a key maps table
func KeyMapsView(km *KeyMaps) {
	winm := "gide-key-maps"
	width := 800
	height := 800
	win := gi.NewWindow2D(winm, "Gide Key Maps", width, height, true)

	vp := win.WinViewport2D()
	updt := vp.UpdateStart()

	mfr := win.SetMainFrame()
	mfr.Lay = gi.LayoutVert

	title := mfr.AddNewChild(gi.KiT_Label, "title").(*gi.Label)
	title.SetText("Available Key Maps: Duplicate an existing map (using Ctxt Menu) as starting point for creating a custom map")
	title.SetProp("word-wrap", true)

	tv := mfr.AddNewChild(giv.KiT_TableView, "tv").(*giv.TableView)
	tv.Viewport = vp
	tv.SetSlice(km, nil)
	tv.SetStretchMaxWidth()
	tv.SetStretchMaxHeight()

	AvailKeyMapsChanged = false
	tv.ViewSig.Connect(mfr.This, func(recv, send ki.Ki, sig int64, data interface{}) {
		AvailKeyMapsChanged = true
	})

	mmen := win.MainMenu
	giv.MainMenuView(km, win, mmen)

	win.OSWin.SetCloseReqFunc(func(w oswin.Window) {
		if AvailKeyMapsChanged { // only for main avail map..
			gi.ChoiceDialog(vp, gi.DlgOpts{Title: "Save KeyMaps Before Closing?",
				Prompt: "Do you want to save any changes to std preferences to std keymaps file before closing, or Cancel the close and do a Save to a different file?"},
				[]string{"Save and Close", "Discard and Close", "Cancel"},
				win.This, func(recv, send ki.Ki, sig int64, data interface{}) {
					switch sig {
					case 0:
						km.SavePrefs()
						fmt.Printf("Preferences Saved to %v\n", PrefsKeyMapsFileName)
						w.Close()
					case 1:
						if km == &AvailKeyMaps {
							km.OpenPrefs() // revert
						}
						w.Close()
					case 2:
						// default is to do nothing, i.e., cancel
					}
				})
		} else {
			w.Close()
		}
	})

	win.MainMenuUpdated()

	vp.UpdateEndNoSig(updt)
	win.GoStartEventLoop()
}

////////////////////////////////////////////////////////////////////////////////////////
//  KeyMapValueView

// ValueView registers KeyMapValueView as the viewer of KeyMapName
func (kn KeyMapName) ValueView() giv.ValueView {
	vv := KeyMapValueView{}
	vv.Init(&vv)
	return &vv
}

// KeyMapValueView presents an action for displaying an KeyMapName and selecting
// from KeyMapChooserDialog
type KeyMapValueView struct {
	giv.ValueViewBase
}

var KiT_KeyMapValueView = kit.Types.AddType(&KeyMapValueView{}, nil)

func (vv *KeyMapValueView) WidgetType() reflect.Type {
	vv.WidgetTyp = gi.KiT_Action
	return vv.WidgetTyp
}

func (vv *KeyMapValueView) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	ac := vv.Widget.(*gi.Action)
	txt := kit.ToString(vv.Value.Interface())
	ac.SetFullReRender()
	ac.SetText(txt)
}

func (vv *KeyMapValueView) ConfigWidget(widg gi.Node2D) {
	vv.Widget = widg
	ac := vv.Widget.(*gi.Action)
	ac.SetProp("border-radius", units.NewValue(4, units.Px))
	ac.ActionSig.ConnectOnly(vv.This, func(recv, send ki.Ki, sig int64, data interface{}) {
		vvv, _ := recv.Embed(KiT_KeyMapValueView).(*KeyMapValueView)
		ac := vvv.Widget.(*gi.Action)
		vvv.Activate(ac.Viewport, nil, nil)
	})
	vv.UpdateWidget()
}

func (vv *KeyMapValueView) HasAction() bool {
	return true
}

func (vv *KeyMapValueView) Activate(vp *gi.Viewport2D, dlgRecv ki.Ki, dlgFunc ki.RecvFunc) {
	if vv.IsInactive() {
		return
	}
	cur := kit.ToString(vv.Value.Interface())
	_, curRow, _ := AvailKeyMaps.MapByName(KeyMapName(cur))
	desc, _ := vv.Tag("desc")
	giv.TableViewSelectDialog(vp, &AvailKeyMaps, giv.DlgOpts{Title: "Select a KeyMap", Prompt: desc}, curRow, nil,
		vv.This, func(recv, send ki.Ki, sig int64, data interface{}) {
			if sig == int64(gi.DialogAccepted) {
				ddlg, _ := send.(*gi.Dialog)
				si := giv.TableViewSelectDialogValue(ddlg)
				if si >= 0 {
					km := AvailKeyMaps[si]
					vv.SetValue(km.Name)
					vv.UpdateWidget()
				}
			}
			if dlgRecv != nil && dlgFunc != nil {
				dlgFunc(dlgRecv, send, sig, data)
			}
		})
}

//////////////////////////////////////////////////////////////////////////////////////
//  LangsView

// LangsView opens a view of a languages table
func LangsView(pt *Langs) {
	winm := "gide-langs"
	width := 800
	height := 800
	win := gi.NewWindow2D(winm, "Gide Languages", width, height, true)

	vp := win.WinViewport2D()
	updt := vp.UpdateStart()

	mfr := win.SetMainFrame()
	mfr.Lay = gi.LayoutVert

	title := mfr.AddNewChild(gi.KiT_Label, "title").(*gi.Label)
	title.SetText("Available Languages: Duplicate an existing (using Ctxt Menu) as starting point for creating a custom entry")
	title.SetProp("word-wrap", true)

	tv := mfr.AddNewChild(giv.KiT_TableView, "tv").(*giv.TableView)
	tv.Viewport = vp
	tv.SetSlice(pt, nil)
	tv.SetStretchMaxWidth()
	tv.SetStretchMaxHeight()

	AvailLangsChanged = false
	tv.ViewSig.Connect(mfr.This, func(recv, send ki.Ki, sig int64, data interface{}) {
		AvailLangsChanged = true
	})

	mmen := win.MainMenu
	giv.MainMenuView(pt, win, mmen)

	win.OSWin.SetCloseReqFunc(func(w oswin.Window) {
		if AvailLangsChanged { // only for main avail map..
			gi.ChoiceDialog(vp, gi.DlgOpts{Title: "Save Langs Before Closing?",
				Prompt: "Do you want to save any changes to std preferences of std languages file before closing, or Cancel the close and do a Save to a different file?"},
				[]string{"Save and Close", "Discard and Close", "Cancel"},
				win.This, func(recv, send ki.Ki, sig int64, data interface{}) {
					switch sig {
					case 0:
						pt.SavePrefs()
						fmt.Printf("Preferences Saved to %v\n", PrefsLangsFileName)
						w.Close()
					case 1:
						if pt == &AvailLangs {
							pt.OpenPrefs() // revert
						}
						w.Close()
					case 2:
						// default is to do nothing, i.e., cancel
					}
				})
		} else {
			w.Close()
		}
	})

	win.MainMenuUpdated()

	vp.UpdateEndNoSig(updt)
	win.GoStartEventLoop()
}

////////////////////////////////////////////////////////////////////////////////////////
//  LangValueView

// ValueView registers LangValueView as the viewer of LangName
func (kn LangName) ValueView() giv.ValueView {
	vv := LangValueView{}
	vv.Init(&vv)
	return &vv
}

// LangValueView presents an action for displaying an LangName and selecting
// from LangChooserDialog
type LangValueView struct {
	giv.ValueViewBase
}

var KiT_LangValueView = kit.Types.AddType(&LangValueView{}, nil)

func (vv *LangValueView) WidgetType() reflect.Type {
	vv.WidgetTyp = gi.KiT_Action
	return vv.WidgetTyp
}

func (vv *LangValueView) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	ac := vv.Widget.(*gi.Action)
	txt := kit.ToString(vv.Value.Interface())
	ac.SetFullReRender()
	ac.SetText(txt)
}

func (vv *LangValueView) ConfigWidget(widg gi.Node2D) {
	vv.Widget = widg
	ac := vv.Widget.(*gi.Action)
	ac.SetProp("border-radius", units.NewValue(4, units.Px))
	ac.ActionSig.ConnectOnly(vv.This, func(recv, send ki.Ki, sig int64, data interface{}) {
		vvv, _ := recv.Embed(KiT_LangValueView).(*LangValueView)
		ac := vvv.Widget.(*gi.Action)
		vvv.Activate(ac.Viewport, nil, nil)
	})
	vv.UpdateWidget()
}

func (vv *LangValueView) HasAction() bool {
	return true
}

func (vv *LangValueView) Activate(vp *gi.Viewport2D, dlgRecv ki.Ki, dlgFunc ki.RecvFunc) {
	if vv.IsInactive() {
		return
	}
	cur := kit.ToString(vv.Value.Interface())
	_, curRow, _ := AvailLangs.LangByName(LangName(cur))
	desc, _ := vv.Tag("desc")
	giv.TableViewSelectDialog(vp, &AvailLangs, giv.DlgOpts{Title: "Select a Lang", Prompt: desc}, curRow, nil,
		vv.This, func(recv, send ki.Ki, sig int64, data interface{}) {
			if sig == int64(gi.DialogAccepted) {
				ddlg, _ := send.(*gi.Dialog)
				si := giv.TableViewSelectDialogValue(ddlg)
				if si >= 0 {
					pt := AvailLangs[si]
					vv.SetValue(pt.Name)
					vv.UpdateWidget()
				}
			}
			if dlgRecv != nil && dlgFunc != nil {
				dlgFunc(dlgRecv, send, sig, data)
			}
		})
}

//////////////////////////////////////////////////////////////////////////////////////
//  CmdsView

// CmdsView opens a view of a commands table
func CmdsView(pt *Commands) {
	winm := "gide-proj-types"
	width := 800
	height := 800
	win := gi.NewWindow2D(winm, "Gide Commands", width, height, true)

	vp := win.WinViewport2D()
	updt := vp.UpdateStart()

	mfr := win.SetMainFrame()
	mfr.Lay = gi.LayoutVert

	title := mfr.AddNewChild(gi.KiT_Label, "title").(*gi.Label)
	title.SetText("Available Commands: Can duplicate an existing (using Ctxt Menu) as starting point for new one")
	title.SetProp("word-wrap", true)

	tv := mfr.AddNewChild(giv.KiT_TableView, "tv").(*giv.TableView)
	tv.Viewport = vp
	tv.SetSlice(pt, nil)
	tv.SetStretchMaxWidth()
	tv.SetStretchMaxHeight()

	AvailCmdsChanged = false
	tv.ViewSig.Connect(mfr.This, func(recv, send ki.Ki, sig int64, data interface{}) {
		AvailCmdsChanged = true
	})

	mmen := win.MainMenu
	giv.MainMenuView(pt, win, mmen)

	win.OSWin.SetCloseReqFunc(func(w oswin.Window) {
		if AvailCmdsChanged { // only for main avail map..
			gi.ChoiceDialog(vp, gi.DlgOpts{Title: "Save Commands Before Closing?",
				Prompt: "Do you want to save any changes to std preferences to std commands file before closing, or Cancel the close and do a Save to a different file?"},
				[]string{"Save and Close", "Discard and Close", "Cancel"},
				win.This, func(recv, send ki.Ki, sig int64, data interface{}) {
					switch sig {
					case 0:
						pt.SavePrefs()
						fmt.Printf("Preferences Saved to %v\n", PrefsCommandsFileName)
						w.Close()
					case 1:
						if pt == &AvailCmds {
							pt.OpenPrefs() // revert
						}
						w.Close()
					case 2:
						// default is to do nothing, i.e., cancel
					}
				})
		} else {
			w.Close()
		}
	})

	win.MainMenuUpdated()

	vp.UpdateEndNoSig(updt)
	win.GoStartEventLoop()
}

////////////////////////////////////////////////////////////////////////////////////////
//  CmdValueView

// ValueView registers CmdValueView as the viewer of CmdName
func (kn CmdName) ValueView() giv.ValueView {
	vv := CmdValueView{}
	vv.Init(&vv)
	return &vv
}

// CmdValueView presents an action for displaying an CmdName and selecting
// from CommandChooserDialog
type CmdValueView struct {
	giv.ValueViewBase
}

var KiT_CmdValueView = kit.Types.AddType(&CmdValueView{}, nil)

func (vv *CmdValueView) WidgetType() reflect.Type {
	vv.WidgetTyp = gi.KiT_Action
	return vv.WidgetTyp
}

func (vv *CmdValueView) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	ac := vv.Widget.(*gi.Action)
	txt := kit.ToString(vv.Value.Interface())
	ac.SetFullReRender()
	ac.SetText(txt)
}

func (vv *CmdValueView) ConfigWidget(widg gi.Node2D) {
	vv.Widget = widg
	ac := vv.Widget.(*gi.Action)
	ac.SetProp("border-radius", units.NewValue(4, units.Px))
	ac.ActionSig.ConnectOnly(vv.This, func(recv, send ki.Ki, sig int64, data interface{}) {
		vvv, _ := recv.Embed(KiT_CmdValueView).(*CmdValueView)
		ac := vvv.Widget.(*gi.Action)
		vvv.Activate(ac.Viewport, nil, nil)
	})
	vv.UpdateWidget()
}

func (vv *CmdValueView) HasAction() bool {
	return true
}

func (vv *CmdValueView) Activate(vp *gi.Viewport2D, dlgRecv ki.Ki, dlgFunc ki.RecvFunc) {
	if vv.IsInactive() {
		return
	}
	cur := kit.ToString(vv.Value.Interface())
	curRow := -1
	if cur != "" {
		_, curRow, _ = AvailCmds.CmdByName(CmdName(cur))
	}
	desc, _ := vv.Tag("desc")
	giv.TableViewSelectDialog(vp, &AvailCmds, giv.DlgOpts{Title: "Select a Command", Prompt: desc}, curRow, nil,
		vv.This, func(recv, send ki.Ki, sig int64, data interface{}) {
			if sig == int64(gi.DialogAccepted) {
				ddlg, _ := send.(*gi.Dialog)
				si := giv.TableViewSelectDialogValue(ddlg)
				if si >= 0 {
					pt := AvailCmds[si]
					vv.SetValue(pt.Name)
					vv.UpdateWidget()
				}
			}
			if dlgRecv != nil && dlgFunc != nil {
				dlgFunc(dlgRecv, send, sig, data)
			}
		})
}