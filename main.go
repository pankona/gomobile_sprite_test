// +build darwin linux

package main

import (
	"image"
	"log"
	"time"

	_ "image/jpeg"
	_ "image/png"

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
	startTime  = time.Now()
	images     *glutil.Images
	eng        sprite.Engine
	scene      *sprite.Node
	fps        *debug.FPS
	ballDeltaX float32 = 10
	ballDeltaY float32 = 10
)

type KonaSprite struct {
	node   *sprite.Node
	width  float32
	height float32
	posX   float32
	posY   float32
	radian float32
}

var Gopher KonaSprite
var Ball KonaSprite

var (
	spriteSizeX float32 = 140
	spriteSizeY float32 = 90
	screenSizeX float32 = 1080 / 2
	screenSizeY float32 = 1920 / 2
	affine      *f32.Affine
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
					// transparency of png
					glctx.Enable(gl.BLEND)
					glctx.BlendEquation(gl.FUNC_ADD)
					glctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
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

				Gopher.Apply()

				Ball.MoveWithReflection()
				Ball.Apply()

				onPaint(glctx, sz)
				a.Publish()
				a.Send(paint.Event{}) // keep animating
			case touch.Event:
				Gopher.Move(e.X, e.Y)
				Gopher.Rotate(Gopher.radian + 5)
				//Gopher.Size(Gopher.width, Gopher.height)
			}
		}
	})
}

func (sprite *KonaSprite) Move(x float32, y float32) {
	sprite.posX = x
	sprite.posY = y
}

func (sprite *KonaSprite) Rotate(radian float32) {
	sprite.radian = radian
}

func (sprite *KonaSprite) Size(w float32, h float32) {
	sprite.width = w
	sprite.height = h
}

func (sprite *KonaSprite) MoveWithReflection() {
	sprite.posX += ballDeltaX
	sprite.posY += ballDeltaY
	if sprite.posX > screenSizeX || sprite.posX < 0 {
		ballDeltaX *= -1
	}

	if sprite.posY > screenSizeY || sprite.posY < 0 {
		ballDeltaY *= -1
	}
}

func (sprite *KonaSprite) Apply() {
	curPosX := sprite.posX
	curPosY := sprite.posY
	r := sprite.radian * 3.141592653 / 180
	affine = &f32.Affine{
		{sprite.width * f32.Cos(r), sprite.height * -f32.Sin(r),
			curPosX - (sprite.width/2)*f32.Cos(r) + (sprite.height/2)*f32.Sin(r)},
		{sprite.width * f32.Sin(r), sprite.height * f32.Cos(r),
			curPosY - (sprite.height/2)*f32.Cos(r) - (sprite.width/2)*f32.Sin(r)},
	}
	eng.SetTransform(sprite.node, *affine)
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

	// load Gopher
	Gopher.Move(screenSizeX/2, screenSizeY/2)
	Gopher.width = spriteSizeX
	Gopher.height = spriteSizeY
	Gopher.radian = 0
	tex_gopher := loadTextures("waza-gophers.jpeg", image.Rect(152, 10, 152+int(Gopher.width), 10+int(Gopher.height)))
	Gopher.node = newNode()
	eng.SetSubTex(Gopher.node, tex_gopher)
	Gopher.Apply()

	// load Ball
	Ball.Move(screenSizeX/3, screenSizeY/3)
	Ball.width = 48
	Ball.height = 48
	Ball.radian = 0
	tex_ball := loadTextures("ball.png", image.Rect(0, 0, int(Ball.width), int(Ball.height)))
	Ball.node = newNode()
	eng.SetSubTex(Ball.node, tex_ball)

	Ball.Apply()
}

func loadTextures(assetName string, rect image.Rectangle) sprite.SubTex {

	a, err := asset.Open(assetName)
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

	return sprite.SubTex{t, rect}
}

type arrangerFunc func(e sprite.Engine, n *sprite.Node, t clock.Time)

func (a arrangerFunc) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) { a(e, n, t) }
