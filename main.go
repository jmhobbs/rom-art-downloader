package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type Rom struct {
	Platform string
	Path     string
	Name     string
	URL      string
	Error    error
}

type PlatformMapping struct {
	Name       string
	Extensions []string
}

var mappings []PlatformMapping = []PlatformMapping{
	PlatformMapping{"NES", []string{"nes"}},
	PlatformMapping{"SNES", []string{"smc"}},
	PlatformMapping{"N64", []string{"z64"}},
}

func usage(err string) {
	if err != "" {
		fmt.Println(err)
	}
	fmt.Printf("usage: %s <rom path> <image path>\n", os.Args[0])
	os.Exit(1)
}

func main() {
	if len(os.Args) != 3 {
		usage("")
	}

	log.Println("Gathering ROMs...")

	roms := []*Rom{}

	for _, mapping := range mappings {
		for _, extension := range mapping.Extensions {
			paths, err := filepath.Glob(filepath.Join(os.Args[1], "*."+extension))
			if err != nil {
				log.Fatal(err)
			}

			for _, fullpath := range paths {
				roms = append(roms, &Rom{mapping.Name, fullpath, strings.TrimSuffix(path.Base(fullpath), "."+extension), "", nil})
			}
		}
	}

	log.Println("Fetching ROM information...")

	var wg sync.WaitGroup
	src := make(chan *Rom)

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rom := range src {
				url, err := SearchForGame(rom.Platform, rom.Name)
				if err != nil {
					rom.Error = err
				} else {
					rom.URL = url
				}
			}
		}()
	}

	for _, rom := range roms {
		src <- rom
	}
	close(src)

	wg.Wait()

	fmt.Println("Downloading covers...")

	src = make(chan *Rom)

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rom := range src {
				out, err := os.Create(path.Join(os.Args[2], rom.Name+".png")) // TODO: PNG assumed here.
				if err != nil {
					rom.Error = err
					continue
				}
				defer out.Close() // TODO: This close is defered until loop exits.
				resp, err := http.Get(rom.URL)
				if err != nil {
					rom.Error = err
					continue
				}
				defer resp.Body.Close() // TODO: Likewise.
				_, err = io.Copy(out, resp.Body)
				if err != nil {
					rom.Error = err
					continue
				}
			}
		}()
	}

	for _, rom := range roms {
		if rom.Error != nil {
			continue
		}
		src <- rom
	}
	close(src)

	wg.Wait()

	for _, rom := range roms {
		if rom.Error != nil {
			log.Printf("[%s] %s - %s\n", rom.Platform, rom.Name, rom.Error)
		} else {
			log.Printf("[%s] %s - Downloaded!\n", rom.Platform, rom.Name)
		}
	}
}
