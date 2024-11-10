//go:build js && wasm

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"log"
	"syscall/js"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	video  js.Value
	stream js.Value
	canvas js.Value
	ctx    js.Value
)

const (
	ScreenWidth  = 640
	ScreenHeight = 480
)

func init() {
	doc := js.Global().Get("document")
	video = doc.Call("createElement", "video")
	canvas = doc.Call("createElement", "canvas")
	video.Set("autoplay", true)
	video.Set("muted", true)
	video.Set("videoWidth", ScreenWidth)
	video.Set("videoHeight", ScreenHeight)
	mediaDevices := js.Global().Get("navigator").Get("mediaDevices")
	promise := mediaDevices.Call("getUserMedia", map[string]interface{}{
		"video": true,
		"audio": false,
	})
	promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		stream = args[0]
		video.Set("srcObject", stream)
		video.Call("play")
		canvas.Set("width", ScreenWidth)
		canvas.Set("height", ScreenHeight)
		ctx = canvas.Call("getContext", "2d")
		return nil
	}))
}

func fetchVideoFrame() []byte {
	ctx.Call("drawImage", video, 0, 0, ScreenWidth, ScreenHeight)
	data := ctx.Call("getImageData", 0, 0, ScreenWidth, ScreenHeight).Get("data")
	jsBin := js.Global().Get("Uint8Array").New(data)
	goBin := make([]byte, data.Get("length").Int())
	_ = js.CopyBytesToGo(goBin, jsBin)
	return goBin
}

type Game struct {
	err         error
	drawImg     *ebiten.Image
	cx, cy, rad float64
}

func loadImage(data []byte) (*ebiten.Image, error) {
	m, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	return ebiten.NewImageFromImage(m), nil
}

func newGame() *Game {
	return &Game{
		drawImg: ebiten.NewImage(ScreenWidth, ScreenHeight),
	}
}

func (g *Game) Update() error {
	if g.err != nil {
		return g.err
	}
	if !ctx.Truthy() {
		return nil
	}
	goBin := fetchVideoFrame()
	g.drawImg = ebiten.NewImageFromImage(newImage(goBin, ScreenWidth, ScreenHeight))
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.err != nil {
		return
	}
	screen.DrawImage(g.drawImg, nil)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %f", ebiten.CurrentFPS()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Face Play")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}

func newImage(data []byte, w, h int) *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < w*h; i++ {
		m.Pix[i*4+0] = uint8(data[i*4+0])
		m.Pix[i*4+1] = uint8(data[i*4+1])
		m.Pix[i*4+2] = uint8(data[i*4+2])
		m.Pix[i*4+3] = uint8(data[i*4+3])
	}
	return m
}

func rgbaToGrayscale(data []uint8, w, h int) []uint8 {
	gs := make([]uint8, w*h)
	for i := 0; i < w*h; i++ {
		r := float64(data[i*4+0])
		g := float64(data[i*4+1])
		b := float64(data[i*4+2])
		gs[i] = uint8(0.5 + 0.2126*r + 0.7152*g + 0.0722*b)
	}
	return gs
}
