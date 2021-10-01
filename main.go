package main

type SafeSlice interface {
	Append(interface{})
	At(int) interface{}
	Close() []interface{}
	Delete(int)
	Len() int
	Update(int, UpdateFunc)
}

type UpdateFunc func(interface{}) interface{}

type commandData struct {
	action       commandAction
	index        int
	data         interface{}
	result       chan<- interface{}
	result_slice chan<- []interface{}
	updater      UpdateFunc
}

type commandAction int

const (
	apend commandAction = iota
	at
	end
	remove
	length
	update
)

type safeSlice chan commandData

func New() safeSlice {
	ss := make(safeSlice)
	go ss.run()
	return ss
}

func (ss safeSlice) run() {
	store := make([]interface{}, 0)
	for command := range ss {
		switch command.action {
		case apend:
			store = append(store, command.data)
		case at:
			if command.index >= 0 && command.index < len(store) {
				command.result <- store[command.index]
			} else {
				command.result <- nil
			}

		case end:
			command.result_slice <- store
			close(ss)
		case remove:
			if command.index >= 0 && command.index < len(store) {
				store = append(store[:command.index],
					store[command.index+1:]...)
			}

		case length:
			command.result <- len(store)
		case update:
			if 0 <= command.index && command.index < len(store) {
				value := store[command.index]
				store[command.index] = command.updater(value)
			}
		}
	}
}

func (ss safeSlice) Append(value interface{}) {
	ss <- commandData{action: apend, data: value}
}

func (ss safeSlice) At(i int) interface{} {
	reply := make(chan interface{})
	ss <- commandData{action: at, index: i, result: reply}
	result := <-reply
	return result
}

func (ss safeSlice) Close() []interface{} {
	reply := make(chan []interface{})
	ss <- commandData{action: end, result_slice: reply}
	result := <-reply
	return result
}

func (ss safeSlice) Delete(i int) {
	ss <- commandData{action: remove, index: i}
}

func (ss safeSlice) Len() int {
	reply := make(chan interface{})
	ss <- commandData{action: length, result: reply}
	result := (<-reply).(int)
	return result
}

func (ss safeSlice) Update(i int, updater UpdateFunc) {
	ss <- commandData{action: update, updater: updater, index: i}
}
