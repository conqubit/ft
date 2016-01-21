package main

import (
    "image/color"
)

type Sheet struct {
    data [][]complex64
}

func NewSheet(w, h int) *Sheet {
    var sh Sheet
    sh.data = make([][]complex64, w)
    for i := range sh.data {
        sh.data[i] = make([]complex64, h)
    }
    return &sh
}

func (sh *Sheet) Set(x, y int, z complex128) {
    sh.data[x][y] = complex64(z)
}

func (sh *Sheet) Get(x, y int) complex128 {
    return complex128(sh.data[x][y])
}






type Stack struct {
    sheets []*Sheet
}

func NewStack(sheets, w, h int) *Stack {
    var cs Stack
    cs.sheets = make([]*Sheet, sheets)
    for i := range cs.sheets {
        cs.sheets[i] = NewSheet(w, h)
    }
    return &cs
}


func (ch *Stack) Set(i, x, y int, z complex128) {
    ch.sheets[i].Set(x, y, z)
}

func (ch *Stack) Get(i, x, y int) complex128 {
    return ch.sheets[i].Get(x, y)
}

func (ch *Stack) colorToRGBfloat64(c color.Color) (rgb [3]float64) {
    var rgbi [3]uint32
    rgbi[0], rgbi[1], rgbi[2], _ = c.RGBA()
    for i := range rgb {
        rgb[i] = colorByteToFloat64(byte(rgbi[i] / 257))
    }
    return
}

func (ch *Stack) StoreColor(x, y int, c color.Color, f func(float64) complex128) {
    rgb := ch.colorToRGBfloat64(c)
    if len(ch.sheets) == 1 {
        rgb[0] += rgb[1] + rgb[2]
        rgb[0] /= 3.0
    }
    for i := range ch.sheets {
        ch.Set(i, x, y, f(rgb[i]))
    }
}

func (ch *Stack) StoreComplex(i, x, y int, c color.Color, f func(float64, float64, float64) complex128) {
    rgb := ch.colorToRGBfloat64(c)
    ch.Set(i, x, y, f(rgb[0], rgb[1], rgb[2]))
}

func (ch *Stack) LoadColor(x, y int, f func(complex128) float64) color.RGBA {
    var rgb [3]byte

    for i := range ch.sheets {
        rgb[i] = colorFloat64ToByte(f(ch.Get(i, x, y)))
    }

    if len(ch.sheets) == 1 {
        rgb[1], rgb[2] = rgb[0], rgb[0]
    }

    return color.RGBA{rgb[0], rgb[1], rgb[2], 255}
}

func (ch *Stack) LoadComplex(i, x, y int, f func(complex128) (float64, float64, float64)) color.RGBA {
    r, g, b := f(ch.Get(i, x, y))
    return color.RGBA{colorFloat64ToByte(r), colorFloat64ToByte(g), colorFloat64ToByte(b), 255}
}




func colorByteToFloat64(x byte) float64 {
    return float64(x) / 256.0 + 1.0 / 512.0
}

func colorFloat64ToByte(x float64) byte {
    return clampByte(int(x * 256.0 - 0.5))
}

func clampByte(x int) byte {
    if x < 0 {
        return 0
    }
    if x > 255 {
        return 255
    }
    return byte(x)
}

