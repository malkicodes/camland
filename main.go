package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

const IMAGES_BASE_URL string = "https://images2.pixel.land/0x0/"
const OVERWORLD string = "db9238ed-8377-4600-9b17-c0ecd06c3f23/97a548ca-4ecc-4776-96f4-f82af16137b0"
const NETHER string = "db9238ed-8377-4600-9b17-c0ecd06c1111/0683de2a-99bb-4943-9787-be88da4c0f9a"

const SCREENSHOTS_BASE_URL string = "https://screenshots-4bn7kvgogq-uc.a.run.app/"

//go:embed data/custom_canvas_query.txt
var CUSTOM_CANVAS_QUERY string

func getFilename(dimension string) string {
	return "camland_" + dimension + "_" + time.Now().UTC().Format(time.RFC3339) + ".png"
}

func getFilenameGif(dimension string) string {
	return "camland_" + dimension + "_" + time.Now().UTC().Format(time.RFC3339) + ".gif"
}

func fetchImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	img, err := png.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func getOverworldChunk(i int) (image.Image, error) {
	x := i % 25
	y := int(i / 25)

	chunkId := y*128 + x + 6708

	imageUrl := IMAGES_BASE_URL + OVERWORLD + "/" + fmt.Sprint(chunkId) + ".png"

	return fetchImage(imageUrl)
}

func getNetherChunk(i int) (image.Image, error) {
	x := i % 5
	y := int(i / 5)

	chunkId := y*128 + x + 7998

	imageUrl := IMAGES_BASE_URL + NETHER + "/" + fmt.Sprint(chunkId) + ".png"

	return fetchImage(imageUrl)
}

func main() {
	dimensionFlag := flag.String("d", "overworld", "dimension: \"overworld\", \"nether\", or canvas id")
	outputFlag := flag.String("o", "", "output file")
	gifFlag := flag.Bool("gif", false, "if downloading custom dimension, save it as a gif")
	maxDownloaders := flag.Int("max-downloaders", 32, "maximum concurrent download tasks")

	flag.Parse()

	dimension := *dimensionFlag

	filename := getFilename(dimension)

	if *gifFlag {
		filename = getFilenameGif(dimension)
	}

	if *outputFlag != "" {
		filename = *outputFlag
	}

	err := os.MkdirAll("./output", 0750)
	if err != nil {
		fmt.Printf("error creating directory: %s\n", err)
		os.Exit(1)
	}

	if dimension == "overworld" {
		fmt.Println("Downloading overworld...")

		overworldImg := image.NewRGBA(image.Rect(0, 0, 12800, 12800))
		mu := sync.Mutex{}

		wg := sync.WaitGroup{}
		tokens := make(chan struct{}, *maxDownloaders)

		for i := range 625 {
			tokens <- struct{}{}
			wg.Add(1)

			go func() {
				defer wg.Done()
				defer func() { <-tokens }()

				fmt.Println(color.BlackString("Downloading chunk %d", i))

				img, err := getOverworldChunk(i)
				if err != nil {
					color.Red("error retrieving chunk %d: %s\n", i, err)
					os.Exit(1)
				}

				startX := (i % 25 * 512)
				startY := 512*25 - ((int(i/25) * 512) + 512)

				fmt.Println(color.BlackString("Pasting chunk %d at (%d,%d)", i, startX, startY))

				mu.Lock()
				draw.Draw(overworldImg, image.Rectangle{
					image.Pt(startX, startY),
					image.Pt(startX+512, startY+512),
				}, img, image.Pt(0, 0), draw.Src)
				mu.Unlock()
			}()
		}

		wg.Wait()

		func() {
			file, err := os.Create(filename)
			if err != nil {
				color.Red("failed to create file: %v", err)
				os.Exit(1)
			}
			defer file.Close()

			fmt.Println("Writing image to " + filename + "...")

			err = png.Encode(file, overworldImg)
			if err != nil {
				log.Fatalf("failed to encode image: %v", err)
			}
		}()
	} else if dimension == "nether" {
		fmt.Println("Downloading nether...")

		netherImg := image.NewRGBA(image.Rect(0, 0, 2560, 2560))
		mu := sync.Mutex{}

		wg := sync.WaitGroup{}
		tokens := make(chan struct{}, *maxDownloaders)

		for i := range 25 {
			tokens <- struct{}{}
			wg.Add(1)

			go func() {
				defer wg.Done()
				defer func() { <-tokens }()

				fmt.Println(color.BlackString("Downloading chunk %d", i))

				img, err := getNetherChunk(i)
				if err != nil {
					color.Red("error retrieving chunk %d: %s\n", i, err)
					os.Exit(1)
				}

				startX := (i % 5 * 512)
				startY := 512*5 - ((int(i/5) * 512) + 512)

				fmt.Println(color.BlackString("Pasting chunk %d at (%d,%d)", i, startX, startY))

				mu.Lock()
				draw.Draw(netherImg, image.Rectangle{
					image.Pt(startX, startY),
					image.Pt(startX+512, startY+512),
				}, img, image.Pt(0, 0), draw.Src)
				mu.Unlock()
			}()
		}

		wg.Wait()

		func() {
			file, err := os.Create(filename)
			if err != nil {
				color.Red("failed to create file: %v", err)
				os.Exit(1)
			}
			defer file.Close()

			fmt.Println("Writing image to " + filename + "...")

			err = png.Encode(file, netherImg)
			if err != nil {
				log.Fatalf("failed to encode image: %v", err)
			}
		}()
	} else {
		fmt.Printf("Downloading dimension %s...\n", dimension)

		body, err := json.Marshal(map[string]any{
			"operationName": "world",
			"query":         CUSTOM_CANVAS_QUERY,
			"variables": map[string]any{
				"slug": dimension,
			},
		})
		if err != nil {
			color.Red("error creating body: %s\n", err)
			os.Exit(1)
		}

		req, err := http.NewRequest(
			http.MethodPost,
			"https://worlds2.pixel.land/graphql",
			bytes.NewBuffer(body),
		)

		req.Header.Add("Accept", "*/*")
		req.Header.Add("Content-Type", "application/json")

		if err != nil {
			color.Red("error creating http request: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("Sending request for world data...")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			color.Red("error making http request: %s\n", err)
			os.Exit(1)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			color.Red("error reading http response: %s\n", err)
			os.Exit(1)
		}

		respData := make(map[string]any)
		err = json.Unmarshal(respBody, &respData)
		if err != nil {
			color.Red("error decoding http response: %s\n", err)
			os.Exit(1)
		}

		errs, prs := respData["errors"]

		if prs {
			color.Red("errors in world response: %s\n", errs)
			os.Exit(1)
		}

		worldId := respData["data"].(map[string]any)["world"].(map[string]any)["node"].(map[string]any)["id"].(string)

		imageUrl := SCREENSHOTS_BASE_URL + worldId + "/" + "canvas.png"

		if *gifFlag {
			imageUrl = SCREENSHOTS_BASE_URL + worldId + "/" + "animation.gif"
		}

		println(imageUrl)

		fmt.Println("Sending request for canvas image...")
		resp, err = http.Get(imageUrl)
		if err != nil {
			color.Red("error in image response: %s\n", err)
			os.Exit(1)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			color.Red("error when reading image: %s\n", err)
			os.Exit(1)
		}

		err = os.WriteFile(filename, data, 0755)
		if err != nil {
			color.Red("error when writing image: %s\n", err)
			os.Exit(1)
		}
	}

	color.Green("Wrote image to " + filename)
}
