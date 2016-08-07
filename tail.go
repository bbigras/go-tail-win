// Package tail behaves like tail
//
// Example
//    package main
//
//    import (
//		"fmt"
//		"github.com/brunoqc/go-tail-win"
//    )
//
//    func main() {
//		t, err := tail.TailFile(filePath)
//		if err != nil {
//			return err
//		}
//
//		for {
//			select {
//			case line := <-t.Lines:
//				fmt.Printf("> %q\n", line)
//			}
//		}
//    }
package tail

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/ActiveState/tail/winfile"
	"gopkg.in/fsnotify.v1"
)

type typeModif uint32

const (
	write typeModif = 1 << iota
	rename
)

type tail struct {
	mutexChose *sync.Mutex
	cond       *sync.Cond
	Lines      chan string
	chExit     chan struct{}
}

func (t tail) Clean() {
	close(t.chExit)
}

func read(r io.Reader, buf *bytes.Buffer, ch chan<- string) error {
	data, errRead := ioutil.ReadAll(r)
	if errRead != nil {
		return errRead
	}

	data2 := append(buf.Bytes(), data...)
	rendu := splitLine(ch, data2)

	buf.Reset()
	if rendu <= len(data2) {
		if _, err := buf.Write(data2[rendu:]); err != nil {
			return err
		}
	} else {
		log.Fatal("erreur: rendu <= len(data2)")
	}

	return nil
}

func openAndFollow(path string, ch chan<- string, chModif <-chan typeModif, aRotated bool, rep *tail) (bool, error) {
	fi, errOpen := winfile.OpenFile(path, os.O_RDONLY, 0)
	if errOpen != nil {
		return aRotated, errOpen
	}
	defer fi.Close()

	if !aRotated {
		// se positionner à la fin
		if _, err := fi.Seek(0, 2); err != nil {
			return aRotated, err
		}
	}

	rep.cond.L.Lock()
	rep.cond.Signal()
	rep.cond.L.Unlock()

	var rotation bool
	var buf bytes.Buffer

	for {
		if rotation {
			return true, nil
		}

		v, ok := <-chModif
		if !ok {
			return false, errors.New("chan fermé")
		}

		if v == rename {
			rotation = true
			log.Println("rotation!")
		}

		if err := read(fi, &buf, ch); err != nil {
			return aRotated, err
		}

		aRotated = false
	}
}

func checkModifLog(path string, rep *tail) <-chan typeModif {
	// TODO: 1 et drop les events quand plein?
	chModif := make(chan typeModif, 1000)

	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		errWatch := watcher.Add(filepath.Dir(path))
		if errWatch != nil {
			log.Fatal(errWatch)
		}

		for {
			select {
			case <-rep.chExit:
				log.Println("rep.chExit,2")
				return
			case event := <-watcher.Events:
				if event.Name == path {
					if event.Op&fsnotify.Write == fsnotify.Write {
						chModif <- write
					} else if event.Op&fsnotify.Rename == fsnotify.Rename {
						chModif <- rename
					}
				}
			case err := <-watcher.Errors:
				log.Println("erreur", err)
			}
		}
	}()

	return chModif
}

func splitLine(ch chan<- string, data []byte) int {
	var dernier int

	for i, un := range data {
		if un == '\n' {
			ch <- string(data[dernier:i])
			dernier = i + 1
		}
	}

	return dernier
}

func TailFile(path string) (*tail, error) {
	var mutexChose sync.Mutex

	rep := tail{
		mutexChose: &mutexChose,
		cond:       sync.NewCond(&mutexChose),
		Lines:      make(chan string, 1000),
		chExit:     make(chan struct{}, 1),
	}

	go func() {
		defer func() {
			log.Println("sortie go func chose de tail()")
			if e := recover(); e != nil {
				log.Printf("%s: %s", e, debug.Stack())
				os.Exit(1)
			}
		}()

		// event quand le log est renommé
		chLogModif := checkModifLog(path, &rep)

		var rotated bool
		for {
			select {
			case <-rep.chExit:
				log.Println("rep.chExit.1")
				return
			default:
				r, err := openAndFollow(path, rep.Lines, chLogModif, rotated, &rep)
				if err != nil {
					if err == syscall.ENOENT {
						log.Println("erreur os.ENOENT", err)
					} else {
						log.Printf("erreur: %T %s", err, err)
					}
					time.Sleep(1 * time.Second)
				}
				rotated = r
			}
		}
	}()

	rep.cond.L.Lock()
	rep.cond.Wait()
	rep.cond.L.Unlock()

	return &rep, nil
}
