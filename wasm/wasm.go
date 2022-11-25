package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"syscall/js"
	"time"
)

type Place struct {
	x int
	y int
}

type Snake struct {
	places []Place
	move   Place
}

const (
	sqauresInEachLine = 16
)

var (
	// js.Value can be any JS object/type/constructor
	window, doc, body, canvas, canvasCtx, scoreDiv js.Value
	windowSize                                     int
	spacing, score                                 int
	snake                                          Snake
	apple                                          Place
	keys                                           = map[string]Place{
		"ArrowUp": {
			x: 0,
			y: -1,
		},
		"ArrowDown": {
			x: 0,
			y: 1,
		},
		"ArrowLeft": {
			x: -1,
			y: 0,
		},
		"ArrowRight": {
			x: 1,
			y: 0,
		},
	}
)

func (s *Snake) draw() {
	for _, p := range s.places {
		drawSquare(p.x, p.y, "blue")
	}
}

func (s *Snake) CheckIfHasSquare(x, y int) error {
	for _, p := range s.places {
		if x == p.x && y == p.y {
			return errors.New("has sqaure")
		}
	}
	return nil
}

func (s *Snake) Move(p *Place) error {

	x := s.places[0].x + p.x
	y := s.places[0].y + p.y

	if x == -1 || x == sqauresInEachLine || y == -1 || y == sqauresInEachLine {
		return errors.New("outbound")
	}

	if err := s.CheckIfHasSquare(x, y); err != nil {
		return err
	}

	newPlaces := []Place{
		{
			x: x,
			y: y,
		},
	}

	l := len(s.places)

	for i := 1; i < l; i++ {
		newPlaces = append(newPlaces, s.places[i-1])
	}

	s.move = *p
	s.places = newPlaces

	if x == apple.x && y == apple.y {
		createApple()
		newPlaces = append(newPlaces, s.places[l-1])
		score++
		fmt.Println(score)
		s.places = newPlaces
	}

	return nil
}

func main() {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	updateCh := make(chan bool)

	setup(js.Value{}, []js.Value{})
	createApple()
	refresh()
	snake.Move(&Place{
		x: 1,
		y: 0,
	})
	refresh()
	go move(updateCh, cancel)

	js.Global().Set("setup", js.FuncOf(setup))
	js.Global().Set("listenToKeys", listenToKeys(updateCh, cancel))

	<-ctx.Done()
	close(updateCh)
	body.Set("innerHTML", fmt.Sprintf("<h1>Game over</h1><h2>score: %d</h2>", score))
}

func refresh() {

	canvasCtx.Call("clearRect", 0, 0, windowSize, windowSize)

	for s := spacing; s < int(windowSize); s += spacing {

		canvasCtx.Call("beginPath")
		canvasCtx.Call("moveTo", s, 0)
		canvasCtx.Call("lineTo", s, windowSize)
		canvasCtx.Call("stroke")

		canvasCtx.Call("beginPath")
		canvasCtx.Call("moveTo", 0, s)
		canvasCtx.Call("lineTo", windowSize, s)
		canvasCtx.Call("stroke")
	}

	drawSquare(apple.x, apple.y, "red")
	snake.draw()
	scoreDiv.Set("innerHTML", fmt.Sprintf("<h1>%d</h1>", score))
}

func move(ch <-chan bool, cancel context.CancelFunc) {
	for {
		select {
		case <-ch:
			refresh()
		case <-time.After(1 * time.Second):
			if err := snake.Move(&snake.move); err != nil {
				cancel()
			}
			refresh()
		}
	}
}

func drawSquare[A int](x, y A, color string) {
	canvasCtx.Call("beginPath")
	canvasCtx.Call("rect", x*A(spacing), y*A(spacing), spacing, spacing)
	canvasCtx.Set("fillStyle", color)
	canvasCtx.Call("fill")
}

func listenToKeys(ch chan<- bool, cancel context.CancelFunc) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {

		if args[0].Type() != js.TypeString {
			panic("Listen to keys wasn't passed a string")
		}

		p := keys[args[0].String()]

		if (p.x != 0 || p.y != 0) && p.x != snake.places[0].x-snake.places[1].x {
			if err := snake.Move(&p); err != nil {
				cancel()
				return nil
			}
			ch <- true
		}

		return nil
	})
}

func setup(this js.Value, args []js.Value) any {

	rand.Seed(time.Now().UnixNano())

	window = js.Global()
	doc = window.Get("document")
	body = doc.Get("body")
	score = 0
	scoreDiv = doc.Call("getElementById", "score")

	var (
		windowSizeH, windowSizeW = 0.0, 0.0
	)

	if heighV := window.Get("innerHeight"); heighV.Type() == js.TypeNumber {
		windowSizeH = heighV.Float() - 20
	} else {
		panic("window height isn't a number")
	}

	if widthV := window.Get("innerWidth"); widthV.Type() == js.TypeNumber {
		windowSizeW = widthV.Float() - 20
	} else {
		panic("window width isn't a number")
	}

	minSize := math.Min(windowSizeH, windowSizeW)
	minSize = (minSize)
	for math.Mod(minSize, sqauresInEachLine) != 0 {
		minSize--
	}
	windowSize = int(minSize)

	canvas = doc.Call("createElement", "canvas")
	canvas.Set("height", windowSize)
	canvas.Set("width", windowSize)
	body.Call("appendChild", canvas)

	canvasCtx = canvas.Call("getContext", "2d")

	canvas.Set("style", "border-style: solid;border-color: black;border-width: 1px;")

	spacing = int(minSize / sqauresInEachLine)

	snake = Snake{
		places: []Place{
			{
				x: 3,
				y: sqauresInEachLine / 2,
			},
			{
				x: 2,
				y: sqauresInEachLine / 2,
			},
			{
				x: 1,
				y: sqauresInEachLine / 2,
			},
		},
	}

	apple = Place{
		x: 0,
		y: 0,
	}

	snake.draw()

	return nil
}

func createApple() {
	x := rand.Intn(sqauresInEachLine)
	y := rand.Intn(sqauresInEachLine)

	err := snake.CheckIfHasSquare(x, y)
	for err != nil {
		x = rand.Intn(sqauresInEachLine)
		y = rand.Intn(sqauresInEachLine)
		err = snake.CheckIfHasSquare(x, y)
	}

	apple.x = x
	apple.y = y

}
