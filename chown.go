package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
)

func Walk(wg *sync.WaitGroup, path string, uid, gid uint32) error {
	wg.Add(1)
	defer wg.Done()
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error while opening file %v:%v\n", path, err)
		return err
	}

	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		fmt.Printf("Error while Readdir %v:%v\n", path, err)
		return err
	}
	for _, file := range list {
		p := filepath.Join(path, file.Name())
		pUid := file.Sys().(*syscall.Stat_t).Uid
		pGid := file.Sys().(*syscall.Stat_t).Gid
		if uid == pUid && gid == pGid {
			err := os.Lchown(p, int(uid), int(gid))
			if err != nil {
				fmt.Printf("Error while chowning file %v:%v\n",
					path, err)
				return err
			}
		}
		if file.IsDir() {
			go Walk(wg, p, uid, gid)
		}
	}
	return nil
}

func main() {
	var wg sync.WaitGroup
	var rLimit syscall.Rlimit

	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)

	if err != nil {
		fmt.Println("Error Getting Rlimit ", err)
		os.Exit(1)
	}

	rLimit.Cur = rLimit.Max

	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println("Error Setting Rlimit ", err)
		os.Exit(1)
	}
	uid, _ := strconv.ParseUint(os.Args[2], 10, 32)
	gid, _ := strconv.ParseUint(os.Args[3], 10, 32)
	err = os.Lchown(os.Args[1], int(uid), int(gid))
	if err != nil {
		fmt.Printf("Error while chowning dir %v:%v\n", os.Args[1], err)
		os.Exit(1)
	}

	Walk(&wg, os.Args[1], uint32(uid), uint32(gid))
	wg.Wait()
}
