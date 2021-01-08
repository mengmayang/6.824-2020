package wc

import (
	"lab01/mr"
	"strconv"
	"strings"
	"unicode"
)
//
////ByKey 需要实现sort.Interface
type ByKey []mr.KeyValue
//
//func (a ByKey) Len() int { return len(a) }
//func (a ByKey) Less(i, j int) bool {return a[i].Key < a[j].Key}
//func (a ByKey) Swap(i, j int) {a[i], a[j] = a[j], a[i]}
//
func Map(contents string) []mr.KeyValue{
	ff := func(r rune) bool {
		//Go语言的字符 rune,rune 类型等价于 int32 类型
		//=是赋值的意思，:=是声明变量并赋值的意思
		return !unicode.IsLetter(r)
	}

	words := strings.FieldsFunc(contents, ff)
	kva := []mr.KeyValue{}

	for _, w := range words {
		kv := mr.KeyValue{w, "1"}
		kva = append(kva, kv)
	}
	return kva
}

func Reduce(key string, values []string) string {
	return strconv.Itoa(len(values))
}
