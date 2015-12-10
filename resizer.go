package main

import (
	"github.com/nfnt/resize"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"

	//_ "golang.org/x/image/bmp" this lib has troubles loading pictures
	_ "golang.org/x/image/tiff"

	"flag"
	"fmt"
	"io"
	"os"
)

// caller will get a 3 lines return on sdtout:
// -"OK" or "Error"
// -a codename (camelCase name of operation) - if Error
// -a string with details (err returned by failing function)- if Error
func exitWithResult(isSuccess bool, strCode string, details error) error {
	if isSuccess {
		fmt.Printf("OK\n\n\n")
		os.Exit(0)
	} else {
		fmt.Printf("Error\n")
		fmt.Printf("%s\n", strCode)
		fmt.Printf("%s\n", details)
		os.Exit(-1)
	}

	return nil //we are dead anyway
}

func getImageDimension(imagePath string) (int, int, error) {
    file, err := os.Open(imagePath)
    if err != nil {
        return 0,0,fmt.Errorf("%v\n", err)
    }

    image, _, err := image.DecodeConfig(file)
    if err != nil {
        return 0,0,fmt.Errorf("%s: %v\n", imagePath, err)
    }
    return image.Width, image.Height, nil
}


// file copy done with io.Copy
func copy(input string, output string) error {
	// open files r and w
	r, err := os.Open(input)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(output)
	if err != nil {
		return err
	}
	defer w.Close()

	// do the actual work
	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}
	return nil

}

// ugly map of input/output formats
func outputFormat(inputFormat string) string {
	if inputFormat == "png" {
		return "png"
	}
	if inputFormat == "gif" {
		return "gif"
	}
	return "jpeg"

}

// open a file and decode it into a *image.Image
func loadImage(inputFile string) (*image.Image, string, int64, error) {

	file, err := os.Open(inputFile)
	if err != nil {
		return nil, "", 0,err
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
    	return nil, "", 0, err
	}

	img, ftype, err := image.Decode(file)
	if err != nil {
		return nil, ftype, 0,err
	}
	return &img, ftype,fi.Size(), err
}

// encode/write a *image.Image to outputFile into outputFormat
func encode(outputFormat string, outputFile string, m image.Image, jpegQuality int) error {

	out, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()

	if outputFormat == "gif" {
		return gif.Encode(out, m, nil)
	} else if outputFormat == "png" {
		encoder:=png.Encoder{CompressionLevel:png.BestCompression }
		return encoder.Encode(out, m)
	} else if outputFormat == "jpeg" {
		jpgOptions := jpeg.Options{Quality: jpegQuality}
		return jpeg.Encode(out, m, &jpgOptions)
	}

	return exitWithResult(false, "openImage", fmt.Errorf("unknown associated output format"))

}

func main() {
	//by default 0 means do not modify unless needed
	// if width and height are set to zero, we blindly copy file
	width := flag.Uint("width", 0, "width, pass 0 if no constraint")
	height := flag.Uint("height", 0, "height, pass 0 if no constraint")
	quality := flag.Int("quality", 85, "export jpg quality, 85 by default (other formats not affected)")
	inputFile := flag.String("input", "", "input path")
	outputFile := flag.String("output", "", "output path (directory must exist)")
	flag.Parse()

	
	// we do not support bmp as golang.org/x/image/bmp seems buggy
	// all other formats are enabled by importing /image/ libraries.
	img, originalFormat, fileSize, err := loadImage(*inputFile)
	if err != nil {
		exitWithResult(false, "openImage", err)
		return
	}

	// if we are not going to change image size or format, just do a copy
	if originalFormat == outputFormat(originalFormat) && *width == 0 && *height == 0 {
		err := copy(*inputFile, *outputFile)
		if err != nil {
			exitWithResult(false, "flatCopy", err)
		}
		exitWithResult(true, "file copied", nil)
	}

	w,h, err :=getImageDimension(*inputFile)
	if err != nil {
		exitWithResult(false, "getImageDimension", err)
	}
	// if no value is given for a dimension, we keep current 
	if *height == 0 {
		*height = uint(h)
	}
	if *width == 0 {
		*width = uint(w)
	}

	// if original dimension fits max, just copy
	if *height>=uint(h) && *width>=uint(w) && fileSize< 150000{
		err := copy(*inputFile, *outputFile)
		if err != nil {
			exitWithResult(false, "flatCopy", err)
		}
		exitWithResult(true, "file copied", nil)
	}

	// dimensions are treated like max, proportions are kept
	m := resize.Thumbnail(*width, *height, *img, resize.Lanczos3)

	// write new image in chosen format
	err = encode(outputFormat(originalFormat), *outputFile, m, *quality)
	if err != nil {
		exitWithResult(false, "writeImage", err)
	} else {
		exitWithResult(true, "file resized", nil)
	}
}
