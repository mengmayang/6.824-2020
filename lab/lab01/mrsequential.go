package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type KeyValue struct {
	Key string
	Value string
}

//ByKey 需要实现sort.Interface
type ByKey []KeyValue

func (a ByKey) Len() int { return len(a) }
func (a ByKey) Less(i, j int) bool {return a[i].Key < a[j].Key}
func (a ByKey) Swap(i, j int) {a[i], a[j] = a[j], a[i]}

func Map(contents string) []KeyValue{
	ff := func(r rune) bool {
		//Go语言的字符 rune,rune 类型等价于 int32 类型
		//=是赋值的意思，:=是声明变量并赋值的意思
		return !unicode.IsLetter(r)
	}

	words := strings.FieldsFunc(contents, ff)
	kva := []KeyValue{}

	for _, w := range words {
		kv := KeyValue{w, "1"}
		kva = append(kva, kv)
	}
	return kva
}

func Reduce(key string, values []string) string {
	return strconv.Itoa(len(values))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: mrsequential inputfiles...\n")
		os.Exit(1)
	}

	//mapf, reducef := loadPlugin(os.Args[1])
	var mapf func(string) []KeyValue
	var reducef func(string, []string) string
	mapf = Map
	reducef = Reduce

	//intermediate是一个KeyValue的数组，每个key出现一次，就增加一个{word，"1"}
	intermediate := []KeyValue{}
	for _, filename := range os.Args[1:] {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatalf("cannot open %v", filename)
		}
		content, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatalf("cannot read %v", filename)
		}
		file.Close()
		kva := mapf(string(content))
		//使用append将两个切片合并，传两个切片，第二个参数要以...结尾
		intermediate = append(intermediate, kva...)
	}

	sort.Sort(ByKey(intermediate))

	oname := "mr-out-0"
	ofile, _ := os.Create(oname)

	i := 0
	for i < len(intermediate) {
		j := i + 1
		//将相同的key找出来（这也是排序的意义）
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j ++
		}

		//将拥有相同的key的键值对合并
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}
		output := reducef(intermediate[i].Key, values)

		fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)

		i = j
	}

	ofile.Close()
}

