package main

import (
	"os"
	"runtime/pprof"
	"strconv"
	"strings"

	"github.com/qiniu/log.v1"
)

func fact(n int) int {
	if n == 0 {
		return 1
	}
	return n * fact(n-1)
}

func Main() {
	f, err := os.Create("write.pprof")
	if err != nil {
		log.Fatal(err)
	}

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	base := []string{}
	cmp := []string{}
	arrayLength := 1000
	commonLength := arrayLength
	for i := 0; i < arrayLength; i++ {
		base = append(base, "abc"+strconv.Itoa(i))
		cmp = append(base, "abc"+strconv.Itoa(i+arrayLength-commonLength))
	}

	for i := 0; i < 1000; i++ {
		findSetDifference(base, cmp)
		findSetDifference1(base, cmp)
	}

	loop0()
	loop1()
	loop2()
	loop3()

	sum := 0
	for i := 0; i < 50; i++ {
		sum += fact(i)
	}

	for {
	}

	log.Println(sum)

}

func loop0() {
	for i := 0; i < 1000000000; i++ {

	}
}

func loop1() {
	for i := 0; i < 1000000000; i++ {

	}
}

func loop2() {
	for i := 0; i < 1000000000; i++ {

	}
}

func loop3() {
	for i := 0; i < 1000000000; i++ {

	}
}

func findSetDifference(base, cmp []string) (common, subInBase, subInCmp []string) {

	if len(base) == 0 {
		subInCmp = cmp
		return
	}

	if len(cmp) == 0 {
		subInBase = base
		return
	}

	swap := func(i, j *string) {
		tmp := *i
		*i = *j
		*j = tmp
	}

	head, tail := 0, len(base)
	front, end := 0, len(cmp)
	for head < tail {
		target := strings.Split(base[head], "{}")[0]
		j := front
		for j < end {
			if target == strings.Split(cmp[j], "{}")[0] {
				swap(&cmp[j], &cmp[head])
				break
			}
			j++
		}
		if j == end {
			tail--
			swap(&base[head], &base[tail])
			continue
		} else {
			front++
			head++
		}
	}

	subInBase = base[tail:]
	common = base[:head]
	subInCmp = cmp[front:]
	return
}

func findSetDifference1(base, cmp []string) (common, subInBase, subInCmp []string) {

	if len(base) == 0 {
		subInCmp = cmp
		return
	}

	if len(cmp) == 0 {
		subInBase = base
		return
	}

	swap := func(i, j *string) {
		tmp := *i
		*i = *j
		*j = tmp
	}

	head, tail := 0, len(base)
	front, end := 0, len(cmp)
	for head < tail {
		target := strings.Split(base[head], "{}")[0]
		j := front
		for j < end {
			if target == strings.Split(cmp[j], "{}")[0] {
				swap(&cmp[j], &cmp[head])
				break
			}
			j++
		}
		if j == end {
			tail--
			swap(&base[head], &base[tail])
			continue
		} else {
			front++
			head++
		}
	}

	subInBase = base[tail:]
	common = base[:head]
	subInCmp = cmp[front:]
	return
}
