// Copyright (c) 2018, The Gide Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gide

import (
	"log"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/goki/gi/gi"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/giv/textbuf"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
	"github.com/goki/pi/filecat"
)

// FileNode is Gide version of FileNode for FileTree view
type FileNode struct {
	giv.FileNode
}

var KiT_FileNode = kit.Types.AddType(&FileNode{}, nil)

func (fn *FileNode) CopyFieldsFrom(frm interface{}) {
	fr := frm.(*FileNode)
	fn.FileNode.CopyFieldsFrom(&fr.FileNode)
	// no copy here
}

// ParentGide returns the Gide parent of given node
func ParentGide(kn ki.Ki) (Gide, bool) {
	if ki.IsRoot(kn) {
		return nil, false
	}
	var ge Gide
	kn.FuncUpParent(0, kn, func(k ki.Ki, level int, d interface{}) bool {
		if kit.EmbedImplements(ki.Type(k), GideType) {
			ge = k.(Gide)
			return false
		}
		return true
	})
	return ge, ge != nil
}

// EditFile pulls up this file in Gide
func (fn *FileNode) EditFile() {
	if fn.IsDir() {
		log.Printf("FileNode Edit -- cannot view (edit) directories!\n")
		return
	}
	ge, ok := ParentGide(fn.This())
	if ok {
		ge.NextViewFileNode(fn.This().Embed(giv.KiT_FileNode).(*giv.FileNode))
	}
}

// SetRunExec sets executable as the RunExec executable that will be run with Run / Debug buttons
func (fn *FileNode) SetRunExec() {
	if !fn.IsExec() {
		log.Printf("FileNode SetRunExec -- only works for executable files!\n")
		return
	}
	ge, ok := ParentGide(fn.This())
	if ok {
		ge.ProjPrefs().RunExec = fn.FPath
		ge.ProjPrefs().BuildDir = gi.FileName(filepath.Dir(string(fn.FPath)))
	}
}

// ExecCmdFile pops up a menu to select a command appropriate for the given node,
// and shows output in MainTab with name of command
func (fn *FileNode) ExecCmdFile() {
	ge, ok := ParentGide(fn.This())
	if ok {
		ge.ExecCmdFileNode(fn.This().Embed(giv.KiT_FileNode).(*giv.FileNode))
	}
}

// ExecCmdNameFile executes given command name on node
func (fn *FileNode) ExecCmdNameFile(cmdNm string) {
	ge, ok := ParentGide(fn.This())
	if ok {
		ge.ExecCmdNameFileNode(fn.This().Embed(giv.KiT_FileNode).(*giv.FileNode), CmdName(cmdNm), true, true)
	}
}

/////////////////////////////////////////////////////////////////////
//   OpenNodes

// OpenNodes is a list of file nodes that have been opened for editing -- it
// is maintained in recency order -- most recent on top -- call Add every time
// a node is opened / visited for editing
type OpenNodes []*giv.FileNode

// Add adds given node to list of open nodes -- if already on the list it is
// moved to the top -- returns true if actually added.
// Connects to fn.TextBuf signal and auto-closes when buffer closes.
func (on *OpenNodes) Add(fn *giv.FileNode) bool {
	added := on.AddImpl(fn)
	if !added {
		return added
	}
	if fn.Buf != nil {
		fn.Buf.TextBufSig.Connect(fn.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
			if sig == int64(giv.TextBufClosed) {
				fno, _ := recv.Embed(giv.KiT_FileNode).(*giv.FileNode)
				on.Delete(fno)
			}
		})
	}
	return added
}

// AddImpl adds given node to list of open nodes -- if already on the list it is
// moved to the top -- returns true if actually added.
func (on *OpenNodes) AddImpl(fn *giv.FileNode) bool {
	sz := len(*on)

	for i, f := range *on {
		if f == fn {
			if i == 0 {
				return false
			}
			copy((*on)[1:i+1], (*on)[0:i])
			(*on)[0] = fn
			return false
		}
	}

	*on = append(*on, nil)
	if sz > 0 {
		copy((*on)[1:], (*on)[0:sz])
	}
	(*on)[0] = fn
	return true
}

// Delete deletes given node in list of open nodes, returning true if found and deleted
func (on *OpenNodes) Delete(fn *giv.FileNode) bool {
	for i, f := range *on {
		if f == fn {
			on.DeleteIdx(i)
			return true
		}
	}
	return false
}

// DeleteIdx deletes at given index
func (on *OpenNodes) DeleteIdx(idx int) {
	*on = append((*on)[:idx], (*on)[idx+1:]...)
}

// DeleteDeleted deletes deleted nodes on list
func (on *OpenNodes) DeleteDeleted() {
	sz := len(*on)
	for i := sz - 1; i >= 0; i-- {
		fn := (*on)[i]
		if fn.This() == nil || fn.FRoot == nil || fn.IsDeleted() {
			on.DeleteIdx(i)
		}
	}
}

// Strings returns a string list of nodes, with paths relative to proj root
func (on *OpenNodes) Strings() []string {
	on.DeleteDeleted()
	sl := make([]string, len(*on))
	for i, fn := range *on {
		rp := fn.FRoot.RelPath(fn.FPath)
		rp = strings.TrimSuffix(rp, fn.Nm)
		if rp != "" {
			sl[i] = fn.Nm + " - " + rp
		} else {
			sl[i] = fn.Nm
		}
		if fn.IsChanged() {
			sl[i] += " *"
		}
	}
	return sl
}

// ByStringName returns the open node with given strings name
func (on *OpenNodes) ByStringName(name string) *giv.FileNode {
	sl := on.Strings()
	for i, ns := range sl {
		if ns == name {
			return (*on)[i]
		}
	}
	return nil
}

// NChanged returns number of changed open files
func (on *OpenNodes) NChanged() int {
	cnt := 0
	for _, fn := range *on {
		if fn.IsChanged() {
			cnt++
		}
	}
	return cnt
}

//////////////////////////////////////////////////////////////////////////
//  Search

// FileSearchResults is used to report search results
type FileSearchResults struct {
	Node    *giv.FileNode
	Count   int
	Matches []textbuf.Match
}

// FileTreeSearch returns list of all nodes starting at given node of given
// language(s) that contain the given string (non regexp version), sorted in
// descending order by number of occurrences -- ignoreCase transforms
// everything into lowercase
func FileTreeSearch(start *giv.FileNode, find string, ignoreCase, regExp bool, loc FindLoc, activeDir string, langs []filecat.Supported) []FileSearchResults {
	fb := []byte(find)
	fsz := len(find)
	if fsz == 0 {
		return nil
	}
	var re *regexp.Regexp
	var err error
	if regExp {
		re, err = regexp.Compile(find)
		if err != nil {
			log.Println(err)
			return nil
		}
	}
	mls := make([]FileSearchResults, 0)
	start.FuncDownMeFirst(0, start, func(k ki.Ki, level int, d interface{}) bool {
		sfn := k.Embed(giv.KiT_FileNode).(*giv.FileNode)
		if sfn.IsDir() && !sfn.IsOpen() {
			// fmt.Printf("dir: %v closed\n", sfn.FPath)
			return ki.Break // don't go down into closed directories!
		}
		if sfn.IsDir() || sfn.IsExec() || sfn.Info.Kind == "octet-stream" || sfn.IsAutoSave() {
			// fmt.Printf("dir: %v opened\n", sfn.Nm)
			return ki.Continue
		}
		if strings.HasSuffix(sfn.Nm, ".gide") { // exclude self
			return ki.Continue
		}
		if !filecat.IsMatchList(langs, sfn.Info.Sup) {
			return ki.Continue
		}
		if loc == FindLocDir {
			cdir, _ := filepath.Split(string(sfn.FPath))
			if activeDir != cdir {
				return ki.Continue
			}
		} else if loc == FindLocNotTop {
			if level == 1 {
				return ki.Continue
			}
		}
		var cnt int
		var matches []textbuf.Match
		if sfn.IsOpen() && sfn.Buf != nil {
			if regExp {
				cnt, matches = sfn.Buf.SearchRegexp(re)
			} else {
				cnt, matches = sfn.Buf.Search(fb, ignoreCase, false)
			}
		} else {
			if regExp {
				cnt, matches = textbuf.SearchFileRegexp(string(sfn.FPath), re)
			} else {
				cnt, matches = textbuf.SearchFile(string(sfn.FPath), fb, ignoreCase)
			}
		}
		if cnt > 0 {
			mls = append(mls, FileSearchResults{sfn, cnt, matches})
		}
		return ki.Continue
	})
	sort.Slice(mls, func(i, j int) bool {
		return mls[i].Count > mls[j].Count
	})
	return mls
}

/////////////////////////////////////////////////////////////////////////
// FileTreeView is the Gide version of the FileTreeView

// FileTreeView is a TreeView that knows how to operate on FileNode nodes
type FileTreeView struct {
	giv.FileTreeView
}

var FileTreeViewProps map[string]interface{}
var FileNodeProps map[string]interface{}

var KiT_FileTreeView = kit.Types.AddType(&FileTreeView{}, nil)

func init() {
	FileNodeProps = make(ki.Props, len(giv.FileNodeProps))
	ki.CopyProps(&FileNodeProps, giv.FileNodeProps, true)
	kit.Types.SetProps(KiT_FileNode, FileNodeProps)

	FileTreeViewProps = make(ki.Props, len(giv.FileTreeViewProps))
	ki.CopyProps(&FileTreeViewProps, giv.FileTreeViewProps, ki.DeepCopy)
	cm := FileTreeViewProps["CtxtMenuActive"].(ki.PropSlice)
	cm = append(ki.PropSlice{
		{"ExecCmdFiles", ki.Props{
			"label":        "Exec Cmd",
			"submenu-func": giv.SubMenuFunc(FileTreeViewExecCmds),
			"Args": ki.PropSlice{
				{"Cmd Name", ki.Props{}},
			},
		}},
		{"EditFiles", ki.Props{
			"label":    "Edit",
			"updtfunc": FileTreeInactiveDirFunc,
		}},
		{"SetRunExec", ki.Props{
			"label":    "Set Run Exec",
			"updtfunc": FileTreeActiveExecFunc,
		}},
		{"sep-view", ki.BlankProp{}},
	}, cm...)
	FileTreeViewProps["CtxtMenuActive"] = cm
	kit.Types.SetProps(KiT_FileTreeView, FileTreeViewProps)
}

// FileNode returns the SrcNode as a *gide* FileNode
func (ft *FileTreeView) FileNode() *FileNode {
	fn := ft.SrcNode.Embed(KiT_FileNode)
	if fn == nil {
		return nil
	}
	return fn.(*FileNode)
}

// EditFiles calls EditFile on selected files
func (ft *FileTreeView) EditFiles() {
	sels := ft.SelectedViews()
	for i := len(sels) - 1; i >= 0; i-- {
		sn := sels[i]
		ftv := sn.Embed(KiT_FileTreeView).(*FileTreeView)
		fn := ftv.FileNode()
		if fn != nil {
			fn.EditFile()
		}
	}
}

// SetRunExec sets executable as the RunExec executable that will be run with Run / Debug buttons
func (ft *FileTreeView) SetRunExec() {
	sels := ft.SelectedViews()
	for i := len(sels) - 1; i >= 0; i-- {
		sn := sels[i]
		ftv := sn.Embed(KiT_FileTreeView).(*FileTreeView)
		fn := ftv.FileNode()
		if fn != nil {
			fn.SetRunExec()
			break
		}
	}
}

// RenameFiles calls RenameFile on any selected nodes
func (ftv *FileTreeView) RenameFiles() {
	fn := ftv.FileNode()
	if fn == nil {
		return
	}
	ge, ok := ParentGide(fn.This())
	if !ok {
		return
	}
	ge.SaveAllCheck(true, func() {
		var nodes []*FileNode
		sels := ftv.SelectedViews()
		for i := len(sels) - 1; i >= 0; i-- {
			sn := sels[i]
			ftvv := sn.Embed(KiT_FileTreeView).(*FileTreeView)
			fn := ftvv.FileNode()
			if fn != nil {
				nodes = append(nodes, fn)
			}
		}
		ge.CloseOpenNodes(nodes) // close before rename because we are async after this
		for _, fn := range nodes {
			giv.CallMethod(fn, "RenameFile", ftv.Viewport)
		}
	})
}

// FileTreeViewExecCmds gets list of available commands for given file node, as a submenu-func
func FileTreeViewExecCmds(it interface{}, vp *gi.Viewport2D) []string {
	ft, ok := it.(ki.Ki).Embed(KiT_FileTreeView).(*FileTreeView)
	if !ok {
		return nil
	}
	if ft.This() == ft.RootView.This() {
		ge, ok := ParentGide(ft.SrcNode)
		if !ok {
			return nil
		}
		return AvailCmds.FilterCmdNames(filecat.NoSupport, ge.VersCtrl())
	}
	fn := ft.FileNode()
	if fn == nil {
		return nil
	}
	ge, ok := ParentGide(fn.This())
	if !ok {
		return nil
	}
	lang := filecat.NoSupport
	if fn != nil {
		lang = fn.Info.Sup
	}
	cmds := AvailCmds.FilterCmdNames(lang, ge.VersCtrl())
	return cmds
}

// ExecCmdFiles calls given command on selected files
func (ft *FileTreeView) ExecCmdFiles(cmdNm string) {
	sels := ft.SelectedViews()
	if len(sels) > 1 {
		CmdWaitOverride = true // force wait mode
	}
	for i := len(sels) - 1; i >= 0; i-- {
		sn := sels[i]
		ftv := sn.Embed(KiT_FileTreeView).(*FileTreeView)
		if ftv.This() == ft.RootView.This() {
			if ft.SrcNode == nil {
				continue
			}
			ftr := ft.SrcNode.(*giv.FileTree)
			ge, ok := ParentGide(ftr)
			if ok {
				ge.ExecCmdNameFileName(string(ftr.FPath), CmdName(cmdNm), true, true)
			}
		} else {
			fn := ftv.FileNode()
			if fn != nil {
				fn.ExecCmdNameFile(cmdNm)
			}
		}
	}
	if CmdWaitOverride {
		CmdWaitOverride = false
	}
}

// FileTreeInactiveDirFunc is an ActionUpdateFunc that inactivates action if node is a dir
var FileTreeInactiveDirFunc = giv.ActionUpdateFunc(func(fni interface{}, act *gi.Action) {
	ft := fni.(ki.Ki).Embed(KiT_FileTreeView).(*FileTreeView)
	fn := ft.FileNode()
	if fn != nil {
		act.SetInactiveState(fn.IsDir())
	}
})

// FileTreeActiveDirFunc is an ActionUpdateFunc that activates action if node is a dir
var FileTreeActiveDirFunc = giv.ActionUpdateFunc(func(fni interface{}, act *gi.Action) {
	ft := fni.(ki.Ki).Embed(KiT_FileTreeView).(*FileTreeView)
	fn := ft.FileNode()
	if fn != nil {
		act.SetActiveState(fn.IsDir())
	}
})

// FileTreeActiveExecFunc is an ActionUpdateFunc that activates action if node is executable
var FileTreeActiveExecFunc = giv.ActionUpdateFunc(func(fni interface{}, act *gi.Action) {
	ft := fni.(ki.Ki).Embed(KiT_FileTreeView).(*FileTreeView)
	fn := ft.FileNode()
	if fn != nil {
		act.SetActiveState(fn.IsExec())
	}
})
