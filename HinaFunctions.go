package Hina

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	_ "image/png"
	"math/rand"
	"os"
	"sync"
)

func AbsSub[T constraints.Integer](a, b T) T {
	if a < b {
		return b - a
	} else {
		return a - b
	}
}

func LoadDirs(dir string) []string {

	// Open the directory
	fileDir, err := os.Open(dir)
	if err != nil {
		fmt.Println("Error:", err)
	}
	err = fileDir.Close()
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Read the directory entries
	entries, err := fileDir.ReadDir(-1) // -1 means read all entries
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Create an empty slice to store the folder names
	folders := make([]string, 0)

	// Loop through the entries and check if they are folders
	for _, entry := range entries {
		// Append the folder name to the slice
		folders = append(folders, dir+entry.Name())
	}

	return folders

}

func LoadImageToSlice(imgDir string, rgb bool) []uint8 {

	img, err := os.Open(imgDir)
	if err != nil {
		panic(err)
	}

	decoded, _, err := image.Decode(img)
	if err != nil {
		panic(err)
	}
	err = img.Close()
	if err != nil {
		fmt.Println("Error:", err)
	}

	switch rgb {
	case false:
		grey := image.NewGray(decoded.Bounds())
		draw.Draw(grey, grey.Bounds(), decoded, decoded.Bounds().Min, draw.Src)
		return grey.Pix
	case true:
		colors := image.NewRGBA(decoded.Bounds())
		draw.Draw(colors, colors.Bounds(), decoded, decoded.Bounds().Min, draw.Src)
		return colors.Pix
	}

	return nil

}

func SaveImage(inputimg []uint8, w int, h int, location string) {

	var wg sync.WaitGroup

	pixels := make([][]color.RGBA, h)
	for y := range pixels {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()
			pixels[y] = make([]color.RGBA, w)
			for x := range pixels[y] {
				index := y*w*4 + x*4
				pixels[y][x] = color.RGBA{
					R: inputimg[index],
					G: inputimg[index+1],
					B: inputimg[index+2],
					A: inputimg[index+3],
				}
			}
		}(y)
		wg.Wait()
	}

	// Create a new RGBA image
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Fill the image with the pixel colors from the array
	for y := 0; y < h; y++ {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()
			for x := 0; x < w; x++ {
				img.Set(x, y, pixels[y][x])
			}
		}(y)
		wg.Wait()

	}

	// Create or open the output file
	outFile, err := os.Create(location)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	// Save the image as a PNG
	err = png.Encode(outFile, img)
	if err != nil {
		panic(err)
	}

}

type Dataset struct {
	Images [][]uint8
	Labels []uint8
}

func (s *Dataset) Shuffle() {
	rand.Shuffle(len(s.Images), func(i, j int) {
		s.Images[i], s.Images[j] = s.Images[j], s.Images[i]
		s.Labels[i], s.Labels[j] = s.Labels[j], s.Labels[i]
	})
}

func Lerp(start, end, p float64) float64 {
	return (1-p)*start + p*end
}

func InverseLerp(start, end, value float64) float64 {
	return (value - start) / (end - start)
}

func PercentDif(a, b float64) float64 {
	if b == 0 {
		return 0.0
	}
	return (a / b) * 100
}

func InsertionSort(arr [][]int) {
	for i := 1; i < len(arr[0]); i++ {
		key1 := arr[0][i]
		key2 := arr[1][i]
		j := i - 1

		for j >= 0 && arr[0][j] > key1 {
			arr[0][j+1] = arr[0][j]
			arr[1][j+1] = arr[1][j]
			j = j - 1
		}
		arr[0][j+1] = key1
		arr[1][j+1] = key2
	}
}

func NNResize(input []uint8, inputW, inputH, outputW, outputH, colorChannels int) []uint8 {
	output := make([]uint8, outputW*outputH*colorChannels)

	xRatio := float64(inputW) / float64(outputW)
	yRatio := float64(inputH) / float64(outputH)

	for y := 0; y < outputH; y++ {
		for x := 0; x < outputW; x++ {
			srcX := int(float64(x) * xRatio)
			srcY := int(float64(y) * yRatio)

			srcIndex := (srcY*inputW + srcX) * colorChannels
			dstIndex := (y*outputW + x) * colorChannels

			for c := 0; c < colorChannels; c++ {
				output[dstIndex+c] = input[srcIndex+c]
			}
		}
	}

	return output
}

func BuildDataset(dir string) Dataset {

	folders := LoadDirs(dir)

	ds := Dataset{
		Images: make([][]uint8, 0),
		Labels: make([]uint8, 0),
	}

	var target uint8 = 0
	for f := range folders {
		images := LoadDirs(folders[f] + "\\")

		for i := range images {
			ds.Images = append(ds.Images, LoadImageToSlice(images[i], false))
			ds.Labels = append(ds.Labels, target)
		}

		target++
	}

	return ds

}
