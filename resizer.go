package main

import (
    "github.com/nfnt/resize"
    "image"
    "image/jpeg"
    "image/png"
    "image/gif"
    
   _ "golang.org/x/image/bmp" 
   _ "golang.org/x/image/tiff" 
    
    "os"
    "io"
    "fmt"
    "flag"
  //  "regexp"
   
)


func printFinalError(isSuccess bool, strCode string, details error) error{
  if isSuccess{
    fmt.Printf("OK\n\n\n")
    os.Exit(0)
  }else{
    fmt.Printf("Error\n")
    fmt.Printf("%s\n",strCode)
    fmt.Printf("%s\n",details)
    os.Exit(-1)
  }

  return nil //we are dead anyway
}

func copy (input string, output string) error{
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


func loadImage(inputFile string)  (*image.Image, string,  error){

  file, err := os.Open(inputFile)
  if err != nil {
    return nil, "", err
  }
  defer file.Close()

  img, ftype, err := image.Decode(file)
  if err != nil {
    return nil, ftype, err
  }
  return &img,  ftype,err
} 


func encode(ftype string, outputFile string, m image.Image ) error{
    
    out, err := os.Create(outputFile)
    if err != nil {
        return err
    }
    defer out.Close()

    if ftype=="gif"{
      return gif.Encode(out, m, nil)
    }else if ftype=="png"{
      return png.Encode(out, m)
    }else{
      jpgOptions:=jpeg.Options{Quality: 85}
      return jpeg.Encode(out, m, &jpgOptions)  
    }

    return printFinalError(false, "openImage", fmt.Errorf("unknown associated output format"))

}


func main() {
  //by default 0 means do not modify unless needed
  // if width and height are set to zero, we blindly copy file
  width := flag.Uint("width", 0, "width") 
  height := flag.Uint("height", 0, "height")
  inputFile := flag.String("input", "", "input path")
  outputFile := flag.String("output", "", "output path")
  flag.Parse()
    

  if *width == 0 && *height == 0 {
    err := copy(*inputFile, *outputFile)
    if err != nil {
      printFinalError(false, "flatCopy", err)
    }
    printFinalError(true, "file copied", nil)
  }

  img, ftype, err :=loadImage(*inputFile)
  if err != nil {
    printFinalError(false, "openImage", err)
    return
  }
  

  if *height==0{
    *height=uint((*img).Bounds().Max.Y)
  }
  if *width==0{
    *width=uint((*img).Bounds().Max.X)
  }

  // resize to width 1000 using Lanczos resampling
  // and preserve aspect ratio
  //m := resize.Resize(700, 0, *img, resize.Lanczos3)
  m := resize.Thumbnail(*width, *height, *img, resize.Lanczos3)

  err=encode(ftype, *outputFile, m)
  if err!=nil {
      printFinalError(false, "writeImage", err)
  }else {
      printFinalError(true, "file resized", nil)
	}
}

