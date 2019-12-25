package conn

import "sync"

var ConnectionMap sync.Map
var ImieMap sync.Map
func GetDefaultGroup() string {
	defaultgroup:="all"
	return defaultgroup
}
func GetDefaultEvent() string {
	return "task"
}