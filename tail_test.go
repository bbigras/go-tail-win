package tail

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
	"flag"
)

var tmpDir string

func faitLe(lePath string, chFinGen chan struct{}, chResultat chan []string, chWatchFait chan struct{}) {
	defer close(chResultat)

	var liste []string

	t, err := TailFile(lePath)
	if err != nil {
		log.Fatal(err)
	}

	close(chWatchFait)

	for {
		select {
		case line := <-t.Lines:
			liste = append(liste, line)
		case <-chFinGen:
			chResultat <- liste
			return
		}
	}
}

func gen(lePath string, chFinGen chan struct{}, min, max, sleep int) {
	defer close(chFinGen)

	split := (max-min)/2 + min

	log.Println("lePath", lePath)

	func() {
		log.Println("open (write)")
		fo, errOpen := os.OpenFile(lePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if errOpen != nil {
			log.Fatal(errOpen)
		}
		defer fo.Close()

		for i := min; i <= split; i++ {
			fmt.Fprintf(fo, "%d\n", i)
			if err := fo.Sync(); err != nil {
				log.Fatal(err)
			}
			if sleep > 0 {
				time.Sleep(time.Duration(sleep) * time.Millisecond)
			}
		}
	}()

	dir1, file1 := filepath.Split(lePath)
	tmpFile := filepath.Join(dir1, file1+".tmp")

	os.Remove(tmpFile)
	if err := os.Rename(lePath, tmpFile); err != nil {
		log.Fatal(err)
	}

	func() {
		log.Println("open (write) 2")
		fo, errOpen := os.OpenFile(lePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if errOpen != nil {
			log.Fatal(errOpen)
		}
		defer fo.Close()

		for i := split + 1; i <= max; i++ {
			fmt.Fprintf(fo, "%d\n", i)
			if err := fo.Sync(); err != nil {
				log.Fatal(err)
			}
			if sleep > 0 {
				time.Sleep(time.Duration(sleep) * time.Millisecond)
			}
		}
	}()

}

func touch(lePath string) error {
	fo, errCreate := os.Create(lePath)
	if errCreate != nil {
		return errCreate
	}
	defer fo.Close()

	return nil
}

func genSlice(min, max int) []string {
	var resultat []string
	for i := min; i <= max; i++ {
		resultat = append(resultat, fmt.Sprintf("%d", i))
	}
	return resultat
}

func TestMain(m *testing.M) {
	flag.Parse()

	tmpDir = os.Getenv("TMP")
	if tmpDir == "" {
		tmpDir = "c:\\tmp"
	}
	if err := os.MkdirAll(tmpDir, 0777); err != nil {
		log.Fatalf("error creating %q: %s", tmpDir, err)
	}
	fmt.Println("tmpDir", tmpDir)
	
	os.Exit(m.Run())
}

func TestChose1(t *testing.T) {
	lePath := filepath.Join(tmpDir, "test-tail1")
	defer os.Remove(lePath)

	touch(lePath)

	chFinGen := make(chan struct{})
	chResultat := make(chan []string)
	chWatchFait := make(chan struct{})

	go faitLe(lePath, chFinGen, chResultat, chWatchFait)

	<-chWatchFait

	min := 1
	max := 55

	gen(lePath, chFinGen, min, max, 0)

	resultat := <-chResultat

	expect := genSlice(min, max)

	if !reflect.DeepEqual(resultat, expect) {
		t.Fail()
	}
}

func TestChose2(t *testing.T) {
	lePath := filepath.Join(tmpDir, "test-tail2")
	// lePath := filepath.Join(tmpDir, "chose")
	// lePath := "d:\\test\\chose"
	defer os.Remove(lePath)

	touch(lePath)

	chFinGen := make(chan struct{})
	chResultat := make(chan []string)
	chWatchFait := make(chan struct{})

	go faitLe(lePath, chFinGen, chResultat, chWatchFait)

	<-chWatchFait

	min := 1
	max := 55

	gen(lePath, chFinGen, min, max, 10)

	resultat := <-chResultat

	expect := genSlice(min, max)

	if !reflect.DeepEqual(resultat, expect) {
		t.Fail()
	}
}

func TestChose3(t *testing.T) {
	lePath := filepath.Join(tmpDir, "test-tail3")
	defer os.Remove(lePath)

	touch(lePath)

	chFinGen := make(chan struct{})
	chResultat := make(chan []string)
	chWatchFait := make(chan struct{})

	go faitLe(lePath, chFinGen, chResultat, chWatchFait)

	<-chWatchFait

	min := 1
	max := 55

	gen(lePath, chFinGen, min, max, 100)

	resultat := <-chResultat

	expect := genSlice(min, max)

	if !reflect.DeepEqual(resultat, expect) {
		t.Fail()
	}
}

func TestChose4(t *testing.T) {
	lePath := filepath.Join(tmpDir, "test-tail4")
	defer os.Remove(lePath)

	touch(lePath)

	chFinGen := make(chan struct{})
	chResultat := make(chan []string)
	chWatchFait := make(chan struct{})

	go faitLe(lePath, chFinGen, chResultat, chWatchFait)

	<-chWatchFait

	min := 1
	max := 55

	gen(lePath, chFinGen, min, max, 500)

	resultat := <-chResultat

	expect := genSlice(min, max)

	if !reflect.DeepEqual(resultat, expect) {
		t.Fail()
	}
}
