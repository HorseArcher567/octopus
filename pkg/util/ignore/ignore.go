package ignore

import "github.com/k8s-practice/octopus/pkg/log"

func Results(v ...interface{}) {
	log.Println(1, log.DebugLevel, v...)
}
