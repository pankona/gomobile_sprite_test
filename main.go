// +build darwin linux

package main

import (
	"fmt"
	"image"
	"log"
	"time"

	_ "image/jpeg"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/asset"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/exp/sprite"
	"golang.org/x/mobile/exp/sprite/clock"
	"golang.org/x/mobile/exp/sprite/glsprite"
	"golang.org/x/mobile/gl"
)

var (
	startTime = time.Now()
	images    *glutil.Images
	eng       sprite.Engine
	scene     *sprite.Node
	fps       *debug.FPS
	node      *sprite.Node
)

type KonaSprite struct {
	node   *sprite.Node
	width  float32
	height float32
	x      float32
	y      float32
	radian float32
}

var (
	spriteSizeX float32 = 140
	spriteSizeY float32 = 90
	screenSizeX float32 = 800
	screenSizeY float32 = 800
	affine      *f32.Affine
	Gopher      *KonaSprite
)

func main() {
	app.Main(func(a app.App) {
		var glctx gl.Context
		var sz size.Event
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					glctx, _ = e.DrawContext.(gl.Context)
					onStart(glctx)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					onStop()
					glctx = nil
				}
			case size.Event:
				sz = e
			case paint.Event:
				if glctx == nil || e.External {
					continue
				}

				Gopher.rotate(Gopher.radian + 5)

				onPaint(glctx, sz)
				a.Publish()
				a.Send(paint.Event{}) // keep animating
			case touch.Event:
				if e.Type == touch.TypeEnd {
					Gopher.move(e.X, e.Y)
				}
			}
		}
	})
}

func (sprite *KonaSprite) move(x float32, y float32) {
	fmt.Println("move to ", x, y)
	sprite.x = x
	sprite.y = y
}

func (sprite *KonaSprite) rotate(radian float32) {
	// Rotation
	sprite.radian = radian
}

func (sprite *KonaSprite) apply() {
	curPosX := sprite.x
	curPosY := sprite.x
	r := sprite.radian * 3.141592653 / 180
	affine = &f32.Affine{
		{spriteSizeX * f32.Cos(r), spriteSizeY * -f32.Sin(r),
			curPosX - (spriteSizeX/2)*f32.Cos(r) + (spriteSizeY/2)*f32.Sin(r)},
		{spriteSizeX * f32.Sin(r), spriteSizeY * f32.Cos(r),
			curPosY - (spriteSizeY/2)*f32.Cos(r) - (spriteSizeX/2)*f32.Sin(r)},
	}
	eng.SetTransform(node, *affine)
}

func onStart(glctx gl.Context) {
	images = glutil.NewImages(glctx)
	fps = debug.NewFPS(images)
	eng = glsprite.Engine(images)
	loadScene()
}

func onStop() {
	eng.Release()
	fps.Release()
	images.Release()
}

func onPaint(glctx gl.Context, sz size.Event) {
	glctx.ClearColor(1, 1, 1, 1) // white background
	glctx.Clear(gl.COLOR_BUFFER_BIT)
	now := clock.Time(time.Since(startTime) * 60 / time.Second)
	eng.Render(scene, now, sz)
	fps.Draw(sz)
}

func newNode() *sprite.Node {
	n := &sprite.Node{}
	eng.Register(n)
	scene.AppendChild(n)
	return n
}

func loadScene() {
	// scene: base texture
	scene = &sprite.Node{}
	eng.Register(scene)
	eng.SetTransform(scene, f32.Affine{
		{1, 0, 0},
		{0, 1, 0},
	})

	// textures
	texs := loadTextures()
	node = newNode()
	eng.SetSubTex(node, texs[texGopherR])

	Gopher.x = screenSizeX / 2
	Gopher.y = screenSizeY / 2
	Gopher.width = spriteSizeX
	Gopher.height = spriteSizeY
	Gopher.radian = 0
	Gopher.apply()

	//affine = &f32.Affine{
	//	{Gopher.width, 0, Gopher.x},
	//	{0, Gopher.height, Gopher.y},
	//}
	//eng.SetTransform(node, *affine)
}

const (
	texGopherR = iota
)

func loadTextures() []sprite.SubTex {
	a, err := asset.Open("waza-gophers.jpeg")
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	img, _, err := image.Decode(a)
	if err != nil {
		log.Fatal(err)
	}
	t, err := eng.LoadTexture(img)
	if err != nil {
		log.Fatal(err)
	}

	return []sprite.SubTex{
		texGopherR: sprite.SubTex{t, image.Rect(152, 10, 152+int(spriteSizeX), 10+int(spriteSizeY))},
	}
}

type arrangerFunc func(e sprite.Engine, n *sprite.Node, t clock.Time)

func (a arrangerFunc) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) { a(e, n, t) }
