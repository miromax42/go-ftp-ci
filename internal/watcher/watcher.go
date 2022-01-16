package watcher

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"
)

const (
	resPath  = "res"
	dirsFile = "res/dirs.json"
)

type Config struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Watcher struct {
	conn *syncConnection
	Dirs map[string]uint32

	wg *sync.WaitGroup
}

func New(cfg Config) (w Watcher, err error) {
	addr := fmt.Sprintf("%s:%v", cfg.Host, cfg.Port)

	c, err := ftp.Dial(addr, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		err = fmt.Errorf("ftp.Dial to %s: %v", addr, err)

		return
	}

	// defer c.Quit()

	err = c.Login(cfg.Username, cfg.Password)
	if err != nil {
		err = fmt.Errorf("Login to %s: %v", addr, err)

		return
	}

	m := &sync.Mutex{}
	w.conn = &syncConnection{
		ServerConn: c,
		m:          m,
	}
	w.wg = &sync.WaitGroup{}

	w.Dirs = make(map[string]uint32)
	w.loadDirs()

	return
}

func (w Watcher) Watch(ctx context.Context, dirs []string, updateTime time.Duration, notify chan string) {
	for _, dir := range dirs {
		w.wg.Add(1)
		go func(ctx context.Context, dir string) {
			defer w.wg.Done()
		loop:
			for {
				select {
				case <-ctx.Done():
					fmt.Printf("watcher for %s stoped\n", dir)
					break loop
				case <-time.After(updateTime):
					if w.CheckChanged(dir) {
						fmt.Printf("%v changed\n", dir)

						files := w.GetAllFiles(dir)
						w.LoadFiles(resPath, files)

						fmt.Printf("%v\n", files)
						w.GetHash(dir)

						notify <- dir

					} else {
						fmt.Printf("%v not changed\n", dir)
					}
				}
			}
		}(ctx, dir)
	}
}

func (w *Watcher) GetAllFiles(path string) (files []string) {
	list, _ := w.conn.ListSave(path)
	for _, e := range list {
		ePath := fmt.Sprintf("%s/%s", path, e.Name)
		if e.Type.String() == "folder" {
			recFiles := w.GetAllFiles(ePath)
			files = append(files, recFiles...)
		} else {
			files = append(files, ePath)
		}
	}

	return
}

func (w Watcher) LoadFiles(path string, files []string) {
	for i := range files {
		var buffer bytes.Buffer
		w.conn.StorSave(path, &buffer)
		filePath := fmt.Sprintf("%s%s", path, files[i])

		file, err := create(filePath)
		if err != nil {
			fmt.Printf("create file: %v\n", err)
		}

		_, err = file.Write(buffer.Bytes())
		if err != nil {
			fmt.Printf("Write file: %v\n", err)
		}

		file.Close()

	}
}

func (w *Watcher) Stop() {
	w.wg.Wait()

	w.dumpDirs()
	err := w.conn.Quit()
	if err != nil {
		log.Fatalf("Cant stop con: %v", err)
	}
}

func (w Watcher) GetHash(dir string) (h uint32, err error) {
	list, err := w.conn.ListSave(dir)
	//for _,e:=range list{
	//	//fmt.Printf("%v time %v",e.Type, e.Time.Format("StampMilli"))
	//}
	if err != nil {
		err = fmt.Errorf("w.conn.List for %v: %v", dir, err)
		return
	}

	var buffer bytes.Buffer

	for _, e := range list {
		str := fmt.Sprintf("%v%v", e.Time.Format(time.StampMilli), e.Size)
		buffer.Write([]byte(str))
	}

	files := w.GetAllFiles(dir)
	str := fmt.Sprintf("%v", files)
	buffer.Write([]byte(str))

	h = getHash(buffer.String())

	return
}

func (w *Watcher) CheckChanged(dir string) (changed bool) {
	newHash, err := w.GetHash(dir)
	if err != nil {
		fmt.Printf("w.GetHash: %v", err)
		return false
	}
	if w.Dirs[dir] != newHash {
		w.Dirs[dir] = newHash
		changed = true
	} else {
		changed = false
	}

	return
}

func Show(cfg *Config, path string) (list []*ftp.Entry, err error) {
	addr := fmt.Sprintf("%s:%v", cfg.Host, cfg.Port)
	c, err := ftp.Dial(addr, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		err = fmt.Errorf("ftp.Dial to %s: %v", addr, err)
		log.Fatal(err)
	}
	defer c.Quit()

	err = c.Login(cfg.Username, cfg.Password)
	if err != nil {
		log.Fatal(err)
	}

	// Do something with the FTP conn
	list, err = c.List(path)

	return
}
