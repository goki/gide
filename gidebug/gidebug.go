// Copyright (c) 2020, The Gide Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gidebug

import (
	"errors"
	"time"

	"github.com/goki/gi/giv"
)

var NotStartedErr = errors.New("debugger not started")

var IsRunningErr = errors.New("debugger is currently running and cannot return info")

// GiDebug is the interface for all supported debuggers.
// It is based directly on the Delve Client interface.
type GiDebug interface {

	// HasTasks returns true if the debugger supports a level of threading
	// below the system thread level.  If true, then use Task data
	// otherwise, use Threads
	HasTasks() bool

	// Start starts the debugger for a given exe path, and overall project
	// root path (for trimming file names), and sets output of debugger
	// session to given textbuf which is used to monitor output.
	// params must have relevant settings in place (StatFunc, Mode, etc).
	Start(path, rootPath string, outbuf *giv.TextBuf, pars *Params) error

	// SetParams sets the current parameters to control how info is returned
	SetParams(params *Params)

	// IsActive returns true if the debugger is active and ready for commands
	IsActive() bool

	// Returns the pid of the process we are debugging.
	ProcessPid() int

	// LastModified returns the time that the process' executable was modified.
	LastModified() time.Time

	// Detach detaches the debugger, optionally killing the process.
	Detach(killProcess bool) error

	// Disconnect closes the connection to the server without
	// sending a Detach request first.  If cont is true a continue
	// command will be sent instead.
	Disconnect(cont bool) error

	// Restarts program.
	Restart() error

	// GetState returns the current debugger state.
	// This will return immediately -- if the target is running then
	// the Running flag will be set and a Stop bus be called to
	// get any further information about the target.
	GetState() (*State, error)

	// Continue resumes process execution.
	Continue() <-chan *State

	// Rewind resumes process execution backwards. (What does this do??)
	Rewind() <-chan *State

	// StepOver continues to the next source line, not entering function calls.
	StepOver() (*State, error)

	// StepInto continues to the next source line, entering function calls.
	StepInto() (*State, error)

	// StepOut continues to the return point of the current function
	StepOut() (*State, error)

	// StepSingle step a single cpu instruction.
	StepSingle() (*State, error)

	// SwitchThread switches the current system thread context to given one
	SwitchThread(threadID int) (*State, error)

	// SwitchTask switches the current thread to given one
	SwitchTask(threadID int) (*State, error)

	// Stop suspends the process.
	Stop() (*State, error)

	// GetBreak gets info about a breakpoint by ID.
	GetBreak(id int) (*Break, error)

	// SetBreak sets a new breakpoint at given file (must be enough to be unique)
	// and line number
	SetBreak(fname string, line int) (*Break, error)

	// ListBreaks gets all breakpoints.
	ListBreaks() ([]*Break, error)

	// ClearBreak deletes a breakpoint by ID.
	ClearBreak(id int) error

	// AmmendBreak updates the Condition and Trace information
	// for the given breakpoint
	AmendBreak(id int, cond string, trace bool) error

	// UpdateBreaks updates current breakpoints based on given list of breakpoints.
	// first gets the current list, and does actions to ensure that the list is set.
	UpdateBreaks(brk *[]*Break) error

	// Cancels a Next or Step call that was interrupted by a
	// manual stop or by another breakpoint
	CancelNext() error

	// InitAllState initializes the given AllState with relevant info
	// for current debugger state.  Does NOT get AllVars.
	InitAllState(all *AllState) error

	// UpdateAllState updates the state for given threadId and
	// frame number (only info different from current results is updated).
	// For given thread (lowest-level supported by language,
	// e.g., Task if supported, else Thread), and frame number.
	UpdateAllState(all *AllState, threadID int, frame int) error

	// CurThreadID returns the proper current threadID (task or thread)
	// based on debugger, from given state.
	CurThreadID(all *AllState) int

	// ListThreads lists all system threads.
	ListThreads() ([]*Thread, error)

	// GetThread gets a thread by its ID.
	GetThread(id int) (*Thread, error)

	// ListTasks lists all the currently active threads (if supported)
	ListTasks() ([]*Task, error)

	// Stack returns the current stack, up to given depth,
	// for given thread (lowest-level supported by language,
	// e.g., Task if supported, else Thread), and frame number.
	Stack(threadID int, depth int) ([]*Frame, error)

	// ListAllVars lists all variables (subject to filter) in the context
	// of the current thread.
	ListAllVars(filter string) ([]*Variable, error)

	// ListVars lists all stack-frame local variables (including args)
	// for given thread (lowest-level supported by language,
	// e.g., Task if supported, else Thread), and frame number.
	ListVars(threadID int, frame int) ([]*Variable, error)

	// GetVar returns a variable for given thread (lowest-level supported by
	// language -- e.g., Task if supported, else Thread), and frame number.
	GetVar(name string, threadID int, frame int) (*Variable, error)

	// SetVar sets the value of a variable.
	// for given thread (lowest-level supported by language,
	// e.g., Task if supported, else Thread), and frame number.
	SetVar(name, value string, threadID int, frame int) error

	// ListSources lists all source files in the process matching filter.
	ListSources(filter string) ([]string, error)

	// ListFuncs lists all functions in the process matching filter.
	ListFuncs(filter string) ([]string, error)

	// ListTypes lists all types in the process matching filter.
	ListTypes(filter string) ([]string, error)
}

// Modes are different modes of running the debugger
type Modes int32

const (
	// Exec means debug a standard executable program
	Exec Modes = iota

	// Test means debug a testing program
	Test

	// Attach means attach to an already-running process
	Attach
)