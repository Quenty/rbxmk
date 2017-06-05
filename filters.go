package main

import (
	"runtime"
)

type Filter func(f FilterArgs, arguments []interface{}) (results []interface{})

// FilterArgs is used by a Filter to indicate that it is done processing its
// arguments.
type FilterArgs interface {
	ProcessedArgs()
}

type filterArgs bool

func (f *filterArgs) ProcessedArgs() {
	*f = true
}

func CallFilter(filter Filter, arguments ...interface{}) (results []interface{}, err error) {
	var argsProcessed filterArgs
	defer func() {
		if !argsProcessed {
			err = recover().(runtime.Error)
		}
	}()
	results = filter(&argsProcessed, arguments)
	return
}
