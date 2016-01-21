package main

import (
    "fmt"
    "os"
    "runtime"
    "time"
    "strings"
    "flag"

    "math"
    "math/cmplx"

    "image"
    "image/png"
    _ "image/jpeg"
    _ "image/gif"
)

var NumCPU, goTines int

const RGB = 3
const BW = 1
const SPEC_COMPONENTS = 1
const SPEC_COLOR = 0

var processRGB bool
var combine bool
var inverse bool
var cornered bool

func init() {
    flag.BoolVar(&processRGB, "rgb", false, "rgb")
    flag.BoolVar(&combine, "c", false, "combine")
    flag.BoolVar(&inverse, "i", false, "inverse")
    flag.BoolVar(&cornered, "n", false, "blah")
}

func main() {
    NumCPU = runtime.NumCPU()
    runtime.GOMAXPROCS(NumCPU)
    goTines = NumCPU * NumCPU

    flag.Parse()
    filename := flag.Arg(0)
    name := strings.Split(filename, ".")[0]

    var channels int = BW
    var spec int = SPEC_COMPONENTS

    if processRGB {
        channels = RGB
    }
    if combine {
        spec = SPEC_COLOR
    }

    ft := NewFT(channels, spec)

    var d time.Duration
    t := time.Now()
    if !inverse {
        ft.decodeSignal(filename)
        t := time.Now()
        ft.computeSpectrum()
        d = time.Now().Sub(t)
        ft.encodeSpectrum(name + "_spec.png")
    } else {
        ft.decodeSpectrum(filename)
        t := time.Now()
        ft.computeSignal()
        d = time.Now().Sub(t)
        ft.encodeSignal(name + "_sig.png")
    }
    fmt.Printf("%v\n", time.Now().Sub(t))
    fmt.Printf("%v\n", d)
}

type FT struct {
    channels *Stack
    temp *Sheet
    w, h, maxDim int
    mode, spec int
}

func NewFT(channels, spec int) *FT {
    ft := FT{}
    ft.mode = channels
    ft.spec = spec

    ft.channels = NewStack(channels, ft.w, ft.h)
    ft.temp = NewSheet(ft.h, ft.w)
    return &ft
}

func (ft *FT) initData() {
    ft.channels = NewStack(ft.mode, ft.w, ft.h)
    ft.temp = NewSheet(ft.h, ft.w)
    if ft.w > ft.h {
        ft.maxDim = ft.w
    } else {
        ft.maxDim = ft.h
    }
}

func loadImage(filename string) (im image.Image) {
    fileIn, _ := os.Open(filename)
    defer fileIn.Close()
    im, _, _ = image.Decode(fileIn)
    return
}

func (ft *FT) decodeSignal(filename string) {
    im := loadImage(filename)
    ft.w = im.Bounds().Dx()
    ft.h = im.Bounds().Dy()
    ft.initData()

    for x := 0; x < ft.w; x++ {
        for y := 0; y < ft.h; y++ {
            f := func(x float64) complex128 {
                return complex(x, 0.0)
            }
            ft.channels.StoreColor(x, y, im.At(x, y), f)
        }
    }
}

func (ft *FT) decodeSpectrum(filename string) {
    im := loadImage(filename)
    ft.w = im.Bounds().Dx() / ft.mode
    ft.h = im.Bounds().Dy()
    ft.initData()

    for x := 0; x < ft.w; x++ {
        for y := 0; y < ft.h; y++ {
            for i := 0; i < ft.mode; i++ {
                f := func(r, g, b float64) complex128 {
                    H, _, L := RGBToHSL(r, g, b)
                    L *= math.Log(float64(ft.w * ft.h))
                    return cmplx.Exp(complex(L, 2.0 * math.Pi * H))
                }
                ft.channels.StoreComplex(i, x, y, im.At(x + ft.w * i, y), f)
            }
        }
    }
}

func saveImage(image image.Image, filename string) {
    fileOut, _ := os.Create(filename)
    defer fileOut.Close()
    png.Encode(fileOut, image)
}

func (ft *FT) encodeSignal(filename string) {
    im := image.NewRGBA(image.Rect(0, 0, ft.w, ft.h))
    for x := 0; x < ft.w; x++ {
        for y := 0; y < ft.h; y++ {
            im.SetRGBA(x, y, ft.channels.LoadColor(x, y, cmplx.Abs))
        }
    }
    saveImage(im, filename)
}

func (ft *FT) encodeSpectrum(filename string) {
    var rec image.Rectangle
    if ft.spec == SPEC_COLOR {
        rec = image.Rect(0, 0, ft.w, ft.h)
    } else {
        rec = image.Rect(0, 0, ft.w * ft.mode, ft.h)
    }
    im := image.NewRGBA(rec)
    if ft.spec == SPEC_COLOR {
        ft.encodeSpectrumColor(im)
    } else {
        for i := 0; i < ft.mode; i++ {
            imSub := im.SubImage(image.Rect(ft.w * i, 0, ft.w * (i + 1), ft.h)).(*image.RGBA)
            ft.encodeSpectrumComponent(i, imSub)
        }
    }
    saveImage(im, filename)
}

func (ft *FT) encodeSpectrumColor(im *image.RGBA) {
    xo, yo := (ft.w - 1) / 2, (ft.h - 1) / 2
    if cornered {
        xo, yo = 0, 0
    }
    for x := 0; x < ft.w; x++ {
        for y := 0; y < ft.h; y++ {
            f := func(c complex128) float64 {
                r := cmplx.Abs(c)
                if r > 1.0 {
                    return math.Log(r) / math.Log(float64(ft.w * ft.h))
                }
                return 0.0
            }
            im.SetRGBA((x + xo) % ft.w, (y + yo) % ft.h, ft.channels.LoadColor(x, y, f))
        }
    }
}

func (ft *FT) encodeSpectrumComponent(i int, im *image.RGBA) {
    xo, yo := (ft.w - 1) / 2, (ft.h - 1) / 2
    for x := 0; x < ft.w; x++ {
    if cornered {
        xo, yo = 0, 0
    }
        for y := 0; y < ft.h; y++ {
            f := func(c complex128) (float64, float64, float64) {
                r, theta := cmplx.Polar(c)
                if r > 1.0 {
                    r = math.Log(r) / math.Log(float64(ft.w * ft.h))
                } else {
                    r = 0.0
                }
                theta = math.Mod(theta / (2.0 * math.Pi) + 1.0, 1.0)
                return HSLToRGB(theta, 1.0, r)
            }
            im.SetRGBA((x + xo) % ft.w + ft.w * i, (y + yo) % ft.h, ft.channels.LoadComplex(i, x, y, f))
        }
    }
}

func (ft *FT) computeSpectrum() {
    for i := range ft.channels.sheets {
        ft.ft2D(ft.channels.sheets[i].data, -2.0 * math.Pi, 1.0)
    }
}

func (ft *FT) computeSignal() {
    for i := range ft.channels.sheets {
        ft.ft2D(ft.channels.sheets[i].data, 2.0 * math.Pi, 1.0 / float64(ft.w * ft.h))
    }
}

func (ft *FT) ft2D(data [][]complex64, dir, scale float64) {
    c := make(chan int, goTines)
    for i := 0; i < goTines; i++ {
        go func(i int) {
            scanLine := make([]complex64, ft.maxDim)
            for iy := i; iy < ft.h; iy += goTines {
                for ix := 0; ix < ft.w; ix++ {
                    scanLine[ix] = data[ix][iy]
                }
                fft(scanLine, ft.temp.data[iy], dir, 1)
            }
            c <- 1
        }(i)
    }
    for i := 0; i < goTines; i++ {
        <-c
    }

    for i := 0; i < goTines; i++ {
        go func(i int) {
            scanLine := make([]complex64, ft.maxDim)
            for ox := i; ox < ft.w; ox += goTines {
                for iy := 0; iy < ft.h; iy++ {
                    scanLine[iy] = ft.temp.data[iy][ox]
                }
                fft(scanLine, data[ox], dir, 1)
                for oy := 0; oy < ft.h; oy++ {
                    data[ox][oy] *= complex(float32(scale), 0.0)
                }
            }
            c <- 1
        }(i)
    }
    for i := 0; i < goTines; i++ {
        <-c
    }
}

func fft(input, output []complex64, dir float64, stride int) {
    N := len(output)
    if N == 1 {
        output[0] = input[0]
        return
    }
    N2 := N / 2
    fft(input, output[:N2], dir, stride * 2)
    fft(input[stride:], output[N2:], dir, stride * 2)
    for fx := 0; fx < N2; fx++ {
        o := output[fx + N2]
        o *= complex64(cmplx.Exp(complex(0.0, dir * float64(fx) / float64(N))))
        output[fx + N2] = output[fx] - o
        output[fx] += o
    }
}

func RGBToHSL(fR, fG, fB float64) (h, s, l float64) {
    max := math.Max(math.Max(fR, fG), fB)
    min := math.Min(math.Min(fR, fG), fB)
    l = (max + min) / 2.0
    if max == min {
        h, s = 0.0, 0.0
    } else {
        d := max - min
        if l > 0.5 {
            s = d / (2.0 - max - min)
        } else {
            s = d / (max + min)
        }
        switch max {
        case fR:
            h = (fG - fB) / d
            if fG < fB {
                h += 6.0
            }
        case fG:
            h = (fB - fR) / d + 2.0
        case fB:
            h = (fR - fG) / d + 4.0
        }
        h /= 6.0
    }
    return
}

func HSLToRGB(h, s, l float64) (r, g, b float64) {
    if s == 0 {
        r, g, b = l, l, l
    } else {
        var q float64
        if l < 0.5 {
                q = l * (1.0 + s)
        } else {
                q = l + s - s*l
        }
        p := 2.0 * l - q
        r = hueToRGB(p, q, h + 1.0 / 3.0)
        g = hueToRGB(p, q, h)
        b = hueToRGB(p, q, h - 1.0 / 3.0)
    }
    return
}

// helper function for HSLToRGB.
func hueToRGB(p, q, t float64) float64 {
    if t < 0.0 {
        t += 1.0
    } else if t > 1.0 {
        t -= 1.0
    }
    switch {
    case t < 1.0 / 6.0:
        return p + (q - p) * 6.0 * t
    case t < 0.5:
        return q
    case t < 2.0 / 3.0:
        return p + (q - p) * (2.0 / 3.0 - t) * 6.0
    default:
        return p
    }
}

